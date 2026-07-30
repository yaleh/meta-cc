package response

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBoundedSample_CapsRecordCount(t *testing.T) {
	records := make([]interface{}, 20)
	for i := range records {
		records[i] = map[string]interface{}{"id": float64(i)}
	}
	sample := BuildBoundedSample(records)
	assert.LessOrEqual(t, len(sample), maxSampleRecords, "sample must be deterministically bounded")
	assert.NotEmpty(t, sample)
}

func TestBuildBoundedSample_RedactsSensitiveFields(t *testing.T) {
	sample := BuildBoundedSample(sensitiveFixtureRecords())
	require.Len(t, sample, 1)

	serialized, err := json.Marshal(sample)
	require.NoError(t, err)
	s := string(serialized)

	for _, leaked := range []string{"sk-live-1234567890", "secret-payload", "tok_abcdef", "shhh", "hunter2"} {
		assert.NotContains(t, s, leaked, "secret-looking field value must be redacted")
	}
	assert.Contains(t, s, redactedMarker)
	assert.Contains(t, s, "short", "non-sensitive fields survive redaction")

	rec := sample[0].(map[string]interface{})
	assert.Equal(t, redactedMarker, rec["api_key"])
	assert.Equal(t, redactedMarker, rec["Authorization"])
	nested := rec["nested"].(map[string]interface{})
	assert.Equal(t, redactedMarker, nested["client_secret"])
}

func TestBuildBoundedSample_TruncatesOversizedStrings(t *testing.T) {
	sample := BuildBoundedSample(sensitiveFixtureRecords())
	require.Len(t, sample, 1)

	rec := sample[0].(map[string]interface{})
	text, ok := rec["text"].(string)
	require.True(t, ok)
	assert.LessOrEqual(t, len(text), maxSampleStringLen+64, "oversized string must be truncated")
	assert.Contains(t, text, "truncated", "truncation must be self-describing")
}

func TestBuildBoundedSample_CapsArrayElementsAndDepth(t *testing.T) {
	bigArray := make([]interface{}, 50)
	for i := range bigArray {
		bigArray[i] = float64(i)
	}
	deep := map[string]interface{}{}
	cur := deep
	for i := 0; i < maxSampleDepth+5; i++ {
		next := map[string]interface{}{}
		cur["child"] = next
		cur = next
	}
	cur["leaf"] = "x"

	sample := BuildBoundedSample([]interface{}{
		map[string]interface{}{"items": bigArray, "deep": deep},
	})
	require.Len(t, sample, 1)

	serialized, err := json.Marshal(sample)
	require.NoError(t, err)
	assert.Contains(t, string(serialized), "more", "capped array must note omitted elements")
	assert.LessOrEqual(t, len(serialized), maxSampleTotalBytes, "total sample must respect byte cap")
}

func TestBuildBoundedSample_TotalByteCap(t *testing.T) {
	wide := map[string]interface{}{}
	for i := 0; i < 40; i++ {
		wide["field_"+strings.Repeat("k", 10)+string(rune('a'+i%26))] = strings.Repeat("v", 200)
	}
	sample := BuildBoundedSample([]interface{}{wide})
	serialized, err := json.Marshal(sample)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(serialized), maxSampleTotalBytes,
		"final sample serialization must respect the byte cap")
}

func TestBuildBoundedSample_EmptyInput(t *testing.T) {
	assert.Empty(t, BuildBoundedSample(nil))
}
