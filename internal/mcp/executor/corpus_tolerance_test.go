package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/testutil"
)

// DIR-099: the cross-tool malformed-file tolerance regression gate.
//
// DIR-018 fixed skip-and-report tolerance in the analysis loadData path, and
// DIR-032/DIR-034 rewrote the ListSessions path behind query_sessions without
// it — a single empty session file then hard-crashed query_sessions with
// "no Claude entries in <file>". DIR-094 restored the behavior on every path
// that makes a whole-file inclusion decision. Nothing mechanical stopped the
// next rewrite from dropping it again, which is what this file is: a gate that
// runs EVERY registered corpus-consuming tool against the checked-in corrupt
// corpus, plus a coverage check that refuses to let a newly registered tool
// go unclassified.
//
// Two ideas carry the gate:
//
//   - registerCorpusTool answers "what must hold for this tool", in one named
//     place per tool. A tool missing from both tables fails
//     TestCorpusTolerance_EveryRegisteredToolIsClassified, so a new tool
//     cannot silently lack tolerance — the gap is a failing test, not a code
//     review someone has to remember to do.
//
//   - every entry is run twice against byte-identical corpora: the corrupted
//     one and a control with only healthy.jsonl. The tool's response, with
//     `warnings` and the volatile file_ref path removed, must be identical.
//     That is the tolerance claim itself — a malformed sibling changed nothing
//     about what the tool returned — and it holds for tools that report
//     exclusions and for tools that never make an exclusion decision, so one
//     assertion covers the whole surface instead of a per-tool paraphrase.
//
// The registration procedure for a new tool is documented in
// docs/reference/corpus-tolerance-gate.md.

// corpusToleranceContract is what "tolerates malformed session files" means
// for one tool. The two kinds are genuinely different behaviors, not two
// strictnesses of the same one, so each entry has to say which it is.
type corpusToleranceContract int

const (
	// exclusionContract is for a tool that decides, per enumerated session
	// file, whether that whole file contributes anything: ListSessions behind
	// query_sessions, and the analysis loader behind get_timeline and the six
	// analysis tools. Tolerance here has two halves — the healthy session's
	// data still reaches the caller AND every excluded file is named in the
	// response warnings (DIR-018's "skip and report", which DIR-094 extended to
	// the zero-message case).
	exclusionContract corpusToleranceContract = iota

	// streamingContract is for a tool that streams records line by line out of
	// every enumerated file and never makes a whole-file inclusion decision
	// (the four consolidated query tools' shared
	// dispatchProviderQuery → StreamFilesWithTimeRange path). There is no
	// excluded file to name: a truncated line is skipped and the rest of the
	// file still contributes, which is why DIR-094's per-path list does not
	// include these tools. Tolerance here means one thing only — a malformed
	// sibling neither aborts the call nor changes the records returned.
	streamingContract
)

// corpusToolCase is one registration entry: how to call a corpus-consuming
// tool, and what its tolerance claim is.
type corpusToolCase struct {
	// name is the registered tool name. It must be present in
	// specialToolRegistry or queryHandlerRegistry; the coverage check enforces
	// that, and that this table has no stale entries.
	name     string
	contract corpusToleranceContract
	// args are the tool arguments that select the corpus-reading data path.
	// working_dir and provider are supplied by the harness.
	args map[string]interface{}
	// healthyMarker is a substring only healthy.jsonl can contribute. It is
	// what makes the control run's data attributable to the healthy session
	// rather than to some zero-valued field that happens to serialize the same
	// way. Required unless emptyCorpusReason is set.
	healthyMarker string
	// emptyCorpusReason explains why the checked-in corpus cannot exercise
	// this tool's data path at all, for the tools where it cannot. An entry
	// may leave healthyMarker empty only by writing one of these, so "no
	// marker" is a reviewed statement rather than an omission.
	emptyCorpusReason string
}

// corpusToolCases is the single registration point. Order is the tool's
// execution order, for stable subtest names.
var corpusToolCases = []corpusToolCase{
	{
		name:     "query_sessions",
		contract: exclusionContract,
		args:     map[string]interface{}{"output_mode": "inline"},
		// The one session id the healthy fixture declares.
		healthyMarker: "aaaaaaaa-1111-4111-8111-aaaaaaaaaaaa",
	},
	{
		name:     "analyze_errors",
		contract: exclusionContract,
		// One is_error tool_result in the healthy session.
		healthyMarker: `"total_errors":1`,
	},
	{
		name:     "analyze_bugs",
		contract: exclusionContract,
		// The failed Read followed by a successful Read → one fix pair.
		healthyMarker: `"total_pairs":1`,
	},
	{
		name:     "quality_scan",
		contract: exclusionContract,
		// 3 tool calls, 1 of them errored.
		healthyMarker: `"name":"error_rate","raw_value":"1/3"`,
	},
	{
		name:     "get_work_patterns",
		contract: exclusionContract,
		// Grep + 2×Read is the healthy session's whole tool surface.
		healthyMarker: `"tool_name":"Read"`,
	},
	{
		name:     "get_tech_debt",
		contract: exclusionContract,
		// The successful Read's output carries a "// TODO:" comment.
		healthyMarker: `"label":"TODO"`,
	},
	{
		name:     "get_timeline",
		contract: exclusionContract,
		// 8 entries spanning 00:00:00Z → 00:00:31Z.
		healthyMarker: `"total_span":"31s"`,
	},
	{
		name:     "query_session_content",
		contract: streamingContract,
		// role=user requires string message content, which the fixture's
		// array-shaped user turns are not; role=all is the path that returns
		// the healthy session.
		args:          map[string]interface{}{"role": "all", "output_mode": "inline"},
		healthyMarker: "msg_healthy_1",
	},
	{
		name:     "query_session_signals",
		contract: streamingContract,
		args:     map[string]interface{}{"type": "tokens", "output_mode": "inline"},
		// One token-usage-bearing assistant turn, by its message id.
		healthyMarker: "msg_healthy_1",
	},
	{
		name:     "query_file_activity",
		contract: streamingContract,
		args:     map[string]interface{}{"type": "snapshots", "output_mode": "inline"},
		emptyCorpusReason: "type=snapshots selects file-history-snapshot records carrying a messageId, " +
			"and healthy.jsonl contains none, so this corpus returns zero records with or without the " +
			"corrupt siblings. Only the no-error and control-equality halves of the contract are " +
			"exercisable for it here; its corpus read path is the shared dispatchProviderQuery → " +
			"StreamFilesWithTimeRange path that query_session_signals exercises with real records.",
	},
}

// nonCorpusTools are the registered tools that do not read session corpus
// content, with the reason. The reason is read by a human and is the whole
// point of the entry: it records that the exclusion was decided, not missed.
var nonCorpusTools = map[string]string{
	"cleanup_temp_files": "operates on meta-cc's own temp files, not on session history",
	"get_session_directory": "reports directory-level metadata (file count, total size, oldest/newest) " +
		"without parsing a single record, so there is no per-file inclusion decision to report",
	"get_session_metadata": "enumerates the corpus but returns per-file metadata (record count, size) " +
		"for every file; nothing is ever excluded, so there is no exclusion to name",
	"inspect_session_files": "takes an explicit `files` array; it never enumerates a corpus, so it has " +
		"no sibling files to tolerate",
	"execute_stage2_query": "takes an explicit `files` array; same as inspect_session_files",
	"query_edit_sequences": "requires an explicit `files` array (verified: a call without it fails " +
		"validation), so it never enumerates a corpus",
}

// TestCorpusTolerance_CorruptSiblingsChangeNothing is the gate itself: every
// registered corpus-consuming tool, run against the corrupt corpus and against
// the same corpus with only the healthy session in it.
func TestCorpusTolerance_CorruptSiblingsChangeNothing(t *testing.T) {
	for _, tc := range corpusToolCases {
		t.Run(tc.name, func(t *testing.T) {
			// One project, run twice: once with the malformed siblings
			// present, once after removing them. A second seeded project
			// would differ in its working_dir, and every session-derived
			// record echoes that path — the comparison would then also be
			// testing that two temp directories are equal.
			project := testutil.SeedMalformedCorpusProject(t)
			gotCorrupt, warningsCorrupt := runCorpusTool(t, tc, project.ProjectPath)

			removeExcludedSiblings(t, project)
			gotControl, warningsControl := runCorpusTool(t, tc, project.ProjectPath)

			assert.Equal(t, gotControl, gotCorrupt,
				"a malformed sibling must not change what %s returns: the response (minus warnings and "+
					"the volatile file_ref path) has to be identical to the clean-corpus run", tc.name)

			if tc.healthyMarker != "" {
				assert.Contains(t, gotCorrupt, tc.healthyMarker,
					"the healthy session's own data must still reach %s despite the excluded siblings", tc.name)
			} else {
				assert.NotEmpty(t, tc.emptyCorpusReason,
					"a %s entry with no healthyMarker must record why the corpus cannot exercise it", tc.name)
			}

			if tc.contract == exclusionContract {
				for _, name := range project.Excluded {
					assert.True(t, testutil.WarningsNameFile(warningsCorrupt, name),
						"%s excludes whole files, so it must name the excluded %s; got warnings %v",
						tc.name, name, warningsCorrupt)
				}
				// Non-vacuity: "names every excluded file" is also satisfied by
				// a tool that names every file. A clean corpus must produce no
				// exclusions at all, so the two halves together pin the
				// warnings to the files that were actually dropped.
				assert.Empty(t, warningsControl,
					"%s must report no excluded files for a corpus that has none; got %v",
					tc.name, warningsControl)
			}
		})
	}
}

// TestCorpusTolerance_EveryRegisteredToolIsClassified is the coverage half:
// a tool present in either registry must appear in exactly one of the two
// tables above. Adding a tool and not registering it — the way the ListSessions
// rewrite silently dropped DIR-018's tolerance — fails here instead of shipping.
func TestCorpusTolerance_EveryRegisteredToolIsClassified(t *testing.T) {
	problems := classifyCorpusTools(registeredToolNames())
	assert.Empty(t, problems, "every registered tool must be registered in corpusToolCases or "+
		"nonCorpusTools (see docs/reference/corpus-tolerance-gate.md):\n  - %s",
		strings.Join(problems, "\n  - "))
}

// TestCorpusTolerance_ClassifierRejectsAnUnclassifiedTool is the non-vacuity
// guard for the coverage check. Without it, a classifier that silently matched
// nothing would let the coverage test pass forever while enforcing nothing —
// which is the failure mode it exists to prevent.
//
// The synthetic registry is the exact shape a newly added tool produces: a
// registered name that neither table mentions.
func TestCorpusTolerance_ClassifierRejectsAnUnclassifiedTool(t *testing.T) {
	registered := registeredToolNames()
	registered["scratch_tool_that_lacks_tolerance"] = true

	problems := classifyCorpusTools(registered)
	require.NotEmpty(t, problems, "a registered but unclassified tool must be reported")
	require.Contains(t, strings.Join(problems, "\n"), "scratch_tool_that_lacks_tolerance",
		"the report must name the offending tool, not just fail")
}

// registeredToolNames is every tool the executor can dispatch: the special
// handlers and the convenience query handlers. It is read from the live
// registries rather than a hand-kept list, so the coverage check cannot drift
// from what the server actually serves.
func registeredToolNames() map[string]bool {
	names := make(map[string]bool, len(specialToolRegistry)+len(queryHandlerRegistry))
	for name := range specialToolRegistry {
		names[name] = true
	}
	for name := range queryHandlerRegistry {
		names[name] = true
	}
	return names
}

// classifyCorpusTools returns one problem string per way the two registration
// tables and the live registry disagree: a registered tool nobody classified,
// a table entry for a tool that is not registered, an entry that declares
// neither or both of healthyMarker and emptyCorpusReason, and a name listed in
// both tables. Sorted, so the report for a given set of inputs is stable.
func classifyCorpusTools(registered map[string]bool) []string {
	var problems []string

	corpusCases := make(map[string]corpusToolCase, len(corpusToolCases))
	for _, tc := range corpusToolCases {
		if _, dup := corpusCases[tc.name]; dup {
			problems = append(problems, "corpusToolCases lists "+tc.name+" more than once")
		}
		corpusCases[tc.name] = tc

		switch {
		case tc.healthyMarker == "" && tc.emptyCorpusReason == "":
			problems = append(problems, tc.name+" declares neither healthyMarker nor emptyCorpusReason")
		case tc.healthyMarker != "" && tc.emptyCorpusReason != "":
			problems = append(problems, tc.name+" declares both healthyMarker and emptyCorpusReason; "+
				"an exercised entry has no need for an exemption")
		}
	}

	for name := range registered {
		_, isCorpus := corpusCases[name]
		_, isNonCorpus := nonCorpusTools[name]
		switch {
		case !isCorpus && !isNonCorpus:
			problems = append(problems, name+" is registered but appears in neither corpusToolCases "+
				"nor nonCorpusTools")
		case isCorpus && isNonCorpus:
			problems = append(problems, name+" appears in both corpusToolCases and nonCorpusTools")
		}
	}

	for name := range corpusCases {
		if !registered[name] {
			problems = append(problems, name+" is in corpusToolCases but is not a registered tool")
		}
	}
	for name := range nonCorpusTools {
		if !registered[name] {
			problems = append(problems, name+" is in nonCorpusTools but is not a registered tool")
		}
	}

	sort.Strings(problems)
	return problems
}

// runCorpusTool dispatches one table entry the same way the server does, from
// the working_dir to the registered handler, and returns the canonicalized
// response plus its warnings. Using ExecuteTool rather than calling the
// handler directly is deliberate: it is the path a real caller takes, so a
// future rewrite that re-wires dispatch (the DIR-032/DIR-034 shape) is
// exercised here instead of bypassed.
func runCorpusTool(t *testing.T, tc corpusToolCase, projectPath string) (canonical string, warnings []string) {
	t.Helper()

	args := make(map[string]interface{}, len(tc.args)+2)
	for k, v := range tc.args {
		args[k] = v
	}
	args["working_dir"] = projectPath
	// Pinned rather than left to host detection (config.OmittedProviderDefault):
	// the corpus is a Claude one, so the gate must not change meaning
	// depending on which agent host runs it.
	args["provider"] = "claude"

	output, err := NewToolExecutor().ExecuteTool(&config.Config{}, tc.name, args)
	require.NoError(t, err, "%s must tolerate a malformed sibling, not fail the whole call", tc.name)

	return canonicalToolResponse(t, output), responseWarnings(t, output)
}

// canonicalToolResponse removes the two parts of a response that are supposed
// to differ between the corrupted and control runs, so the remainder can be
// compared byte for byte:
//
//   - `warnings` is the exclusion report itself, compared separately by the
//     caller (and expected to be empty on the clean corpus).
//   - file_ref's `path` is a fresh temp file per call, so two runs over
//     identical data still name different paths. The rest of the file_ref
//     envelope — record count, field list, recipes — is derived from the data
//     and is exactly what must not drift.
//
// json.Marshal sorts object keys, so the result is a stable byte string.
func canonicalToolResponse(t *testing.T, output string) string {
	t.Helper()

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &decoded), "tool output must be valid JSON: %s", output)
	delete(decoded, "warnings")
	if ref, ok := decoded["file_ref"].(map[string]interface{}); ok {
		delete(ref, "path")
	}

	canonical, err := json.Marshal(decoded)
	require.NoError(t, err)
	return string(canonical)
}

// responseWarnings returns the response's warnings array, empty when the key
// is absent (an analysis tool with nothing to report omits it rather than
// emitting an empty array).
func responseWarnings(t *testing.T, output string) []string {
	t.Helper()

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &decoded), "tool output must be valid JSON: %s", output)

	raw, ok := decoded["warnings"].([]interface{})
	if !ok {
		return nil
	}
	warnings := make([]string, 0, len(raw))
	for _, entry := range raw {
		if s, ok := entry.(string); ok {
			warnings = append(warnings, s)
		}
	}
	return warnings
}

// removeExcludedSiblings turns a seeded project into the control corpus: the
// same directory, same healthy session, and none of the malformed files. It is
// applied in place so the control run and the corrupted run share a
// working_dir, which keeps every session-derived path out of the diff.
func removeExcludedSiblings(t *testing.T, project testutil.MalformedCorpusProject) {
	t.Helper()

	for _, name := range project.Excluded {
		if err := os.Remove(filepath.Join(project.SessionDir, name)); err != nil {
			t.Fatalf("cannot remove %s from the control corpus: %v", name, err)
		}
	}
}
