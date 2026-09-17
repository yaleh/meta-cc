package executor

// DIR-099: the cross-tool malformed-file tolerance gate.
//
// DIR-018 established that one unusable session file must not erase the results
// derived from the rest of the corpus, and must not be dropped silently.
// DIR-094 re-unified that contract across the enumerating paths and pinned it
// for the analysis tools and for query_sessions. What was still missing is any
// MECHANICAL statement that EVERY session-corpus-consuming tool holds it: a
// future ListSessions-style rewrite (the DIR-032/DIR-034 change that lost the
// behavior in the first place) could drop tolerance in a tool nobody wrote a
// test for, and nothing would go red.
//
// This file is that statement. It is deliberately one table, in one place:
//
//   - `corpusToolGates()` registers every corpus-consuming tool exactly once,
//     by name. Adding tolerance behavior to a tool means adding an entry here.
//   - `TestCorpusToleranceGate_EveryToolIsClassified` fails when a tool
//     definition exists that is neither registered nor explicitly declared
//     non-corpus-consuming, so a NEW tool cannot silently skip registration.
//   - `TestCorpusToleranceGate` drives each registered tool against three
//     corpora (nothing usable / control only / control + every corruption
//     shape) and applies the shared verdict helpers below.
//   - `TestCorpusToleranceGate_RejectsAnIntolerantTool` is the gate's own
//     negative control: it feeds the verdict helpers the response of a tool
//     that DID drop the corpus, and fails if the gate would have accepted it.
//     Without this, a gate that silently stopped checking anything would look
//     identical to a gate that passes.
//
// The corpus itself lives in internal/testutil (DIR-099 fixture), so the same
// three corruption shapes are shared rather than re-spelled per test.
//
// Registration procedure for a new corpus-consuming tool:
// docs/guides/corpus-tolerance-gate.md.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/locator"
	toolspkg "github.com/yaleh/meta-cc/internal/mcp/tools"
	"github.com/yaleh/meta-cc/internal/testutil"
)

// toleranceKind classifies how far a tool's malformed-file tolerance reaches.
// It selects the assertion set the gate applies to that tool; the registry
// entry's `note` records the reasoning for every tool that is not fully
// conforming.
type toleranceKind int

const (
	// kindEnumerationReported: the tool enumerates the corpus, skips unusable
	// files, and NAMES every one it skipped in both the prose (`warnings`) and
	// machine-readable (`skipped_files`) channels — the full DIR-018/DIR-094
	// contract.
	kindEnumerationReported toleranceKind = iota

	// kindEnumerationSilent: the tool enumerates the corpus and still returns
	// the usable session's results, but does not surface which files it
	// skipped. The tolerance half of the contract is asserted; the reporting
	// half is not implemented for these tools today and is recorded as such
	// rather than pretended.
	kindEnumerationSilent

	// kindSchemaSpecific: the tool's response shape is not a corpus result set
	// (it is an inventory, or it was pointed at explicit file paths), so there
	// is no clean-vs-dirty payload equality to assert. The entry supplies its
	// own `assertSurvives`, which must show the control session still reached
	// the response and the unusable files did not fail the call.
	kindSchemaSpecific
)

// corpusToolGate is one registry entry: one tool, one place.
type corpusToolGate struct {
	// name is the MCP tool name as it appears in the tool definitions.
	name string
	// kind selects the assertion set.
	kind toleranceKind
	// argsFor builds the tool arguments for a given corpus. Entries that
	// enumerate pass working_dir; entries pointed at explicit paths pass the
	// corpus's file list.
	argsFor func(c gateCorpus) map[string]interface{}
	// mustName lists the corrupt shapes whose session IDs the response MUST
	// contain somewhere. It is the per-tool record of how far the reporting
	// half of the contract reaches; empty means the tool's contract does not
	// require naming (see note).
	mustName []testutil.CorruptShape
	// assertSurvives is required for kindSchemaSpecific and must check, against
	// this tool's own response schema, that the control session still reached
	// the response under a dirty corpus.
	assertSurvives func(t *testing.T, clean, dirty gateCorpus, cleanOut, dirtyOut string)
	// note explains any departure from the full contract.
	note string
}

// corpusToolGates is the single registration point. One entry per
// session-corpus-consuming tool; see the file header for what "consuming"
// means and TestCorpusToleranceGate_EveryToolIsClassified for the check that
// keeps this list complete.
func corpusToolGates() []corpusToolGate {
	return []corpusToolGate{
		// ── corpus-enumerating; the skip is reported (DIR-018/DIR-094) ───────
		{
			name:     "query_sessions",
			kind:     kindEnumerationReported,
			mustName: allCorruptShapes(),
			// inline_threshold_bytes is raised so the response body stays
			// inline: the gate compares response bodies, and a file_ref
			// envelope carries a random temp-file path that differs between
			// corpora for reasons unrelated to tolerance.
			argsFor: enumeratingArgs(map[string]interface{}{"inline_threshold_bytes": 1 << 20}),
		},
		{
			name:     "get_timeline",
			kind:     kindEnumerationReported,
			mustName: allCorruptShapes(),
			argsFor:  enumeratingArgs(nil),
		},
		{
			name:     "analyze_errors",
			kind:     kindEnumerationReported,
			mustName: allCorruptShapes(),
			argsFor:  enumeratingArgs(nil),
		},
		{
			name:     "analyze_bugs",
			kind:     kindEnumerationReported,
			mustName: allCorruptShapes(),
			argsFor:  enumeratingArgs(nil),
		},
		{
			name:     "quality_scan",
			kind:     kindEnumerationReported,
			mustName: allCorruptShapes(),
			argsFor:  enumeratingArgs(nil),
		},
		{
			name:     "get_work_patterns",
			kind:     kindEnumerationReported,
			mustName: allCorruptShapes(),
			argsFor:  enumeratingArgs(nil),
		},
		{
			name:     "get_tech_debt",
			kind:     kindEnumerationReported,
			mustName: allCorruptShapes(),
			argsFor:  enumeratingArgs(nil),
		},

		// ── corpus-enumerating; the skip is NOT reported ─────────────────────
		{
			name: "query_session_content",
			kind: kindEnumerationSilent,
			// role=all rather than role=user: the control session's user
			// records carry ARRAY content, and the role=user path defaults its
			// pattern filter to the STRING content type, which matches nothing
			// on an array. role=all exercises the widest read of the corpus.
			argsFor: enumeratingArgs(map[string]interface{}{
				"role": "all", "inline_threshold_bytes": 1 << 20,
			}),
			note: "returns the usable session's records and never fails the batch, but a malformed sibling file " +
				"produces no entry in warnings/skipped_files",
		},
		{
			name: "query_session_signals",
			kind: kindEnumerationSilent,
			argsFor: enumeratingArgs(map[string]interface{}{
				"type": "errors", "inline_threshold_bytes": 1 << 20,
			}),
			note: "see query_session_content; same pipeline, same silent skip",
		},
		{
			name: "query_file_activity",
			kind: kindEnumerationSilent,
			argsFor: enumeratingArgs(map[string]interface{}{
				"type": "snapshots", "inline_threshold_bytes": 1 << 20,
			}),
			note: "see query_session_content; same pipeline, same silent skip",
		},

		// ── response shape is an inventory, not a result set ─────────────────
		{
			name:    "get_session_directory",
			kind:    kindSchemaSpecific,
			argsFor: enumeratingArgs(map[string]interface{}{"scope": "project"}),
			// The response is directory-level aggregates, so there is no
			// per-file path to name; "nothing was dropped" is asserted as
			// file_count accounting for every file in the corpus.
			assertSurvives: assertDirectoryAccountsForEveryFile,
		},
		{
			name:           "get_session_metadata",
			kind:           kindSchemaSpecific,
			argsFor:        enumeratingArgs(map[string]interface{}{"scope": "project"}),
			mustName:       allCorruptShapes(),
			assertSurvives: assertMetadataListsEveryFile,
			note:           "lists a per-file entry (path, records, size_bytes) for every file in the corpus, including the unusable ones, so an unusable file is visible rather than skipped",
		},
		{
			name:           "inspect_session_files",
			kind:           kindSchemaSpecific,
			argsFor:        fileArgs(),
			mustName:       allCorruptShapes(),
			assertSurvives: assertInspectReportsControlIntact,
			note:           "pointed at explicit paths; echoes a per-file entry for every path it was given",
		},
		{
			name:    "execute_stage2_query",
			kind:    kindSchemaSpecific,
			argsFor: fileArgsWith(map[string]interface{}{"filter": `select(.type == "user")`}),
			// Names the files it could not parse, in diagnostics.skip_warnings.
			// It does NOT name a zero-message file: that one loads without
			// error and contributes zero records (see note).
			mustName:       []testutil.CorruptShape{testutil.ShapeTruncated, testutil.ShapeWrongShape},
			assertSurvives: assertStage2KeepsControlRecords,
			note: "reports unparseable files as files_skipped + diagnostics.skip_warnings, but a zero-message " +
				"file loads successfully with zero records and is not named — the 2026-07-30 shape is still silent here",
		},
		{
			name:           "query_edit_sequences",
			kind:           kindSchemaSpecific,
			argsFor:        fileArgs(),
			assertSurvives: assertEditSequencesUnchanged,
			note: "this corpus yields no edit sequences even when clean, so the assertion pins only that the " +
				"unusable files neither fail the call nor perturb the result",
		},
	}
}

// nonCorpusTools are defined tools that legitimately never read the session
// corpus. Everything else must be registered in corpusToolGates(). Each entry
// must say why, so "not corpus-consuming" is a checked claim rather than an
// omission.
var nonCorpusTools = map[string]string{
	"cleanup_temp_files": "removes old MCP temp files; never opens a session transcript",
}

// allCorruptShapes returns every shape in the shared fixture, so an entry that
// names "everything" keeps naming everything when the fixture grows.
func allCorruptShapes() []testutil.CorruptShape {
	shapes := make([]testutil.CorruptShape, 0, len(testutil.CorruptFiles()))
	for _, file := range testutil.CorruptFiles() {
		shapes = append(shapes, file.Shape)
	}
	return shapes
}

// enumeratingArgs builds the argument builder for a tool that discovers its
// corpus from working_dir, applying any tool-specific extras.
func enumeratingArgs(extra map[string]interface{}) func(gateCorpus) map[string]interface{} {
	return func(c gateCorpus) map[string]interface{} {
		args := map[string]interface{}{"working_dir": c.project}
		for key, value := range extra {
			args[key] = value
		}
		return args
	}
}

// fileArgs builds the argument builder for a tool pointed at explicit paths.
func fileArgs() func(gateCorpus) map[string]interface{} {
	return fileArgsWith(nil)
}

func fileArgsWith(extra map[string]interface{}) func(gateCorpus) map[string]interface{} {
	return func(c gateCorpus) map[string]interface{} {
		args := map[string]interface{}{"files": c.sessionPaths()}
		for key, value := range extra {
			args[key] = value
		}
		return args
	}
}

// gateCorpus is one installed corpus: a resolved project directory, the
// transcript directory its sessions live in, and which files are present.
type gateCorpus struct {
	project    string
	sessionDir string
	corrupt    []testutil.CorruptFile
	// withControl records whether the control session was installed.
	withControl bool
}

func (c gateCorpus) path(sessionID string) string {
	return filepath.Join(c.sessionDir, sessionID+".jsonl")
}

// sessionPaths returns every session file in the corpus, control first, in the
// shape the tools' `files` parameter expects.
func (c gateCorpus) sessionPaths() []interface{} {
	paths := make([]interface{}, 0, len(c.corrupt)+1)
	if c.withControl {
		paths = append(paths, c.path(testutil.ControlSessionName))
	}
	for _, file := range c.corrupt {
		paths = append(paths, c.path(file.Name))
	}
	return paths
}

// corruptNames returns the session IDs of the corrupt files in this corpus.
func (c gateCorpus) corruptNames() []string {
	names := make([]string, 0, len(c.corrupt))
	for _, file := range c.corrupt {
		names = append(names, file.Name)
	}
	return names
}

// installCorpus replaces the contents of one transcript directory with the
// requested corpus.
//
// Every corpus a gate entry is checked against is installed at the SAME project
// path and projects root, reusing the same transcript directory: the three
// responses are compared field by field, and the project path is echoed back
// inside them (a session's `cwd`, a transcript path), so corpora installed at
// different paths would differ for reasons that have nothing to do with
// tolerance. Callers therefore install a corpus, take the response, and install
// the next one over it.
func installCorpus(t *testing.T, projectsRoot, project string, withControl bool, corrupt []testutil.CorruptFile) gateCorpus {
	t.Helper()
	sessionDir := filepath.Join(projectsRoot, locator.PathToHash(project))

	require.NoError(t, os.RemoveAll(sessionDir), "clear the previous corpus")
	testutil.InstallCorpus(t, testutil.CorpusOptions{
		SessionDir:  sessionDir,
		ProjectPath: project,
		WithControl: withControl,
		Corrupt:     corrupt,
	})

	return gateCorpus{
		project:     project,
		sessionDir:  sessionDir,
		corrupt:     corrupt,
		withControl: withControl,
	}
}

// requireFixtureIsVisible fails once, clearly, when the shared fixture is not
// wired up correctly — a control session whose cwd or transcript directory the
// listing paths cannot see. Without this, a broken fixture would surface as a
// dozen confusing per-tool failures (or, worse, as vacuously passing
// assertions), because every tool would report the same empty corpus.
func requireFixtureIsVisible(t *testing.T) {
	t.Helper()
	corpus := newCorpusContext(t)
	corpus.install(t, true, nil)

	output, err := NewToolExecutor().ExecuteTool(&config.Config{}, "query_sessions", map[string]interface{}{
		"working_dir": corpus.project,
	})
	require.NoError(t, err, "the fixture's control corpus must be listable")
	require.Contains(t, output, testutil.ControlSessionName,
		"the control session is not visible to query_sessions, so every tolerance assertion below would be vacuous — "+
			"check the fixture, this test's projects root, and the control session's cwd")
}

// corpusContext is the projects root and project path every corpus of one test
// is installed at, so the three responses stay comparable.
type corpusContext struct {
	projectsRoot string
	project      string
}

func newCorpusContext(t *testing.T) corpusContext {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	return corpusContext{projectsRoot: testutil.IsolatedProjectsRoot(t), project: project}
}

func (c corpusContext) install(t *testing.T, withControl bool, corrupt []testutil.CorruptFile) gateCorpus {
	t.Helper()
	return installCorpus(t, c.projectsRoot, c.project, withControl, corrupt)
}

// callTool runs one tool and requires the call itself to have succeeded. A
// whole-batch error on a corpus that still contains the usable control session
// is the headline regression this gate exists to catch, so it fails here rather
// than being folded into a verdict string.
func callTool(t *testing.T, g corpusToolGate, c gateCorpus) string {
	t.Helper()
	output, err := NewToolExecutor().ExecuteTool(&config.Config{}, g.name, g.argsFor(c))
	require.NoError(t, err, "%s: a corpus containing an unusable session file must not fail the whole call", g.name)
	require.NotEmpty(t, output, "%s: the call must still produce a response", g.name)
	return output
}

// baselineTool runs one tool against a corpus of ONLY unusable files, where a
// failure is itself an acceptable baseline: a corpus with no usable session at
// all has nothing to report on, so erroring there is informative rather than a
// tolerance violation. Nothing is required of this call except that it be
// distinguishable from the clean corpus — which is the whole point of taking
// the baseline.
func baselineTool(t *testing.T, g corpusToolGate, c gateCorpus) string {
	t.Helper()
	output, err := NewToolExecutor().ExecuteTool(&config.Config{}, g.name, g.argsFor(c))
	if err != nil {
		return "<no usable session: " + err.Error() + ">"
	}
	return output
}

// TestCorpusToleranceGate drives every registered tool against the shared
// corrupt corpus and applies its registered contract.
func TestCorpusToleranceGate(t *testing.T) {
	shapes := testutil.CorruptFiles()
	requireFixtureIsVisible(t)

	for _, gate := range corpusToolGates() {
		t.Run(gate.name, func(t *testing.T) {
			// Three corpora, installed over one another, so every assertion
			// below has a baseline:
			//   nothingUsable — the unusable files alone: what "the tool
			//                   returned nothing" looks like
			//   clean         — the control session alone: the expected result
			//   dirty         — both: what a tolerant tool must still return
			corpus := newCorpusContext(t)

			nothingUsable := corpus.install(t, false, shapes)
			nothingUsableOut := baselineTool(t, gate, nothingUsable)

			clean := corpus.install(t, true, nil)
			cleanOut := callTool(t, gate, clean)

			dirty := corpus.install(t, true, shapes)
			dirtyOut := callTool(t, gate, dirty)

			switch gate.kind {
			case kindEnumerationReported, kindEnumerationSilent:
				assert.Empty(t,
					enumerationVerdict(gate.name, nothingUsableOut, cleanOut, dirtyOut),
					strings.Join(enumerationVerdict(gate.name, nothingUsableOut, cleanOut, dirtyOut), "\n"))

				if gate.kind == kindEnumerationReported {
					assert.Empty(t,
						reportingVerdict(gate.name, dirtyOut, dirty.corruptNames()),
						"%s: the skip must be reported, not silent", gate.name)

					// The machine-readable channel, checked by name rather than
					// by prose match: a caller reconciles against skipped_files.
					skipped, ok := decodeField[[]interface{}](t, dirtyOut, "skipped_files")
					require.True(t, ok, "%s: a dirty corpus must expose skipped_files; got: %s", gate.name, dirtyOut)
					for _, name := range dirty.corruptNames() {
						assert.True(t, anyStringContains(skipped, name),
							"%s: skipped_files must list %s; got %v", gate.name, name, skipped)
					}
				}

			case kindSchemaSpecific:
				require.NotNil(t, gate.assertSurvives,
					"%s: a schema-specific entry must say how it survives the corrupt siblings", gate.name)
				gate.assertSurvives(t, clean, dirty, cleanOut, dirtyOut)
			}

			// Naming, for the shapes whose session ID the response must carry.
			for _, shape := range gate.mustName {
				name := corruptNameForShape(t, shape)
				assert.Contains(t, dirtyOut, name,
					"%s: the response must name the %s session file; got: %s", gate.name, shape, dirtyOut)
			}
		})
	}
}

// TestCorpusToleranceGate_EveryToolIsClassified is what makes "one
// registration point" mechanical: a tool definition that is neither registered
// in corpusToolGates() nor declared in nonCorpusTools fails here, so a newly
// added corpus-consuming tool cannot silently lack tolerance coverage.
func TestCorpusToleranceGate_EveryToolIsClassified(t *testing.T) {
	gated := make(map[string]bool)
	for _, gate := range corpusToolGates() {
		require.False(t, gated[gate.name], "corpusToolGates() registers %s twice", gate.name)
		gated[gate.name] = true
	}

	defined := make(map[string]bool)
	for _, definition := range toolspkg.GetToolDefinitions() {
		defined[definition.Name] = true

		if reason, excluded := nonCorpusTools[definition.Name]; excluded {
			assert.NotEmpty(t, reason,
				"%s is declared non-corpus-consuming without a reason", definition.Name)
			assert.False(t, gated[definition.Name],
				"%s is both registered in corpusToolGates() and declared non-corpus-consuming", definition.Name)
			continue
		}

		assert.True(t, gated[definition.Name],
			"tool %q is neither registered in corpusToolGates() nor declared in nonCorpusTools: "+
				"every session-corpus-consuming tool must be registered so its malformed-file tolerance is "+
				"checked (registration procedure: docs/guides/corpus-tolerance-gate.md)", definition.Name)
	}

	for name := range gated {
		assert.True(t, defined[name],
			"corpusToolGates() registers %q, which is not a defined MCP tool", name)
	}
	for name := range nonCorpusTools {
		assert.True(t, defined[name],
			"nonCorpusTools declares %q, which is not a defined MCP tool", name)
	}
}

// TestCorpusToleranceGate_RejectsAnIntolerantTool is the gate's negative
// control. A gate that quietly stopped checking anything would pass every test
// above, so this feeds the verdict helpers the response of a tool that DID
// lose the corpus because one sibling file was unusable — the DIR-094 defect —
// and fails if the gate would have accepted it.
func TestCorpusToleranceGate_RejectsAnIntolerantTool(t *testing.T) {
	cleanOut := `{"data":[{"session_id":"` + testutil.ControlSessionName + `"}],"warnings":[],"mode":"inline"}`
	nothingUsableOut := `{"data":[],"warnings":[],"mode":"inline"}`
	// The intolerant tool: one unusable file in the corpus and the call dies,
	// which is what the executor yields as an empty response plus an error.
	intolerantOut := ""

	assert.NotEmpty(t, enumerationVerdict("intolerant-scratch-tool", nothingUsableOut, cleanOut, intolerantOut),
		"the gate must reject a tool whose call fails when a sibling file is unusable")

	// And the silent-loss variant: the call succeeds but returns only the
	// nothing-usable baseline, having dropped the usable session entirely.
	assert.NotEmpty(t, enumerationVerdict("result-dropping-scratch-tool", nothingUsableOut, cleanOut, nothingUsableOut),
		"the gate must reject a tool that drops the usable session's results")

	// The conforming case must NOT be rejected, or the two checks above would
	// be satisfied by a verdict function that rejects everything.
	conformingOut := `{"data":[{"session_id":"` + testutil.ControlSessionName + `"}],"warnings":["skipped session file /x/` + testutil.ControlSessionName + `-sibling.jsonl: no message entries"],"mode":"inline"}`
	assert.Empty(t, enumerationVerdict("conforming-control", nothingUsableOut, cleanOut, conformingOut),
		"the gate must accept a tool that returns the clean corpus's results alongside a reported skip")

	assert.NotEmpty(t, reportingVerdict("silent-scratch-tool", `{"warnings":[]}`, []string{"bad-1"}),
		"the gate must reject a tool that skips a file without naming it")
	assert.Empty(t, reportingVerdict("reporting-control", `{"warnings":["skipped session file bad-1: unreadable"]}`, []string{"bad-1"}),
		"the gate must accept a tool that names every skipped file")
}

// ── verdict helpers ──────────────────────────────────────────────────────────
//
// These are pure functions over tool responses, deliberately free of
// *testing.T, so the negative control above can exercise the real checks the
// gate applies rather than a mock of them.

// enumerationVerdict returns the ways a tool's three responses violate the
// shared tolerance contract: the unusable files must not fail the call, must
// not change the results derived from the usable control session, and must not
// collapse the response to the nothing-usable baseline. Empty means it holds.
func enumerationVerdict(name, nothingUsableOut, cleanOut, dirtyOut string) []string {
	if strings.TrimSpace(dirtyOut) == "" {
		return []string{name + ": a corpus containing an unusable session file produced no response at all (whole-batch failure)"}
	}

	cleanPayload, err := stripExclusionMetadata(cleanOut)
	if err != nil {
		return []string{fmt.Sprintf("%s: clean-corpus response is not a JSON object: %v", name, err)}
	}
	dirtyPayload, err := stripExclusionMetadata(dirtyOut)
	if err != nil {
		return []string{fmt.Sprintf("%s: dirty-corpus response is not a JSON object: %v", name, err)}
	}

	var problems []string
	// Vacuity guard: the clean corpus must be distinguishable from a corpus
	// with nothing usable, or "dirty equals clean" below would pass for a tool
	// that returns the same empty answer either way. A baseline that is not a
	// response at all (the corpus had no usable session, so the tool failed)
	// already satisfies this by construction.
	if nothingUsablePayload, err := stripExclusionMetadata(nothingUsableOut); err == nil && nothingUsablePayload == cleanPayload {
		problems = append(problems, name+
			": the control session contributes nothing even on a clean corpus, so this entry cannot detect a loss — "+
			"the fixture or the tool arguments need fixing")
	}
	if dirtyPayload != cleanPayload {
		problems = append(problems, name+
			": results differ once unusable siblings are present; the unusable files must contribute nothing and "+
			"must not alter what the usable session produced\n  clean: "+cleanPayload+"\n  dirty: "+dirtyPayload)
	}
	return problems
}

// reportingVerdict returns the ways a tool failed to NAME a file it skipped.
// Empty means every skipped file is named.
func reportingVerdict(name, dirtyOut string, skipped []string) []string {
	var problems []string
	for _, sessionID := range skipped {
		if !strings.Contains(dirtyOut, sessionID) {
			problems = append(problems, fmt.Sprintf(
				"%s: skipped %s without naming it anywhere in the response; a silently shorter result set is the DIR-018 "+
					"contract violation this gate exists to catch", name, sessionID))
		}
	}
	return problems
}

// stripExclusionMetadata removes the exclusion-reporting keys from a response
// so the remainder can be compared against a clean-corpus control. Key order is
// normalised by json.Marshal, so equal payloads compare equal.
func stripExclusionMetadata(output string) (string, error) {
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		return "", err
	}
	delete(decoded, "warnings")
	delete(decoded, "skipped_files")

	encoded, err := json.Marshal(decoded)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// ── schema-specific survival checks ──────────────────────────────────────────

// assertDirectoryAccountsForEveryFile checks that the inventory still counts
// the control session once unusable siblings are present.
func assertDirectoryAccountsForEveryFile(t *testing.T, clean, dirty gateCorpus, cleanOut, dirtyOut string) {
	t.Helper()
	assert.Equal(t, 1, jsonIntField(t, cleanOut, "file_count"),
		"the clean corpus holds exactly the control session")
	assert.Equal(t, 1+len(dirty.corrupt), jsonIntField(t, dirtyOut, "file_count"),
		"the inventory must account for the control session AND every unusable file, not drop them")
	assert.Greater(t, jsonIntField(t, dirtyOut, "total_size_bytes"), jsonIntField(t, cleanOut, "total_size_bytes"),
		"the control session's bytes must still be counted")
}

// assertMetadataListsEveryFile checks that the per-file listing still carries
// the control session's real record count alongside the unusable entries.
func assertMetadataListsEveryFile(t *testing.T, clean, dirty gateCorpus, cleanOut, dirtyOut string) {
	t.Helper()
	assert.Equal(t, controlRecordCount, recordsForPath(t, cleanOut, clean.path(testutil.ControlSessionName)),
		"clean corpus: the control session's record count")
	assert.Equal(t, controlRecordCount, recordsForPath(t, dirtyOut, dirty.path(testutil.ControlSessionName)),
		"the control session must keep its full record count when unusable siblings are present")
	for _, name := range dirty.corruptNames() {
		assert.Contains(t, dirtyOut, dirty.path(name),
			"an unusable file must appear in the listing, not vanish from it")
	}
}

// assertInspectReportsControlIntact checks that the control session is still
// inspected at full length while the unusable files are reported as unusable.
func assertInspectReportsControlIntact(t *testing.T, clean, dirty gateCorpus, cleanOut, dirtyOut string) {
	t.Helper()
	controlPath := dirty.path(testutil.ControlSessionName)
	assert.Equal(t, controlRecordCount, recordsForPath(t, dirtyOut, controlPath),
		"the control session must be inspected at full length")
	assert.Equal(t, recordsForPath(t, cleanOut, clean.path(testutil.ControlSessionName)),
		recordsForPath(t, dirtyOut, controlPath),
		"the control session's inspection must be identical with and without unusable siblings")
	for _, name := range dirty.corruptNames() {
		assert.True(t, anyStringContains(pathsInFilesField(t, dirtyOut), dirty.path(name)),
			"an unusable file must be reported, not skipped silently")
	}
}

// assertStage2KeepsControlRecords checks that the jq results still carry the
// control session's records once unusable files are in the file list.
func assertStage2KeepsControlRecords(t *testing.T, clean, dirty gateCorpus, cleanOut, dirtyOut string) {
	t.Helper()
	cleanReturned := jsonIntFieldAt(t, cleanOut, "metadata", "results_returned")
	require.NotZero(t, cleanReturned, "the control session must produce jq matches, or this entry proves nothing")
	assert.Equal(t, cleanReturned, jsonIntFieldAt(t, dirtyOut, "metadata", "results_returned"),
		"the unusable files must not change how many records the usable session yields")
	assert.Contains(t, dirtyOut, dir099ControlMarker,
		"the control session's content must still be in the results")
}

// assertEditSequencesUnchanged pins the only property this corpus can express
// for this tool: the unusable files neither fail the call nor perturb a result
// that is legitimately empty even when clean.
func assertEditSequencesUnchanged(t *testing.T, clean, dirty gateCorpus, cleanOut, dirtyOut string) {
	t.Helper()
	cleanPayload, err := stripExclusionMetadata(cleanOut)
	require.NoError(t, err)
	dirtyPayload, err := stripExclusionMetadata(dirtyOut)
	require.NoError(t, err)
	assert.Equal(t, cleanPayload, dirtyPayload,
		"unusable files must not perturb the result")
}

// ── small JSON helpers ───────────────────────────────────────────────────────

// controlRecordCount is the control session's entry count. Changing the
// fixture means changing this and the shape it is asserted against.
const controlRecordCount = 10

// dir099ControlMarker is a string that appears only in the control session's
// content, used to prove the control session reached a tool's results.
const dir099ControlMarker = "authError: token invalid"

func decodeField[T any](t *testing.T, output, field string) (T, bool) {
	t.Helper()
	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &decoded), "response must be a JSON object; got: %s", output)
	raw, ok := decoded[field]
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := raw.(T)
	return typed, ok
}

func jsonIntField(t *testing.T, output, field string) int {
	t.Helper()
	value, ok := decodeField[float64](t, output, field)
	require.True(t, ok, "response must carry numeric %q; got: %s", field, output)
	return int(value)
}

// jsonIntFieldAt reads a numeric field through a path of nested objects.
func jsonIntFieldAt(t *testing.T, output string, path ...string) int {
	t.Helper()
	var current interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &current), "response must be a JSON object; got: %s", output)
	for _, key := range path {
		object, ok := current.(map[string]interface{})
		require.True(t, ok, "response must be a JSON object at %v; got: %s", path, output)
		current, ok = object[key]
		require.True(t, ok, "response must carry %v; got: %s", path, output)
	}
	value, ok := current.(float64)
	require.True(t, ok, "response field %v must be numeric; got: %s", path, output)
	return int(value)
}

// recordsForPath returns the record count a listing tool reported for one path.
func recordsForPath(t *testing.T, output, path string) int {
	t.Helper()
	for _, entry := range fileEntries(t, output) {
		if entry["path"] != path {
			continue
		}
		for _, key := range []string{"records", "line_count"} {
			if value, ok := entry[key].(float64); ok {
				return int(value)
			}
		}
	}
	t.Fatalf("no listing entry for %s; got: %s", path, output)
	return 0
}

func pathsInFilesField(t *testing.T, output string) []interface{} {
	t.Helper()
	paths := make([]interface{}, 0)
	for _, entry := range fileEntries(t, output) {
		if path, ok := entry["path"].(string); ok {
			paths = append(paths, path)
		}
	}
	return paths
}

func fileEntries(t *testing.T, output string) []map[string]interface{} {
	t.Helper()
	files, ok := decodeField[[]interface{}](t, output, "files")
	require.True(t, ok, "response must carry a files array; got: %s", output)
	entries := make([]map[string]interface{}, 0, len(files))
	for _, raw := range files {
		entry, ok := raw.(map[string]interface{})
		require.True(t, ok, "each files entry must be an object; got: %v", raw)
		entries = append(entries, entry)
	}
	return entries
}

// corruptNameForShape resolves a shape to the session ID the fixture installs
// it under, so entries can name shapes without hard-coding IDs.
func corruptNameForShape(t *testing.T, shape testutil.CorruptShape) string {
	t.Helper()
	for _, file := range testutil.CorruptFiles() {
		if file.Shape == shape {
			return file.Name
		}
	}
	t.Fatalf("the shared fixture has no corrupt file with shape %q", shape)
	return ""
}
