package validation

// Validator runs validation checks on tools
type Validator struct {
	fast bool
}

// NewValidator creates a new validator
func NewValidator(fast bool) *Validator {
	return &Validator{fast: fast}
}

// Validate runs all validation checks on provided tools
func (v *Validator) Validate(tools []Tool) *Report {
	report := &Report{
		TotalTools: len(tools),
		Results:    []Result{},
	}

	for _, tool := range tools {
		// Check 1: Naming pattern validation
		namingResult := ValidateNaming(tool)
		report.Results = append(report.Results, namingResult)
		report.ChecksRun++

		// Check 2: Parameter ordering validation
		orderingResult := ValidateParameterOrdering(tool)
		report.Results = append(report.Results, orderingResult)
		report.ChecksRun++

		// Check 3: Description format validation
		descResult := ValidateDescription(tool)
		report.Results = append(report.Results, descResult)
		report.ChecksRun++
	}

	// Calculate summary stats
	for _, result := range report.Results {
		switch result.Status {
		case "PASS":
			report.Passed++
		case "FAIL":
			report.Failed++
		case "WARN":
			report.Warnings++
		}
	}

	// Generate summary message
	if report.Failed > 0 {
		report.Summary = "FAILED"
	} else if report.Warnings > 0 {
		report.Summary = "PASSED (with warnings)"
	} else {
		report.Summary = "PASSED"
	}

	return report
}
