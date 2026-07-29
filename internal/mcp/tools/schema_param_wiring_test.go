package tools

// DIR-045 statically verifies that every effective BuildTool schema parameter
// is read on that tool's concrete registered execution path.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// handlerScopedExemptions is intentionally keyed by tool and parameter. Never
// exempt a parameter name globally: the same name may be wired on one execution
// path and silently ignored on another.
var handlerScopedExemptions = map[string]map[string]string{}

// commonExecutionEntryPoints contains only work performed before ExecuteTool's
// mutually exclusive special/query dispatch. Starting every tool at ExecuteTool
// would make the static call graph flow through both branches and let special
// tools inherit query-only reads from NewToolPipelineConfig/BuildResponse.
var commonExecutionEntryPoints = []string{"DetermineScope"}

// specialExecutionEntryPoints models the branch taken for registerHandler tools.
var specialExecutionEntryPoints = []string{"ExecuteSpecialTool"}

// centralPipelineEntryPoints are always unioned into the reachable-code set
// for tools dispatched via registerQueryHandler (see queryHandlerRegistry),
// because executor.ExecuteTool always threads every such tool's args through
// these two functions after the handler returns -- e.g. output_format is read
// in NewToolPipelineConfig, not inside any individual query handler.
var centralPipelineEntryPoints = []string{"NewToolPipelineConfig", "BuildResponse"}

type funcNode struct {
	text  string
	calls []string
}

// readArgPattern matches Get*Param(args, "name" / Get*Param(params, "name" /
// stringArg(args, "name" / parseOptionalRFC3339(args, "name" -- i.e. any call
// whose first argument identifier looks like an args/params map and whose
// second argument is a string literal.
var readArgPattern = regexp.MustCompile(`[A-Za-z0-9_]*(?:[Aa]rgs|[Pp]arams)[A-Za-z0-9_]*\s*,\s*"([A-Za-z0-9_]+)"`)

// readIndexPattern matches raw map indexing: args["name"] / params["name"].
var readIndexPattern = regexp.MustCompile(`[A-Za-z0-9_]*(?:[Aa]rgs|[Pp]arams)[A-Za-z0-9_]*\s*\[\s*"([A-Za-z0-9_]+)"\s*\]`)

func thisDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to resolve this test file's path")
	}
	return filepath.Dir(file)
}

func extractCalleeName(call *ast.CallExpr) (string, bool) {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name, true
	case *ast.SelectorExpr:
		return fn.Sel.Name, true
	}
	return "", false
}

func collectCalls(node ast.Node) []string {
	var calls []string
	ast.Inspect(node, func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			if name, ok := extractCalleeName(ce); ok {
				calls = append(calls, name)
			}
		}
		return true
	})
	return calls
}

func sliceText(src []byte, fset *token.FileSet, start, end token.Pos) string {
	return string(src[fset.Position(start).Offset:fset.Position(end).Offset])
}

type registryEntry struct {
	class      string // "special" or "query"
	startNames []string
}

// analyzeSources parses every non-test .go file under dirs, returning (1) a
// name -> funcNode map for every top-level func/method declaration (used to
// build the reachable-code call graph) and (2) the tool -> registryEntry
// map extracted from registerHandler/registerQueryHandler calls found in
// executorDir.
func analyzeSources(dirs []string, executorDir string) (map[string]funcNode, map[string]registryEntry, error) {
	funcs := map[string]funcNode{}
	registry := map[string]registryEntry{}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, nil, fmt.Errorf("reading dir %s: %w", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(dir, name)
			src, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, fmt.Errorf("reading %s: %w", path, err)
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, src, 0)
			if err != nil {
				return nil, nil, fmt.Errorf("parsing %s: %w", path, err)
			}

			isExecutorFile := dir == executorDir

			for _, decl := range file.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Body == nil {
					continue
				}
				funcs[fd.Name.Name] = funcNode{
					text:  sliceText(src, fset, fd.Body.Pos(), fd.Body.End()),
					calls: collectCalls(fd.Body),
				}

				if isExecutorFile {
					registerCalls, err := extractRegistryCalls(fd.Body, src, fset, funcs)
					if err != nil {
						return nil, nil, err
					}
					for tool, entry := range registerCalls {
						registry[tool] = entry
					}
				}
			}
		}
	}
	return funcs, registry, nil
}

// extractRegistryCalls scans a func body (expected to be an init() func) for
// registerHandler("name", handlerExpr) / registerQueryHandler("name", handlerExpr)
// calls. handlerExpr is either a bare identifier referencing an existing
// funcs[] entry, or an inline func literal -- in the latter case a synthetic
// funcs[] entry is created so the call graph can flow through it uniformly.
func extractRegistryCalls(body ast.Node, src []byte, fset *token.FileSet, funcs map[string]funcNode) (map[string]registryEntry, error) {
	out := map[string]registryEntry{}
	var walkErr error
	ast.Inspect(body, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := ce.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		var class string
		switch ident.Name {
		case "registerHandler":
			class = "special"
		case "registerQueryHandler":
			class = "query"
		default:
			return true
		}
		if len(ce.Args) < 2 {
			return true
		}
		lit, ok := ce.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		toolName, err := strconv.Unquote(lit.Value)
		if err != nil {
			walkErr = fmt.Errorf("unquoting tool name %s: %w", lit.Value, err)
			return false
		}

		var startNames []string
		switch h := ce.Args[1].(type) {
		case *ast.Ident:
			startNames = []string{h.Name}
		case *ast.FuncLit:
			synthetic := "__registration_" + toolName
			funcs[synthetic] = funcNode{
				text:  sliceText(src, fset, h.Body.Pos(), h.Body.End()),
				calls: collectCalls(h.Body),
			}
			startNames = []string{synthetic}
		default:
			walkErr = fmt.Errorf("tool %q: unrecognized handler expression shape %T in registration call", toolName, h)
			return false
		}
		out[toolName] = registryEntry{class: class, startNames: startNames}
		return true
	})
	return out, walkErr
}

// reachableParamNames does a BFS over funcs starting at starts, and returns
// the set of parameter names found (via readArgPattern/readIndexPattern)
// anywhere in the reachable functions' source text.
func reachableParamNames(starts []string, funcs map[string]funcNode) map[string]bool {
	visited := map[string]bool{}
	queue := append([]string{}, starts...)
	var allText strings.Builder

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if visited[name] {
			continue
		}
		visited[name] = true
		fn, ok := funcs[name]
		if !ok {
			continue // crosses out of scanned packages (e.g. into internal/analyzer) -- nothing more to find
		}
		allText.WriteString(fn.text)
		allText.WriteString("\n")
		queue = append(queue, fn.calls...)
	}

	found := map[string]bool{}
	text := allText.String()
	for _, m := range readArgPattern.FindAllStringSubmatch(text, -1) {
		found[m[1]] = true
	}
	for _, m := range readIndexPattern.FindAllStringSubmatch(text, -1) {
		found[m[1]] = true
	}
	return found
}

func executionStarts(entry registryEntry) ([]string, error) {
	starts := append([]string{}, commonExecutionEntryPoints...)
	starts = append(starts, entry.startNames...)
	switch entry.class {
	case "special":
		return append(starts, specialExecutionEntryPoints...), nil
	case "query":
		return append(starts, centralPipelineEntryPoints...), nil
	default:
		return nil, fmt.Errorf("unknown registry class %q", entry.class)
	}
}

func unwiredParams(tool string, props []string, entry registryEntry, funcs map[string]funcNode) ([]string, error) {
	starts, err := executionStarts(entry)
	if err != nil {
		return nil, err
	}
	reachable := reachableParamNames(starts, funcs)
	var gaps []string
	for _, p := range props {
		if reason := handlerScopedExemptions[tool][p]; reason != "" {
			continue
		}
		if !reachable[p] {
			gaps = append(gaps, p)
		}
	}
	return gaps, nil
}

// extractBuildToolProperties parses toolsGoPath and returns, for every
// BuildTool(...) call, the tool name and the set of Property keys declared
// directly in that call's own properties map literal (i.e. NOT including
// names merged in later by MergeParameters/StandardToolParameters).
func extractBuildToolProperties(toolsGoPath string) (map[string][]string, error) {
	src, err := os.ReadFile(toolsGoPath)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, toolsGoPath, src, 0)
	if err != nil {
		return nil, err
	}

	result := map[string][]string{}
	var walkErr error
	ast.Inspect(file, func(n ast.Node) bool {
		if walkErr != nil {
			return false
		}
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := ce.Fun.(*ast.Ident)
		if !ok || ident.Name != "BuildTool" {
			return true
		}
		if len(ce.Args) < 3 {
			walkErr = fmt.Errorf("BuildTool call at %s has fewer than 3 args, parser out of date", fset.Position(ce.Pos()))
			return false
		}
		nameLit, ok := ce.Args[0].(*ast.BasicLit)
		if !ok || nameLit.Kind != token.STRING {
			walkErr = fmt.Errorf("BuildTool call at %s: first arg is not a string literal, parser out of date", fset.Position(ce.Pos()))
			return false
		}
		toolName, err := strconv.Unquote(nameLit.Value)
		if err != nil {
			walkErr = fmt.Errorf("unquoting tool name %s: %w", nameLit.Value, err)
			return false
		}

		props, ok := ce.Args[2].(*ast.CompositeLit)
		if !ok {
			walkErr = fmt.Errorf("BuildTool(%q, ...) properties arg is not a composite literal (%T), parser out of date", toolName, ce.Args[2])
			return false
		}

		var keys []string
		for _, elt := range props.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				walkErr = fmt.Errorf("BuildTool(%q, ...) properties map has a non-KeyValueExpr element, parser out of date", toolName)
				return false
			}
			keyLit, ok := kv.Key.(*ast.BasicLit)
			if !ok || keyLit.Kind != token.STRING {
				walkErr = fmt.Errorf("BuildTool(%q, ...) properties map has a non-string-literal key, parser out of date", toolName)
				return false
			}
			key, err := strconv.Unquote(keyLit.Value)
			if err != nil {
				walkErr = fmt.Errorf("unquoting property key %s: %w", keyLit.Value, err)
				return false
			}
			keys = append(keys, key)
		}
		result[toolName] = keys
		return true
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return result, nil
}

func sourceAnalysisPaths(base string) ([]string, string) {
	executorDir := filepath.Join(base, "..", "executor")
	return []string{
		executorDir,
		filepath.Join(base, "..", "pipeline"),
		filepath.Join(base, "..", "response"),
		filepath.Join(base, "..", "query"),
		filepath.Join(base, "..", "..", "analysis"),
	}, executorDir
}

// TestSchemaGateRejectsQueryOnlyParamOnAnalysisTool is the audit regression:
// adding jq_filter to analyze_errors must fail because that special-handler path
// returns before the mutually exclusive query response pipeline reads it.
func TestSchemaGateRejectsQueryOnlyParamOnAnalysisTool(t *testing.T) {
	base := thisDir(t)
	dirs, executorDir := sourceAnalysisPaths(base)
	funcs, registry, err := analyzeSources(dirs, executorDir)
	if err != nil {
		t.Fatalf("analyzing handler/pipeline sources: %v", err)
	}
	entry, ok := registry["analyze_errors"]
	if !ok {
		t.Fatal("analyze_errors registration not found")
	}

	gaps, err := unwiredParams("analyze_errors", []string{"jq_filter"}, entry, funcs)
	if err != nil {
		t.Fatal(err)
	}
	if len(gaps) != 1 || gaps[0] != "jq_filter" {
		t.Fatalf("adding unused jq_filter to analyze_errors must fail the gate; gaps=%v", gaps)
	}
}

// TestSchemaParamsAreWired is the DIR-045 gate: every tool-specific Property
// declared in a BuildTool(...) call in tools.go must have a corresponding
// read reachable from that tool's registered handler (directly, or via the
// central pipeline stage for query-handler-registry tools), unless the
// parameter name is on the documented allowedGenericParams list above.
func TestSchemaParamsAreWired(t *testing.T) {
	base := thisDir(t)
	dirs, executorDir := sourceAnalysisPaths(base)

	funcs, registry, err := analyzeSources(dirs, executorDir)
	if err != nil {
		t.Fatalf("analyzing handler/pipeline sources: %v", err)
	}

	toolProps, err := extractBuildToolProperties(filepath.Join(base, "tools.go"))
	if err != nil {
		t.Fatalf("extracting BuildTool(...) schema properties: %v", err)
	}
	// Replace the source-literal keys with each BuildTool's effective runtime
	// schema, including only the shared parameters selected for that tool's path.
	for _, definition := range GetToolDefinitions() {
		if _, built := toolProps[definition.Name]; !built {
			continue
		}
		keys := make([]string, 0, len(definition.InputSchema.Properties))
		for key := range definition.InputSchema.Properties {
			keys = append(keys, key)
		}
		toolProps[definition.Name] = keys
	}
	if len(toolProps) == 0 {
		t.Fatal("extracted zero BuildTool(...) tools from tools.go -- parser is almost certainly broken")
	}

	var toolNames []string
	for name := range toolProps {
		toolNames = append(toolNames, name)
	}
	sort.Strings(toolNames)

	var gaps []string
	for _, tool := range toolNames {
		entry, ok := registry[tool]
		if !ok {
			gaps = append(gaps, fmt.Sprintf(
				"tool %q: declared via BuildTool(...) in tools.go but no registerHandler/registerQueryHandler "+
					"registration was found for it in internal/mcp/executor -- schema/dispatch wiring is missing entirely",
				tool))
		} else {
			unwired, err := unwiredParams(tool, toolProps[tool], entry, funcs)
			if err != nil {
				gaps = append(gaps, fmt.Sprintf("tool %q: %v", tool, err))
			}
			for _, p := range unwired {
				gaps = append(gaps, fmt.Sprintf(
					"tool %q: schema declares parameter %q but no reachable code from its registered handler "+
						"(under internal/mcp/executor, internal/mcp/pipeline, internal/mcp/response, internal/mcp/query, or internal/analysis) "+
						"reads it via a Get*Param/argument-style access -- wire it, remove it from this tool's schema, "+
						"or add a narrowly-scoped handlerScopedExemptions[tool][parameter] entry with justification",
					tool, p))
			}
		}
	}

	if len(gaps) > 0 {
		t.Errorf("DIR-045 schema/handler param-wiring gate found %d gap(s):\n  - %s",
			len(gaps), strings.Join(gaps, "\n  - "))
	}
}
