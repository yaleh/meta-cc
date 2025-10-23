package query

// QueryParams represents unified query parameters
type QueryParams struct {
	// Tier 1: Resource Selection
	Resource string `json:"resource"` // "entries" | "messages" | "tools"

	// Tier 2: Scope
	Scope string `json:"scope"` // "session" | "project"

	// Tier 3: Filtering
	Filter FilterSpec `json:"filter"`

	// Tier 4: Transformation
	Transform TransformSpec `json:"transform"`

	// Tier 5: Aggregation
	Aggregate AggregateSpec `json:"aggregate"`

	// Tier 6: Output Control
	Output OutputSpec `json:"output"`

	// Advanced: jq filter (escape hatch for complex queries)
	JQFilter string `json:"jq_filter,omitempty"`
}

// FilterSpec represents structured filter conditions
type FilterSpec struct {
	// Entry-level filters
	Type       string     `json:"type,omitempty"`        // Entry type: "user", "assistant", etc.
	SessionID  string     `json:"session_id,omitempty"`  // Session ID filter
	UUID       string     `json:"uuid,omitempty"`        // Specific entry UUID
	ParentUUID string     `json:"parent_uuid,omitempty"` // Parent UUID for dialog chain
	GitBranch  string     `json:"git_branch,omitempty"`  // Git branch filter
	TimeRange  *TimeRange `json:"time_range,omitempty"`  // Time range filter

	// Message-level filters
	Role         string `json:"role,omitempty"`          // Message role: "user" | "assistant"
	ContentType  string `json:"content_type,omitempty"`  // Content block type filter
	ContentMatch string `json:"content_match,omitempty"` // Regex pattern for content matching

	// Tool-level filters
	ToolName   string `json:"tool_name,omitempty"`   // Tool name (supports regex)
	ToolStatus string `json:"tool_status,omitempty"` // Tool execution status
	HasError   *bool  `json:"has_error,omitempty"`   // Filter by error presence
}

// TimeRange represents a time range filter
type TimeRange struct {
	Start string `json:"start,omitempty"` // ISO8601 timestamp
	End   string `json:"end,omitempty"`   // ISO8601 timestamp
}

// TransformSpec represents transformation operations
type TransformSpec struct {
	Extract []string  `json:"extract,omitempty"`  // JSONPath expressions to extract fields
	GroupBy string    `json:"group_by,omitempty"` // Field to group by
	Join    *JoinSpec `json:"join,omitempty"`     // Join with related entries
}

// JoinSpec represents a join operation
type JoinSpec struct {
	Type string `json:"type"` // Entry type to join with
	On   string `json:"on"`   // Field to join on
}

// AggregateSpec represents aggregation operations
type AggregateSpec struct {
	Function string `json:"function,omitempty"` // "count" | "sum" | "avg" | "min" | "max" | "group"
	Field    string `json:"field,omitempty"`    // Field to aggregate on
}

// OutputSpec represents output control options
type OutputSpec struct {
	Format    string `json:"format,omitempty"`     // "jsonl" | "tsv" | "summary"
	Limit     int    `json:"limit,omitempty"`      // Maximum number of results
	SortBy    string `json:"sort_by,omitempty"`    // Field to sort by
	SortOrder string `json:"sort_order,omitempty"` // "asc" | "desc"
}

// IsEmpty returns true if the filter has no conditions
func (f FilterSpec) IsEmpty() bool {
	return f.Type == "" &&
		f.SessionID == "" &&
		f.UUID == "" &&
		f.ParentUUID == "" &&
		f.GitBranch == "" &&
		f.TimeRange == nil &&
		f.Role == "" &&
		f.ContentType == "" &&
		f.ContentMatch == "" &&
		f.ToolName == "" &&
		f.ToolStatus == "" &&
		f.HasError == nil
}

// IsEmpty returns true if the aggregate has no operations
func (a AggregateSpec) IsEmpty() bool {
	return a.Function == ""
}

// ValidResourceTypes lists valid resource types
var ValidResourceTypes = []string{"entries", "messages", "tools"}

// ValidScopes lists valid scope values
var ValidScopes = []string{"session", "project"}

// ValidAggregateFunctions lists valid aggregate functions
var ValidAggregateFunctions = []string{"count", "sum", "avg", "min", "max", "group"}

// ValidOutputFormats lists valid output formats
var ValidOutputFormats = []string{"jsonl", "tsv", "summary"}

// ValidateQueryParams validates query parameters and returns an error if invalid
func ValidateQueryParams(params QueryParams) error {
	// Apply defaults first
	params = ApplyDefaults(params)

	// Validate resource type
	if !isValidValue(params.Resource, ValidResourceTypes) {
		return &ValidationError{Field: "resource", Value: params.Resource, ValidValues: ValidResourceTypes}
	}

	// Validate scope
	if !isValidValue(params.Scope, ValidScopes) {
		return &ValidationError{Field: "scope", Value: params.Scope, ValidValues: ValidScopes}
	}

	// Validate aggregate function if specified
	if !params.Aggregate.IsEmpty() {
		if !isValidValue(params.Aggregate.Function, ValidAggregateFunctions) {
			return &ValidationError{Field: "aggregate.function", Value: params.Aggregate.Function, ValidValues: ValidAggregateFunctions}
		}
	}

	// Validate output format if specified
	if params.Output.Format != "" {
		if !isValidValue(params.Output.Format, ValidOutputFormats) {
			return &ValidationError{Field: "output.format", Value: params.Output.Format, ValidValues: ValidOutputFormats}
		}
	}

	return nil
}

// ApplyDefaults applies default values to query parameters
func ApplyDefaults(params QueryParams) QueryParams {
	if params.Resource == "" {
		params.Resource = "entries"
	}
	if params.Scope == "" {
		params.Scope = "project"
	}
	if params.Output.Format == "" {
		params.Output.Format = "jsonl"
	}
	return params
}

// isValidValue checks if a value is in a list of valid values
func isValidValue(value string, validValues []string) bool {
	for _, v := range validValues {
		if value == v {
			return true
		}
	}
	return false
}

// ValidationError represents a validation error
type ValidationError struct {
	Field       string
	Value       string
	ValidValues []string
}

func (e *ValidationError) Error() string {
	if len(e.ValidValues) > 0 {
		return "invalid " + e.Field + ": \"" + e.Value + "\", valid values: " + joinStrings(e.ValidValues, ", ")
	}
	return "invalid " + e.Field + ": \"" + e.Value + "\""
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
