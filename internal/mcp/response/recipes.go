package response

import (
	"fmt"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
)

// DIR-080 / ADR-009: jq recipes GENERATED FROM the actual emitted result
// shape and validated against the same records before emission, so file_ref
// consumers receive query programs that are proven to run against the
// result — never hand-written expressions that can drift from the data.

const (
	// maxRecipes bounds how many recipes are attached to one file_ref.
	maxRecipes = 16
	// maxRecipeValidationErrors bounds error detail in ValidateRecipes.
	maxRecipeValidationErrors = 3
)

// Recipe is one generated, validated jq program. Scope "record" means the
// program is applied to one result record at a time (e.g. a JSONL line, or
// `.[]` iteration over an inline data array).
type Recipe struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	JQ          string `json:"jq"`
	Scope       string `json:"scope"`
}

type recipeCandidate struct {
	jq       string
	nodeType string
}

// GenerateRecipes derives bounded jq recipes from a result shape and
// self-validates each recipe against the records the shape was derived from;
// any recipe that fails validation is dropped rather than emitted. Returns
// nil for a nil shape or non-object roots with no object variant.
func GenerateRecipes(shape *ResultShape, records []interface{}) []Recipe {
	if shape == nil || shape.Root == nil {
		return nil
	}

	var candidates []recipeCandidate
	seen := map[string]bool{}
	emit := func(jq, nodeType string) {
		if seen[jq] {
			return
		}
		seen[jq] = true
		candidates = append(candidates, recipeCandidate{jq: jq, nodeType: nodeType})
	}

	root := shape.Root
	switch root.Type {
	case ShapeTypeObject:
		collectRecipeLeaves(root, "", false, emit)
	case ShapeTypeMixed:
		for _, variant := range root.Variants {
			if variant.Type == ShapeTypeObject {
				collectVariantRecipeLeaves(variant, "", true, emit)
			}
		}
	default:
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		if len(candidates[i].jq) != len(candidates[j].jq) {
			return len(candidates[i].jq) < len(candidates[j].jq)
		}
		return candidates[i].jq < candidates[j].jq
	})
	if len(candidates) > maxRecipes {
		candidates = candidates[:maxRecipes]
	}

	recipes := make([]Recipe, 0, len(candidates))
	ids := map[string]int{}
	for _, c := range candidates {
		r := Recipe{
			ID:          uniqueRecipeID(c.jq, ids),
			Description: fmt.Sprintf("Extract %s (%s value) — generated from the emitted result shape, validated server-side.", c.jq, c.nodeType),
			JQ:          c.jq,
			Scope:       "record",
		}
		if ValidateRecipes([]Recipe{r}, records) != nil {
			continue // derived expressions should pass; drop any that don't
		}
		recipes = append(recipes, r)
	}
	return recipes
}

// collectRecipeLeaves walks an object/array shape accumulating leaf-path jq
// expressions. throughMixed marks paths crossing a heterogeneous node; those
// get a trailing try operator so they execute across all record variants.
func collectRecipeLeaves(node *ShapeNode, expr string, throughMixed bool, emit func(string, string)) {
	if len(node.Properties) > 0 {
		for _, k := range sortedKeys(node.Properties) {
			childExpr := expr + "." + k
			collectRecipeLeaves(node.Properties[k], childExpr, throughMixed, emit)
		}
	}
	if node.Elements != nil {
		suffix := "[]"
		if node.Nullable {
			suffix = "[]?"
		}
		collectRecipeLeaves(node.Elements, expr+suffix, throughMixed, emit)
	}
	for _, variant := range node.Variants {
		collectVariantRecipeLeaves(variant, expr, true, emit)
	}
	if len(node.Properties) == 0 && node.Elements == nil && len(node.Variants) == 0 && expr != "" {
		emit(tryExpr(expr, throughMixed), node.Type)
	}
}

// collectVariantRecipeLeaves handles one concrete shape of a heterogeneous
// path: object variants extend the expression with field access, array
// variants use safe iteration (records where this path holds the OTHER
// variant must not make the recipe fail), scalar variants terminate.
func collectVariantRecipeLeaves(variant *ShapeNode, expr string, throughMixed bool, emit func(string, string)) {
	switch variant.Type {
	case ShapeTypeObject:
		if len(variant.Properties) == 0 {
			emit(tryExpr(expr, throughMixed), variant.Type)
			return
		}
		for _, k := range sortedKeys(variant.Properties) {
			collectRecipeLeaves(variant.Properties[k], expr+"."+k, throughMixed, emit)
		}
	case ShapeTypeArray:
		if variant.Elements == nil {
			emit(tryExpr(expr+"[]?", throughMixed), variant.Type)
			return
		}
		collectRecipeLeaves(variant.Elements, expr+"[]?", throughMixed, emit)
	default:
		emit(tryExpr(expr, throughMixed), variant.Type)
	}
}

func tryExpr(expr string, throughMixed bool) string {
	if throughMixed && !strings.HasSuffix(expr, "?") {
		return expr + "?"
	}
	return expr
}

func uniqueRecipeID(jq string, counts map[string]int) string {
	replacer := strings.NewReplacer("[]?", "_", "[]", "_", ".", "_", "?", "")
	base := strings.Trim(replacer.Replace(jq), "_")
	if base == "" {
		base = "root"
	}
	id := base
	if n := counts[base]; n > 0 {
		id = fmt.Sprintf("%s_%d", base, n+1)
	}
	counts[base]++
	return id
}

// ValidateRecipes executes every recipe against every record with gojq and
// returns an actionable error naming the failing recipe and record index.
func ValidateRecipes(recipes []Recipe, records []interface{}) error {
	var failures []string
	for _, recipe := range recipes {
		query, err := gojq.Parse(recipe.JQ)
		if err != nil {
			failures = append(failures, fmt.Sprintf("recipe %q: invalid jq %q: %v", recipe.ID, recipe.JQ, err))
			continue
		}
		for i, record := range records {
			// gojq normalizes numeric values in place before execution. File-ref
			// responses may be built concurrently from the same result slice, so
			// never let recipe validation mutate caller-owned records.
			iter := query.Run(cloneJQInput(record))
			for {
				value, ok := iter.Next()
				if !ok {
					break
				}
				if runErr, isErr := value.(error); isErr {
					failures = append(failures, fmt.Sprintf("recipe %q (jq: %s) failed on record %d: %v", recipe.ID, recipe.JQ, i, runErr))
					break
				}
			}
		}
		if len(failures) >= maxRecipeValidationErrors {
			break
		}
	}
	if len(failures) == 0 {
		return nil
	}
	return fmt.Errorf("recipe validation failed: %s", strings.Join(failures, "; "))
}

func cloneJQInput(value interface{}) interface{} {
	switch value := value.(type) {
	case map[string]interface{}:
		cloned := make(map[string]interface{}, len(value))
		for key, child := range value {
			cloned[key] = cloneJQInput(child)
		}
		return cloned
	case []interface{}:
		cloned := make([]interface{}, len(value))
		for i, child := range value {
			cloned[i] = cloneJQInput(child)
		}
		return cloned
	default:
		return value
	}
}
