package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveResultShape_NestedPathsAndArrays(t *testing.T) {
	shape := DeriveResultShape(claudeFixtureRecords())
	require.NotNil(t, shape)
	assert.Equal(t, ShapeSchemaVersion, shape.SchemaVersion)
	assert.Equal(t, "object", shape.Root.Type)

	msg := shape.Root.Properties["message"]
	require.NotNil(t, msg, "nested .message must be described")
	assert.Equal(t, "object", msg.Type)
	assert.Equal(t, ".message", msg.Path)

	content := msg.Properties["content"]
	require.NotNil(t, content)
	assert.Equal(t, "array", content.Type)
	assert.Equal(t, ".message.content", content.Path)
	require.NotNil(t, content.Elements, "array element shape must be described")
	assert.Equal(t, ".message.content[]", content.Elements.Path)
	assert.Equal(t, "object", content.Elements.Type)
	assert.Contains(t, content.Elements.Properties, "type")

	usage := msg.Properties["usage"]
	require.NotNil(t, usage)
	assert.True(t, usage.Optional, "usage is absent on one record -> optional")
	assert.Contains(t, usage.Properties, "input_tokens")
}

func TestDeriveResultShape_NullableAndOptional(t *testing.T) {
	shape := DeriveResultShape(heterogeneousFixtureRecords())
	require.NotNil(t, shape)

	msg := shape.Root.Properties["message"]
	require.NotNil(t, msg)
	assert.True(t, msg.Nullable, "message is null on one record -> nullable")
	assert.True(t, msg.Optional || msg.Nullable, "absence must be surfaced")
}

func TestDeriveResultShape_HeterogeneousVariants(t *testing.T) {
	shape := DeriveResultShape(heterogeneousFixtureRecords())
	require.NotNil(t, shape)

	content := shape.Root.Properties["message"].Properties["content"]
	require.NotNil(t, content)
	assert.Equal(t, "mixed", content.Type, "string vs array content -> mixed")
	types := map[string]bool{}
	for _, v := range content.Variants {
		types[v.Type] = true
	}
	assert.True(t, types["string"], "string variant must be declared")
	assert.True(t, types["array"], "array variant must be declared")
}

func TestDeriveResultShape_ProviderProvenance(t *testing.T) {
	shape := DeriveResultShape(heterogeneousFixtureRecords())
	require.NotNil(t, shape)

	provider := shape.Root.Properties["provider"]
	require.NotNil(t, provider)
	assert.ElementsMatch(t, []string{"claude", "codex"}, provider.Values,
		"observed provider values must be enumerated for provenance")
}

func TestDeriveResultShape_EmptyAndNonObjectRecords(t *testing.T) {
	assert.Nil(t, DeriveResultShape(nil))

	shape := DeriveResultShape([]interface{}{"scalar", 42})
	require.NotNil(t, shape)
	assert.Equal(t, "mixed", shape.Root.Type)
}

func TestDeriveResultShape_DepthAndScanBounds(t *testing.T) {
	// Deeply nested chain beyond the derivation depth cap must terminate
	// with a bounded node instead of recursing forever.
	deep := map[string]interface{}{}
	cur := deep
	for i := 0; i < shapeMaxDepth+4; i++ {
		next := map[string]interface{}{}
		cur["level"] = next
		cur = next
	}
	cur["leaf"] = "bottom"

	shape := DeriveResultShape([]interface{}{deep})
	require.NotNil(t, shape)

	node := shape.Root
	depth := 0
	for node != nil && node.Properties["level"] != nil {
		node = node.Properties["level"]
		depth++
		require.LessOrEqual(t, depth, shapeMaxDepth+1, "depth must be capped")
	}
}

func TestFlattenShapePaths_CoversNestedArrayPaths(t *testing.T) {
	shape := DeriveResultShape(claudeFixtureRecords())
	paths := shape.FlattenPaths()
	assert.Contains(t, paths, ".message")
	assert.Contains(t, paths, ".message.content")
	assert.Contains(t, paths, ".message.content[].type")
	assert.Contains(t, paths, ".provider")
}
