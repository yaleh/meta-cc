package response

import (
	"fmt"
	"testing"

	"github.com/itchyny/gojq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRecipes_ExecuteAgainstAllFixtures(t *testing.T) {
	fixtures := map[string][]interface{}{
		"claude":        claudeFixtureRecords(),
		"codex":         codexFixtureRecords(),
		"grouped":       groupedFixtureRecords(),
		"heterogeneous": heterogeneousFixtureRecords(),
	}
	for name, records := range fixtures {
		t.Run(name, func(t *testing.T) {
			shape := DeriveResultShape(records)
			recipes := GenerateRecipes(shape, records)
			require.NotEmpty(t, recipes, "every fixture must yield recipes")

			for _, recipe := range recipes {
				_, err := gojq.Parse(recipe.JQ)
				require.NoError(t, err, "recipe %q must be syntactically valid jq", recipe.ID)
				assert.Equal(t, "record", recipe.Scope)
				assert.NotEmpty(t, recipe.Description)
			}
			require.NoError(t, ValidateRecipes(recipes, records),
				"every generated recipe must execute against its own fixture")
		})
	}
}

func TestGenerateRecipes_ClaudeNestedArrayPaths(t *testing.T) {
	records := claudeFixtureRecords()
	recipes := GenerateRecipes(DeriveResultShape(records), records)

	jqs := map[string]bool{}
	for _, r := range recipes {
		jqs[r.JQ] = true
	}
	assert.True(t, jqs[".message.content[].type"], "pure-array iteration needs no try operator")
	assert.True(t, jqs[".message.usage.input_tokens"], "optional object leaves are addressable")
}

func TestGenerateRecipes_GroupedTurns(t *testing.T) {
	records := groupedFixtureRecords()
	recipes := GenerateRecipes(DeriveResultShape(records), records)

	jqs := map[string]bool{}
	for _, r := range recipes {
		jqs[r.JQ] = true
	}
	assert.True(t, jqs[".turns[].timestamp"], "grouped fixture exposes nested turns paths")
	assert.True(t, jqs[".session_id"])
}

func TestGenerateRecipes_HeterogeneousUsesSafeIteration(t *testing.T) {
	records := heterogeneousFixtureRecords()
	recipes := GenerateRecipes(DeriveResultShape(records), records)

	sawSafe := false
	for _, r := range recipes {
		assert.NotContains(t, r.JQ, ".message.content[].", "unsafe iteration over mixed string/array content")
		if r.JQ == ".message.content[]?" || r.JQ == ".message.content[]?.type?" {
			sawSafe = true
		}
	}
	assert.True(t, sawSafe, "mixed content must produce []? safe-iteration recipes")
	require.NoError(t, ValidateRecipes(recipes, records))
}

func TestGenerateRecipes_Bounded(t *testing.T) {
	wide := map[string]interface{}{}
	for i := 0; i < 40; i++ {
		wide[fmt.Sprintf("field_%02d", i)] = float64(i)
	}
	recipes := GenerateRecipes(DeriveResultShape([]interface{}{wide}), []interface{}{wide})
	assert.LessOrEqual(t, len(recipes), maxRecipes)
	assert.NotEmpty(t, recipes)
}

func TestGenerateRecipes_Deterministic(t *testing.T) {
	records := claudeFixtureRecords()
	first := GenerateRecipes(DeriveResultShape(records), records)
	second := GenerateRecipes(DeriveResultShape(records), records)
	require.Equal(t, len(first), len(second))
	for i := range first {
		assert.Equal(t, first[i], second[i])
	}
}

func TestGenerateRecipes_NilShape(t *testing.T) {
	assert.Empty(t, GenerateRecipes(nil, nil))
}

func TestValidateRecipes_ActionableError(t *testing.T) {
	// Unsafe iteration over mixed string/array content MUST fail validation
	// with the recipe and record index named.
	recipes := []Recipe{{ID: "unsafe", JQ: ".message.content[].text", Scope: "record"}}
	err := ValidateRecipes(recipes, heterogeneousFixtureRecords())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe")
	assert.Contains(t, err.Error(), ".message.content[].text")
	assert.Contains(t, err.Error(), "record")
}
