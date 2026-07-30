package response

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// DIR-080 / ADR-009: versioned, typed nested result-shape description for
// file_ref results. The shape is DERIVED FROM the actual emitted records (not
// maintained by hand) so metadata cannot drift from the data it describes —
// the ADR-009 risk mitigation for "result metadata could drift from the
// emitted data".

const (
	// ShapeSchemaVersion versions the shape vocabulary so consumers can
	// detect contract changes.
	ShapeSchemaVersion = 1

	// shapeScanLimit bounds how many records are inspected when deriving a
	// shape, so a huge result set cannot make metadata generation quadratic.
	shapeScanLimit = 64
	// shapeMaxDepth bounds recursion for pathologically nested records.
	shapeMaxDepth = 8
	// shapeMaxValues bounds how many distinct scalar values are recorded per
	// node (enough for provider/role provenance, not enough to leak content).
	shapeMaxValues = 8
	// shapeMaxValueLen bounds each recorded scalar value.
	shapeMaxValueLen = 32
)

// Shape node type constants (JSON type vocabulary plus "mixed" for
// heterogeneous paths and "unknown" for depth-capped subtrees).
const (
	ShapeTypeObject  = "object"
	ShapeTypeArray   = "array"
	ShapeTypeString  = "string"
	ShapeTypeNumber  = "number"
	ShapeTypeBoolean = "boolean"
	ShapeTypeNull    = "null"
	ShapeTypeMixed   = "mixed"
)

// ResultShape is the top-level, versioned shape description attached to
// file_ref metadata.
type ResultShape struct {
	SchemaVersion  int        `json:"schema_version"`
	Root           *ShapeNode `json:"root"`
	RecordsScanned int        `json:"records_scanned"`
}

// ShapeNode describes one addressable path in the result. Paths use jq
// notation: object fields as ".field", array elements with a trailing "[]"
// (e.g. ".message.content[].text").
type ShapeNode struct {
	// Type is object|array|string|number|boolean|null|mixed. "mixed" means
	// different records produce structurally different shapes at this path
	// (e.g. message.content is a string on user records and an array on
	// assistant records); the concrete shapes are listed in Variants.
	Type string `json:"type"`
	// Path is the jq-style absolute path of this node ("." for the root).
	Path string `json:"path"`
	// Optional marks object fields that are absent on at least one record
	// where the parent object was present.
	Optional bool `json:"optional,omitempty"`
	// Nullable marks paths where an explicit JSON null was observed.
	Nullable bool `json:"nullable,omitempty"`
	// Observed counts how many parent instances exhibited this node.
	Observed int `json:"observed,omitempty"`
	// Values lists bounded distinct scalar values observed at this path
	// (enumeration/provenance signal, e.g. provider: ["claude","codex"]).
	Values []string `json:"values,omitempty"`
	// Properties describes object fields (Type == object).
	Properties map[string]*ShapeNode `json:"properties,omitempty"`
	// Elements describes array element shape (Type == array).
	Elements *ShapeNode `json:"elements,omitempty"`
	// Variants lists the concrete shapes for heterogeneous paths
	// (Type == mixed).
	Variants []*ShapeNode `json:"variants,omitempty"`
}

// shapeBuilder accumulates observations for one path during derivation.
type shapeBuilder struct {
	path      string
	count     int // observations of this path under its parent
	nullCount int
	types     map[string]*shapeBuilder // per concrete type (object/array/string/number/boolean)
	values    map[string]bool
	objSeen   int
	elemSeen  int
}

func newShapeBuilder(path string) *shapeBuilder {
	return &shapeBuilder{
		path:   path,
		types:  make(map[string]*shapeBuilder),
		values: make(map[string]bool),
	}
}

func (b *shapeBuilder) observe(v interface{}, depth int) {
	b.count++
	switch val := v.(type) {
	case nil:
		b.nullCount++
	case map[string]interface{}:
		b.objSeen++
		tb := b.types[ShapeTypeObject]
		if tb == nil {
			tb = newShapeBuilder(b.path)
			b.types[ShapeTypeObject] = tb
		}
		tb.count++
		for k, child := range val {
			cb := tb.types[k] // reuse per-field builders stored on the object builder
			if cb == nil {
				childPath := b.path
				if childPath == "." {
					childPath = "." + k
				} else {
					childPath = childPath + "." + k
				}
				cb = newShapeBuilder(childPath)
				tb.types[k] = cb
			}
			if depth >= shapeMaxDepth {
				continue
			}
			cb.observe(child, depth+1)
		}
	case []interface{}:
		b.elemSeen++
		eb := b.types[ShapeTypeArray]
		if eb == nil {
			eb = newShapeBuilder(b.path + "[]")
			b.types[ShapeTypeArray] = eb
		}
		if depth < shapeMaxDepth {
			for _, elem := range val {
				eb.observe(elem, depth+1)
			}
		}
	case string:
		b.addScalar(ShapeTypeString, val)
	case bool:
		b.addScalar(ShapeTypeBoolean, strconv.FormatBool(val))
	case float64:
		b.addScalar(ShapeTypeNumber, strconv.FormatFloat(val, 'f', -1, 64))
	case int:
		b.addScalar(ShapeTypeNumber, strconv.Itoa(val))
	case int64:
		b.addScalar(ShapeTypeNumber, strconv.FormatInt(val, 10))
	}
}

func (b *shapeBuilder) addScalar(typeName, value string) {
	if _, ok := b.types[typeName]; !ok {
		b.types[typeName] = newShapeBuilder(b.path)
	}
	if len(b.values) < shapeMaxValues && len(value) <= shapeMaxValueLen {
		b.values[value] = true
	}
}

// DeriveResultShape inspects up to shapeScanLimit records and returns a
// versioned shape description of the result set, or nil for empty input.
// Non-object root records are supported (scalar roots surface as their scalar
// type, or "mixed" when records disagree).
func DeriveResultShape(records []interface{}) *ResultShape {
	if len(records) == 0 {
		return nil
	}

	scanned := records
	if len(scanned) > shapeScanLimit {
		scanned = scanned[:shapeScanLimit]
	}

	root := newShapeBuilder(".")
	for _, record := range scanned {
		root.observe(record, 0)
	}

	return &ResultShape{
		SchemaVersion:  ShapeSchemaVersion,
		Root:           finalize(root, len(scanned)),
		RecordsScanned: len(scanned),
	}
}

// finalize converts an observation builder into an immutable ShapeNode.
// parentCount is how many times the parent was observed; it drives Optional.
func finalize(b *shapeBuilder, parentCount int) *ShapeNode {
	node := &ShapeNode{
		Path:     b.path,
		Observed: b.count,
		Nullable: b.nullCount > 0,
		Optional: parentCount > 0 && b.count < parentCount,
	}

	// Distinct scalar values (bounded) for provenance/enumeration.
	if len(b.values) > 0 {
		values := make([]string, 0, len(b.values))
		for v := range b.values {
			values = append(values, v)
		}
		sort.Strings(values)
		node.Values = values
	}

	objB := b.types[ShapeTypeObject]
	arrB := b.types[ShapeTypeArray]
	concreteTypes := make([]string, 0, 5)
	for _, t := range []string{ShapeTypeObject, ShapeTypeArray, ShapeTypeString, ShapeTypeNumber, ShapeTypeBoolean} {
		if _, ok := b.types[t]; ok {
			concreteTypes = append(concreteTypes, t)
		}
	}

	switch len(concreteTypes) {
	case 0:
		node.Type = ShapeTypeNull
	case 1:
		node.Type = concreteTypes[0]
		switch node.Type {
		case ShapeTypeObject:
			node.Properties = finalizeObjectProps(objB)
		case ShapeTypeArray:
			// Element optionality is measured against element count, not
			// array count, so empty arrays don't mark elements optional.
			node.Elements = finalize(arrB, arrB.count)
		}
	default:
		node.Type = ShapeTypeMixed
		for _, t := range concreteTypes {
			node.Variants = append(node.Variants, finalizeVariant(b, b.types[t], t))
		}
	}
	return node
}

// finalizeObjectProps finalizes every field builder stored on an object
// builder. Field optionality is measured against how many object instances
// were observed (objB.count).
func finalizeObjectProps(objB *shapeBuilder) map[string]*ShapeNode {
	props := make(map[string]*ShapeNode, len(objB.types))
	keys := make([]string, 0, len(objB.types))
	for k := range objB.types {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		props[k] = finalize(objB.types[k], objB.count)
	}
	return props
}

// finalizeVariant materializes one concrete shape of a heterogeneous path.
// parent is the path-level builder (supplies the shared path and array
// observation counts); typeBuilder is the per-type builder (its .types holds
// field builders for objects, or the element observations for arrays).
func finalizeVariant(parent *shapeBuilder, typeBuilder *shapeBuilder, typeName string) *ShapeNode {
	node := &ShapeNode{
		Type:     typeName,
		Path:     parent.path,
		Observed: typeBuilder.count,
	}
	switch typeName {
	case ShapeTypeObject:
		node.Observed = typeBuilder.count
		node.Properties = finalizeObjectProps(typeBuilder)
	case ShapeTypeArray:
		// The array-type builder accumulates element observations directly;
		// finalize it as the element node (path already ends in "[]").
		node.Observed = parent.elemSeen
		node.Elements = finalize(typeBuilder, typeBuilder.count)
	}
	return node
}

// FlattenPaths returns every addressable jq-style path in the shape, sorted
// (root "." excluded). Used for projection validation and recipe generation.
func (s *ResultShape) FlattenPaths() []string {
	if s == nil || s.Root == nil {
		return nil
	}
	var paths []string
	flattenNode(s.Root, &paths)
	return dedupeSorted(paths)
}

// dedupeSorted sorts and removes duplicates (variant traversal can emit the
// same path more than once).
func dedupeSorted(paths []string) []string {
	sort.Strings(paths)
	out := paths[:0]
	var prev string
	for i, p := range paths {
		if i > 0 && p == prev {
			continue
		}
		out = append(out, p)
		prev = p
	}
	return out
}

func flattenNode(node *ShapeNode, paths *[]string) {
	if node.Path != "." {
		*paths = append(*paths, node.Path)
	}
	keys := make([]string, 0, len(node.Properties))
	for k := range node.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		flattenNode(node.Properties[k], paths)
	}
	if node.Elements != nil {
		flattenNode(node.Elements, paths)
	}
	for _, variant := range node.Variants {
		flattenNode(variant, paths)
	}
}

// LeafPaths returns paths whose node carries no object properties and no
// array elements (terminal values), sorted. Array paths themselves count as
// leaves when their element shape is scalar or mixed.
func (s *ResultShape) LeafPaths() []string {
	if s == nil || s.Root == nil {
		return nil
	}
	var leaves []string
	collectLeaves(s.Root, &leaves)
	return dedupeSorted(leaves)
}

func collectLeaves(node *ShapeNode, leaves *[]string) {
	hasChildren := len(node.Properties) > 0 || node.Elements != nil
	if !hasChildren && node.Path != "." {
		*leaves = append(*leaves, node.Path)
		return
	}
	for _, k := range sortedKeys(node.Properties) {
		collectLeaves(node.Properties[k], leaves)
	}
	if node.Elements != nil {
		collectLeaves(node.Elements, leaves)
	}
	for _, variant := range node.Variants {
		collectLeaves(variant, leaves)
	}
}

func sortedKeys(m map[string]*ShapeNode) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// String renders a compact human-readable path listing (for actionable
// errors).
func (s *ResultShape) String() string {
	return fmt.Sprintf("ResultShape(v%d, paths=%s)", s.SchemaVersion, strings.Join(s.FlattenPaths(), ", "))
}
