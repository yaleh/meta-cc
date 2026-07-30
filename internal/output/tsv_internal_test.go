package output

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/types"
)

type tsvFixture struct {
	Name     string            `json:"name"`
	Ignored  string            `json:"-"`
	Count    int               `json:"count,omitempty"`
	Unsigned uint              `json:"unsigned"`
	Ratio    float64           `json:"ratio"`
	Active   bool              `json:"active"`
	Labels   []string          `json:"labels"`
	Metadata map[string]string `json:"metadata"`
	Child    struct{ ID int }  `json:"child"`
	Optional *string           `json:"optional"`
}

func TestFormatTSVDispatchesToolCallsAndGenericValues(t *testing.T) {
	got, err := FormatTSV([]types.ToolCall{{UUID: "id", ToolName: "Read", Status: "error", Error: "a\tb\nc"}})
	require.NoError(t, err)
	assert.Equal(t, "UUID\tToolName\tStatus\tError\nid\tRead\terror\ta\\tb\\nc\n", got)

	fixture := tsvFixture{Name: "a\tb", Count: -2, Unsigned: 3, Ratio: 1.25, Active: true, Labels: []string{"x"}, Metadata: map[string]string{"k": "v"}}
	got, err = FormatTSV(fixture)
	require.NoError(t, err)
	assert.Contains(t, got, "name\ta\\tb\n")
	assert.Contains(t, got, "labels\t[\"x\"]\n")
}

func TestFormatGenericTSVHandlesNilEmptyAndUnsupported(t *testing.T) {
	got, err := FormatGenericTSV(nil)
	require.NoError(t, err)
	assert.Empty(t, got)

	var fixture *tsvFixture
	got, err = FormatGenericTSV(fixture)
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = FormatGenericTSV([]tsvFixture{})
	require.NoError(t, err)
	assert.Empty(t, got)

	_, err = FormatGenericTSV(42)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported data type")
}

func TestReflectionHelpersHandlePointersAndNonStructs(t *testing.T) {
	var nilFixture *tsvFixture
	assert.Empty(t, getStructFields(reflect.ValueOf(nilFixture)))
	assert.Empty(t, getStructValues(reflect.ValueOf(nilFixture)))
	assert.Empty(t, getStructFields(reflect.ValueOf(1)))
	assert.Empty(t, getStructValues(reflect.ValueOf(1)))

	fixture := &tsvFixture{Name: "name"}
	fields := getStructFields(reflect.ValueOf(fixture))
	assert.Contains(t, fields, "name")
	assert.Contains(t, fields, "Ignored")
	values := getStructValues(reflect.ValueOf(fixture))
	assert.Len(t, values, len(fields))
}

func TestFormatTSVValueCoversScalarComplexAndPointers(t *testing.T) {
	text := "line\nvalue"
	assert.Equal(t, "line\\nvalue", formatTSVValue(reflect.ValueOf(&text)))
	var nilText *string
	assert.Empty(t, formatTSVValue(reflect.ValueOf(nilText)))
	assert.Equal(t, "-2", formatTSVValue(reflect.ValueOf(int8(-2))))
	assert.Equal(t, "3", formatTSVValue(reflect.ValueOf(uint16(3))))
	assert.Equal(t, "1.50", formatTSVValue(reflect.ValueOf(float32(1.5))))
	assert.Equal(t, "true", formatTSVValue(reflect.ValueOf(true)))
	assert.Equal(t, "[1,2]", formatTSVValue(reflect.ValueOf([]int{1, 2})))
	assert.Equal(t, "(1+2i)", formatTSVValue(reflect.ValueOf(complex(1, 2))))
}

func TestFormatProjectedTSVUsesFirstRecordFields(t *testing.T) {
	assert.Empty(t, FormatProjectedTSV(nil))
	got := FormatProjectedTSV([]ProjectedToolCall{{"b": "x\ny", "a": 1}, {"a": 2, "b": "z"}})
	assert.Equal(t, "a\tb\n1\tx\\ny\n2\tz\n", got)
}
