package response

import (
	"fmt"
	"sort"
	"strings"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
)

// DIR-080 / ADR-009: server-side field projection over flat and grouped
// query results. Paths are validated against the result shape DERIVED FROM
// the records themselves, so unknown or shape-incompatible paths are
// rejected with an actionable error instead of silently coercing to null —
// the ADR-009 requirement for actionable errors over silent coercion.

// maxKnownPathsInError bounds how many known paths an actionable projection
// error enumerates.
const maxKnownPathsInError = 24

// ProjectRecords returns one object per input record keyed by the requested
// jq-style paths (".timestamp", ".turns[].timestamp"). Paths containing "[]"
// collect matching elements into arrays; missing optional paths yield null.
// Every path is validated against the records' derived shape first.
func ProjectRecords(records []interface{}, paths []string) ([]interface{}, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("%w: projection requires at least one projection path (e.g. [\".timestamp\", \".message.content\"])", mcerrors.ErrInvalidInput)
	}
	if len(records) == 0 {
		return []interface{}{}, nil
	}

	shape := DeriveResultShape(records)
	nodes := indexShapeNodes(shape)

	parsed := make([][]pathSegment, len(paths))
	for i, p := range paths {
		if err := validateProjectionPath(p, nodes, shape); err != nil {
			return nil, err
		}
		parsed[i] = parsePath(p)
	}

	out := make([]interface{}, 0, len(records))
	for _, record := range records {
		projected := make(map[string]interface{}, len(paths))
		for i, p := range paths {
			projected[p] = extractPath(record, parsed[i])
		}
		out = append(out, projected)
	}
	return out, nil
}

// pathSegment is one step of a projection path.
type pathSegment struct {
	key     string
	iterate bool // trailing "[]": collect array elements
}

func parsePath(p string) []pathSegment {
	trimmed := strings.TrimPrefix(p, ".")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, ".")
	segments := make([]pathSegment, 0, len(parts))
	for _, part := range parts {
		seg := pathSegment{key: part}
		if strings.HasSuffix(part, "[]") || strings.HasSuffix(part, "[]?") {
			seg.key = strings.TrimSuffix(strings.TrimSuffix(part, "[]?"), "[]")
			seg.iterate = true
		}
		segments = append(segments, seg)
	}
	return segments
}

func extractPath(record interface{}, segments []pathSegment) interface{} {
	if len(segments) == 0 {
		return record
	}
	values := []interface{}{record}
	collecting := false
	for _, seg := range segments {
		var next []interface{}
		for _, v := range values {
			obj, ok := v.(map[string]interface{})
			if !ok {
				if !seg.iterate {
					next = append(next, nil)
				}
				continue
			}
			field := obj[seg.key]
			if seg.iterate {
				collecting = true
				if arr, ok := field.([]interface{}); ok {
					next = append(next, arr...)
				}
				continue
			}
			next = append(next, field)
		}
		values = next
	}
	if collecting {
		if values == nil {
			return []interface{}{}
		}
		return values
	}
	if len(values) == 1 {
		return values[0]
	}
	return nil
}

func indexShapeNodes(shape *ResultShape) map[string]*ShapeNode {
	nodes := map[string]*ShapeNode{}
	if shape == nil || shape.Root == nil {
		return nodes
	}
	var walk func(n *ShapeNode)
	walk = func(n *ShapeNode) {
		nodes[n.Path] = n
		for _, child := range n.Properties {
			walk(child)
		}
		if n.Elements != nil {
			walk(n.Elements)
		}
		for _, variant := range n.Variants {
			walk(variant)
		}
	}
	walk(shape.Root)
	return nodes
}

func validateProjectionPath(p string, nodes map[string]*ShapeNode, shape *ResultShape) error {
	if _, ok := nodes[p]; ok {
		return nil
	}

	// Find the longest addressable prefix to distinguish "unknown path" from
	// "attempt to descend into a scalar" (shape-incompatible).
	prefix := p
	for {
		idx := strings.LastIndex(prefix, ".")
		if idx <= 0 {
			prefix = ""
			break
		}
		prefix = prefix[:idx]
		if _, ok := nodes[prefix]; ok {
			break
		}
	}

	if prefix != "" {
		if node := nodes[prefix]; isScalarShapeType(node.Type) {
			return fmt.Errorf(
				"%w: projection path %q is not addressable: %q is a %s — cannot descend into a scalar value; see known paths below",
				mcerrors.ErrInvalidInput, p, prefix, node.Type)
		}
	}

	return fmt.Errorf("%w: unknown projection path %q; known paths: %s",
		mcerrors.ErrInvalidInput, p, knownPathsSummary(shape))
}

func isScalarShapeType(t string) bool {
	switch t {
	case ShapeTypeString, ShapeTypeNumber, ShapeTypeBoolean, ShapeTypeNull:
		return true
	}
	return false
}

func knownPathsSummary(shape *ResultShape) string {
	paths := shape.FlattenPaths()
	if len(paths) > maxKnownPathsInError {
		return strings.Join(paths[:maxKnownPathsInError], ", ") + fmt.Sprintf(", … (%d more)", len(paths)-maxKnownPathsInError)
	}
	sort.Strings(paths)
	return strings.Join(paths, ", ")
}
