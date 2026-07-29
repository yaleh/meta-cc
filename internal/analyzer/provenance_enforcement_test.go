package analyzer

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

// provenanceExemptTypes is intentionally small. Every exported analyzer struct
// ending in Result or Stats is discovered mechanically; only truly internal,
// non-serialized aggregates may be exempted here with a justification.
var provenanceExemptTypes = map[string]string{
	"SessionStats": "internal CalculateStats aggregate with no JSON serialization tags",
}

// TestProvenanceConvention_DataSourceField parses the analyzer package and
// asserts every exported *Result / *Stats struct has a DataSource field of type
// DataSource serialized under the JSON key "data_source". Because declarations
// are enumerated from source, adding a result type cannot silently bypass this
// test by being omitted from a manually maintained registry.
func TestProvenanceConvention_DataSourceField(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info fs.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse analyzer package: %v", err)
	}
	pkg, ok := pkgs["analyzer"]
	if !ok {
		t.Fatal("parsed analyzer package not found")
	}

	structs := exportedResultStructs(pkg.Files)
	if len(structs) == 0 {
		t.Fatal("no exported *Result / *Stats structs discovered; ADR-007 enforcement would pass vacuously")
	}
	for name := range provenanceExemptTypes {
		if _, ok := structs[name]; !ok {
			t.Errorf("stale ADR-007 exemption %s: no matching exported *Result / *Stats struct", name)
		}
	}
	for name, structType := range structs {
		name, structType := name, structType
		t.Run(name, func(t *testing.T) {
			if reason, exempt := provenanceExemptTypes[name]; exempt {
				if strings.TrimSpace(reason) == "" {
					t.Fatalf("ADR-007 exemption %s has no justification", name)
				}
				return
			}
			if err := validateProvenanceStruct(name, structType); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func exportedResultStructs(files map[string]*ast.File) map[string]*ast.StructType {
	result := make(map[string]*ast.StructType)
	for _, file := range files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				structType, ok := typeSpec.Type.(*ast.StructType)
				if ok && ast.IsExported(typeSpec.Name.Name) && (strings.HasSuffix(typeSpec.Name.Name, "Result") || strings.HasSuffix(typeSpec.Name.Name, "Stats")) {
					result[typeSpec.Name.Name] = structType
				}
			}
		}
	}
	return result
}

func validateProvenanceStruct(name string, structType *ast.StructType) error {
	for _, field := range structType.Fields.List {
		if len(field.Names) != 1 || field.Names[0].Name != "DataSource" {
			continue
		}
		fieldType, ok := field.Type.(*ast.Ident)
		if !ok || fieldType.Name != "DataSource" {
			return &provenanceViolation{name + ".DataSource must have type analyzer.DataSource"}
		}
		if field.Tag == nil {
			return &provenanceViolation{name + ".DataSource must have JSON key data_source"}
		}
		tag, err := strconv.Unquote(field.Tag.Value)
		if err != nil || strings.Split(reflect.StructTag(tag).Get("json"), ",")[0] != "data_source" {
			return &provenanceViolation{name + ".DataSource must have JSON key data_source"}
		}
		return nil
	}
	return &provenanceViolation{name + " is a governed analysis result type but has no DataSource field"}
}

type provenanceViolation struct{ message string }

func (e *provenanceViolation) Error() string { return "ADR-007 violation: " + e.message }

// TestProvenanceConvention_DiscoversNewResultOmissions is the regression proof
// for the audit finding: a newly declared exported result is discovered from
// source and rejected without requiring any registry update.
func TestProvenanceConvention_DiscoversNewResultOmissions(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "synthetic.go", "package analyzer\ntype SyntheticResult struct { Value int `json:\"value\"` }", 0)
	if err != nil {
		t.Fatalf("parse synthetic result: %v", err)
	}
	structs := exportedResultStructs(map[string]*ast.File{"synthetic.go": file})
	structType, ok := structs["SyntheticResult"]
	if !ok {
		t.Fatal("new exported SyntheticResult was not discovered")
	}
	if err := validateProvenanceStruct("SyntheticResult", structType); err == nil || !strings.Contains(err.Error(), "no DataSource field") {
		t.Fatalf("omitted DataSource was not detected; got %v", err)
	}
}

// TestProvenanceConvention_EstimatedFieldsField avoids a second, partial
// registry for mixed-provenance outputs: every governed result declaration
// carries the same optional EstimatedFields field. Pure results leave it empty;
// mixed results populate it with the heuristic-derived JSON paths.
func TestProvenanceConvention_EstimatedFieldsField(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info fs.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse analyzer package: %v", err)
	}

	structs := exportedResultStructs(pkgs["analyzer"].Files)
	for name, structType := range structs {
		if _, exempt := provenanceExemptTypes[name]; exempt {
			continue
		}
		t.Run(name, func(t *testing.T) {
			if err := validateEstimatedFieldsStruct(name, structType); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func validateEstimatedFieldsStruct(name string, structType *ast.StructType) error {
	for _, field := range structType.Fields.List {
		if len(field.Names) != 1 || field.Names[0].Name != "EstimatedFields" {
			continue
		}
		arrayType, ok := field.Type.(*ast.ArrayType)
		element, elementOK := arrayType.Elt.(*ast.Ident)
		if !ok || arrayType.Len != nil || !elementOK || element.Name != "string" {
			return &provenanceViolation{name + ".EstimatedFields must have type []string"}
		}
		if field.Tag == nil {
			return &provenanceViolation{name + ".EstimatedFields must have JSON key estimated_fields"}
		}
		tag, err := strconv.Unquote(field.Tag.Value)
		if err != nil || strings.Split(reflect.StructTag(tag).Get("json"), ",")[0] != "estimated_fields" {
			return &provenanceViolation{name + ".EstimatedFields must have JSON key estimated_fields"}
		}
		return nil
	}
	return &provenanceViolation{name + " is a governed analysis result type but has no EstimatedFields field"}
}

// TestProvenanceConvention_DiscoversEstimatedFieldsOmissions proves a new result
// cannot evade mixed-provenance support through a partial type registry.
func TestProvenanceConvention_DiscoversEstimatedFieldsOmissions(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "synthetic.go", "package analyzer\ntype SyntheticResult struct { DataSource DataSource `json:\"data_source\"` }", 0)
	if err != nil {
		t.Fatalf("parse synthetic result: %v", err)
	}
	structType := exportedResultStructs(map[string]*ast.File{"synthetic.go": file})["SyntheticResult"]
	if err := validateEstimatedFieldsStruct("SyntheticResult", structType); err == nil || !strings.Contains(err.Error(), "no EstimatedFields field") {
		t.Fatalf("omitted EstimatedFields was not detected; got %v", err)
	}
}

// TestGetTechDebt_SurfacesEstimatedOpenIssues verifies the runtime behavior of
// the ADR-007 rule: a session-scan tech-debt result keeps the dominant
// "measured" provenance and names "open_issues" in EstimatedFields. A
// source-dir scan instead names "markers" and "hotspot_files" because its
// language-aware markerScanner is a lexical heuristic, not a parser.
func TestGetTechDebt_SurfacesEstimatedOpenIssues(t *testing.T) {
	session, err := GetTechDebt(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.DataSource != DataSourceMeasured {
		t.Errorf("session DataSource = %q, want %q", session.DataSource, DataSourceMeasured)
	}
	if !containsString(session.EstimatedFields, "open_issues") {
		t.Errorf("session EstimatedFields = %v, want it to contain %q", session.EstimatedFields, "open_issues")
	}

	tmp := t.TempDir()
	source, err := ScanSourceDir(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, field := range []string{"markers", "hotspot_files"} {
		if !containsString(source.EstimatedFields, field) {
			t.Errorf("source EstimatedFields = %v, want it to contain %q", source.EstimatedFields, field)
		}
	}

	data, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("marshal source result: %v", err)
	}
	var decoded struct {
		EstimatedFields []string `json:"estimated_fields"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal source result: %v", err)
	}
	for _, field := range []string{"markers", "hotspot_files"} {
		if !containsString(decoded.EstimatedFields, field) {
			t.Errorf("source JSON estimated_fields = %v, want it to contain %q; JSON: %s", decoded.EstimatedFields, field, data)
		}
	}
}

// TestGetWorkPatterns_SurfacesEstimatedContextSwitches verifies the runtime
// behavior of the ADR-007 rule for work patterns: dominant "measured"
// provenance with "context_switches" named in EstimatedFields.
func TestGetWorkPatterns_SurfacesEstimatedContextSwitches(t *testing.T) {
	result, err := GetWorkPatterns(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("DataSource = %q, want %q", result.DataSource, DataSourceMeasured)
	}
	if !containsString(result.EstimatedFields, "context_switches") {
		t.Errorf("EstimatedFields = %v, want it to contain %q", result.EstimatedFields, "context_switches")
	}
}

func TestAnalyzeBugs_SurfacesEstimatedCausalPairing(t *testing.T) {
	toolCalls := []types.ToolCall{
		{ToolName: "Bash", Status: "error", Error: "failed"},
		{ToolName: "Bash", Status: "success"},
	}
	full, err := AnalyzeBugs(nil, toolCalls, 1)
	if err != nil {
		t.Fatalf("AnalyzeBugs: %v", err)
	}
	for _, field := range []string{"patterns", "total_pairs"} {
		if !containsString(full.EstimatedFields, field) {
			t.Errorf("BugAnalysisResult.EstimatedFields = %v, want %q", full.EstimatedFields, field)
		}
	}

	stats, err := AnalyzeBugsStats(nil, toolCalls)
	if err != nil {
		t.Fatalf("AnalyzeBugsStats: %v", err)
	}
	for _, field := range []string{"patterns", "total_pairs", "total_patterns"} {
		if !containsString(stats.EstimatedFields, field) {
			t.Errorf("BugAnalysisStats.EstimatedFields = %v, want %q", stats.EstimatedFields, field)
		}
	}
}

func TestQualityScan_SurfacesEstimatedRetryRate(t *testing.T) {
	result, err := QualityScan(nil, []types.ToolCall{
		{ToolName: "Read", Status: "error", Error: "failed"},
		{ToolName: "Read", Status: "success"},
	})
	if err != nil {
		t.Fatalf("QualityScan: %v", err)
	}
	if !containsString(result.EstimatedFields, "dimensions") {
		t.Errorf("QualityScanResult.EstimatedFields = %v, want dimensions", result.EstimatedFields)
	}
}

func TestMixedProvenanceJSONIncludesEstimatedFields(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
		want  []string
	}{
		{"quality", QualityScanResult{DataSource: DataSourceMeasured, EstimatedFields: []string{"dimensions"}}, []string{"dimensions"}},
		{"bugs", BugAnalysisResult{DataSource: DataSourceMeasured, EstimatedFields: []string{"patterns", "total_pairs"}}, []string{"patterns", "total_pairs"}},
		{"bug_stats", BugAnalysisStats{DataSource: DataSourceMeasured, EstimatedFields: []string{"patterns", "total_pairs", "total_patterns"}}, []string{"patterns", "total_pairs", "total_patterns"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded struct {
				EstimatedFields []string `json:"estimated_fields"`
			}
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(decoded.EstimatedFields, tc.want) {
				t.Errorf("estimated_fields = %v, want %v; JSON: %s", decoded.EstimatedFields, tc.want, data)
			}
		})
	}
}

func assertEstimatedJSONPathsExist(t *testing.T, value interface{}) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal serialized output: %v", err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("decode serialized output: %v", err)
	}
	fields, _ := root["estimated_fields"].([]interface{})
	for _, raw := range fields {
		path, ok := raw.(string)
		if !ok {
			t.Fatalf("estimated_fields contains non-string %T", raw)
		}
		if !serializedPathExists(root, strings.Split(path, ".")) {
			t.Errorf("estimated_fields path %q does not exist in serialized output: %s", path, data)
		}
	}
}

func serializedPathExists(value interface{}, path []string) bool {
	if len(path) == 0 {
		return true
	}
	switch current := value.(type) {
	case map[string]interface{}:
		if path[0] == "*" {
			if len(current) == 0 {
				return false
			}
			for _, child := range current {
				if !serializedPathExists(child, path[1:]) {
					return false
				}
			}
			return true
		}
		child, ok := current[path[0]]
		return ok && serializedPathExists(child, path[1:])
	case []interface{}:
		if len(current) == 0 {
			return false
		}
		if path[0] == "*" {
			path = path[1:]
		}
		for _, child := range current {
			if !serializedPathExists(child, path) {
				return false
			}
		}
		return true
	}
	return false
}

func TestEstimatedFieldsPathsExistInRuntimeJSON(t *testing.T) {
	calls := []types.ToolCall{{ToolName: "Read", Status: "error", Error: "not found"}, {ToolName: "Read", Status: "success"}}
	quality, _ := QualityScan(nil, calls)
	bugs, _ := AnalyzeBugs(nil, calls, 1)
	bugStats, _ := AnalyzeBugsStats(nil, calls)
	errors, _ := AnalyzeErrors(nil, calls, 1)
	errorStats, _ := AnalyzeErrorsStats(nil, calls)
	work, _ := GetWorkPatterns(nil, nil)
	editEntries := buildEntries("edit-session", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"edit-read", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/src.go"}},
		{"edit-doc", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
	})
	edit := BuildEditSequences(editEntries, nil, false, 10)
	// The empty result still advertises summary.patternDistribution; the
	// universal path validator must confirm every advertised path exists.
	emptyEdit := BuildEditSequences(nil, nil, false, 0)

	for name, value := range map[string]interface{}{
		"quality": quality, "bugs": bugs, "bug_stats": bugStats,
		"errors": errors, "error_stats": errorStats, "work": work,
		"edit_sequences": edit, "edit_sequences_empty": emptyEdit,
	} {
		t.Run(name, func(t *testing.T) { assertEstimatedJSONPathsExist(t, value) })
	}
}

func TestEditSequencesEstimatedFieldsMatchConcreteJSONShapes(t *testing.T) {
	makeResult := func(paths ...string) EditSequencesResult {
		specs := make([]struct {
			uuid      string
			timestamp string
			tool      string
			input     map[string]interface{}
		}, len(paths))
		for i, path := range paths {
			specs[i] = struct {
				uuid      string
				timestamp string
				tool      string
				input     map[string]interface{}
			}{itoa(i), "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": path}}
		}
		return BuildEditSequences(buildEntries("shape-session", specs), nil, false, 10)
	}

	cases := []struct {
		name      string
		result    EditSequencesResult
		want      []string
		doNotWant []string
	}{
		// An empty result still serializes the heuristic summary.patternDistribution
		// zeros, so that path must be advertised; every absent optional file/event
		// path stays omitted (ADR-007).
		{"empty", makeResult(),
			[]string{"summary.patternDistribution"},
			[]string{
				"files.*.events.*.fileType",
				"files.*.patternHint",
				"files.*.docVoid",
				"files.*.specPrecisionGap",
				"files.*.events.*.docRole",
				"files.*.coAccessedDocs.*.docRole",
			}},
		{"source_only", makeResult("/src.go"), []string{"files.*.events.*.fileType", "files.*.patternHint"}, []string{"files.*.events.*.docRole", "files.*.coAccessedDocs.*.docRole"}},
		{"documents", makeResult("/A.md", "/B.md"), []string{"files.*.events.*.docRole", "files.*.coAccessedDocs.*.docRole"}, nil},
		{"mixed", makeResult("/src.go", "/SPEC.md"), []string{"files.*.events.*.fileType"}, []string{"files.*.events.*.docRole", "files.*.coAccessedDocs.*.docRole"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertEstimatedJSONPathsExist(t, tc.result)
			for _, path := range tc.want {
				if !containsString(tc.result.EstimatedFields, path) {
					t.Errorf("EstimatedFields = %v, want %q", tc.result.EstimatedFields, path)
				}
			}
			for _, path := range tc.doNotWant {
				if containsString(tc.result.EstimatedFields, path) {
					t.Errorf("EstimatedFields = %v, must not advertise absent optional path %q", tc.result.EstimatedFields, path)
				}
			}
		})
	}
}
func TestTechDebtEstimatedPathsExistForResultMergeAndStatsJSON(t *testing.T) {
	session, err := GetTechDebt(nil, []types.ToolCall{{ToolName: "Read", Status: "error"}})
	if err != nil {
		t.Fatalf("GetTechDebt: %v", err)
	}
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "x.go"), []byte("package x\n// TODO: fix\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	source, err := ScanSourceDir(sourceDir)
	if err != nil {
		t.Fatalf("ScanSourceDir: %v", err)
	}
	merged := MergeTechDebtResults(session, source, DataSourceMeasured)
	stats := TechDebtResultStats(merged)

	assertEstimatedJSONPathsExist(t, session)
	assertEstimatedJSONPathsExist(t, source)
	assertEstimatedJSONPathsExist(t, merged)
	assertEstimatedJSONPathsExist(t, stats)

	wantStats := []string{"open_issue_count", "markers", "marker_count", "hotspot_file_count"}
	for _, field := range wantStats {
		if !containsString(stats.EstimatedFields, field) {
			t.Errorf("TechDebtStats.EstimatedFields = %v, want %q", stats.EstimatedFields, field)
		}
	}
	for _, stale := range []string{"open_issues", "hotspot_files"} {
		if containsString(stats.EstimatedFields, stale) {
			t.Errorf("TechDebtStats.EstimatedFields retains result-only path %q: %v", stale, stats.EstimatedFields)
		}
	}
}
func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
