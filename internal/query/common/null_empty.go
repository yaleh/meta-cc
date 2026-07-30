// Package common provides helpers shared by query implementations.
package common

// AllNullOrEmpty reports whether every entry is nil or a map containing only
// nil and empty-string values. It reports false for empty slices and non-map scalars.
func AllNullOrEmpty(entries []interface{}) bool {
	if len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		m, ok := entry.(map[string]interface{})
		if !ok {
			return false
		}
		for _, value := range m {
			if value == nil {
				continue
			}
			if text, ok := value.(string); ok && text == "" {
				continue
			}
			return false
		}
	}
	return true
}
