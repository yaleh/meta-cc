package executor

// GetStringParam extracts a string parameter from args map.
func GetStringParam(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}

// GetBoolParam extracts a bool parameter from args map.
func GetBoolParam(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return defaultVal
}

// GetIntParam extracts an int parameter from args map.
func GetIntParam(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	if v, ok := args[key].(int); ok {
		return v
	}
	return defaultVal
}

// GetFloatParam extracts a float64 parameter from args map.
func GetFloatParam(args map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := args[key].(float64); ok {
		return v
	}
	return defaultVal
}

// GetBoolPtrParam extracts an optional bool parameter, returning nil when
// the key is absent so callers can distinguish "not set" (no constraint)
// from "explicitly false" — GetBoolParam collapses both to its default.
func GetBoolPtrParam(args map[string]interface{}, key string) *bool {
	v, ok := args[key].(bool)
	if !ok {
		return nil
	}
	return &v
}

// GetStringSliceParam extracts a []string parameter from args map (JSON
// arrays decode as []interface{}; non-string elements are skipped rather
// than erroring, matching this file's other Get*Param helpers' "best
// effort, defaultVal on shape mismatch" convention).
func GetStringSliceParam(args map[string]interface{}, key string) []string {
	raw, ok := args[key].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}
