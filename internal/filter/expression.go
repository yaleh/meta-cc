package filter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Expression represents a filter expression that can be evaluated
type Expression interface {
	Evaluate(record map[string]interface{}) (bool, error)
}

// ComparisonExpression represents a comparison operation (=, !=, >, <, >=, <=)
type ComparisonExpression struct {
	Field    string
	Operator string
	Value    interface{}
}

func (e *ComparisonExpression) Evaluate(record map[string]interface{}) (bool, error) {
	fieldValue, exists := record[e.Field]
	if !exists {
		return false, nil
	}

	return compareValues(fieldValue, e.Operator, e.Value)
}

// BinaryExpression represents a binary operation (AND, OR)
type BinaryExpression struct {
	Operator string
	Left     Expression
	Right    Expression
}

func (e *BinaryExpression) Evaluate(record map[string]interface{}) (bool, error) {
	left, err := e.Left.Evaluate(record)
	if err != nil {
		return false, err
	}

	right, err := e.Right.Evaluate(record)
	if err != nil {
		return false, err
	}

	switch e.Operator {
	case "AND":
		return left && right, nil
	case "OR":
		return left || right, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", e.Operator)
	}
}

// UnaryExpression represents a unary operation (NOT)
type UnaryExpression struct {
	Operator string
	Operand  Expression
}

func (e *UnaryExpression) Evaluate(record map[string]interface{}) (bool, error) {
	result, err := e.Operand.Evaluate(record)
	if err != nil {
		return false, err
	}
	return !result, nil
}

// InExpression represents set membership check (IN, NOT IN)
type InExpression struct {
	Field  string
	Values []interface{}
	Negate bool
}

func (e *InExpression) Evaluate(record map[string]interface{}) (bool, error) {
	fieldValue, exists := record[e.Field]
	if !exists {
		return false, nil
	}

	found := false
	for _, v := range e.Values {
		if valueEquals(fieldValue, v) {
			found = true
			break
		}
	}

	if e.Negate {
		return !found, nil
	}
	return found, nil
}

// BetweenExpression represents range check (BETWEEN ... AND ...)
type BetweenExpression struct {
	Field string
	Lower interface{}
	Upper interface{}
}

func (e *BetweenExpression) Evaluate(record map[string]interface{}) (bool, error) {
	fieldValue, exists := record[e.Field]
	if !exists {
		return false, nil
	}

	lowerOk, _ := compareValues(fieldValue, ">=", e.Lower)
	upperOk, _ := compareValues(fieldValue, "<=", e.Upper)

	return lowerOk && upperOk, nil
}

// LikeExpression represents SQL LIKE pattern matching
type LikeExpression struct {
	Field   string
	Pattern string
}

func (e *LikeExpression) Evaluate(record map[string]interface{}) (bool, error) {
	fieldValue, exists := record[e.Field]
	if !exists {
		return false, nil
	}

	str, ok := fieldValue.(string)
	if !ok {
		return false, nil
	}

	// Convert SQL LIKE pattern to regex
	var pattern strings.Builder
	pattern.WriteString("^")

	for i := 0; i < len(e.Pattern); i++ {
		ch := e.Pattern[i]
		switch ch {
		case '%':
			pattern.WriteString(".*")
		case '_':
			pattern.WriteString(".")
		default:
			// Escape regex special characters
			pattern.WriteString(regexp.QuoteMeta(string(ch)))
		}
	}

	pattern.WriteString("$")

	matched, _ := regexp.MatchString(pattern.String(), str)
	return matched, nil
}

// RegexpExpression represents regular expression matching
type RegexpExpression struct {
	Field   string
	Pattern string
}

func (e *RegexpExpression) Evaluate(record map[string]interface{}) (bool, error) {
	fieldValue, exists := record[e.Field]
	if !exists {
		return false, nil
	}

	str, ok := fieldValue.(string)
	if !ok {
		return false, nil
	}

	matched, err := regexp.MatchString(e.Pattern, str)
	if err != nil {
		return false, fmt.Errorf("invalid regexp: %v", err)
	}

	return matched, nil
}

// compareValues compares two values with the given operator
func compareValues(left interface{}, operator string, right interface{}) (bool, error) {
	// String comparison
	leftStr, leftIsStr := left.(string)
	rightStr, rightIsStr := right.(string)

	if leftIsStr && rightIsStr {
		switch operator {
		case "=":
			return leftStr == rightStr, nil
		case "!=":
			return leftStr != rightStr, nil
		case ">":
			return leftStr > rightStr, nil
		case "<":
			return leftStr < rightStr, nil
		case ">=":
			return leftStr >= rightStr, nil
		case "<=":
			return leftStr <= rightStr, nil
		}
	}

	// Numeric comparison
	leftNum, leftErr := toFloat64(left)
	rightNum, rightErr := toFloat64(right)

	if leftErr == nil && rightErr == nil {
		switch operator {
		case "=":
			return leftNum == rightNum, nil
		case "!=":
			return leftNum != rightNum, nil
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}

	return false, fmt.Errorf("unsupported comparison: %v %s %v", left, operator, right)
}

// valueEquals checks if two values are equal
func valueEquals(left, right interface{}) bool {
	if left == right {
		return true
	}

	// String comparison
	leftStr, leftIsStr := left.(string)
	rightStr, rightIsStr := right.(string)
	if leftIsStr && rightIsStr {
		return leftStr == rightStr
	}

	// Numeric comparison
	leftNum, leftErr := toFloat64(left)
	rightNum, rightErr := toFloat64(right)
	if leftErr == nil && rightErr == nil {
		return leftNum == rightNum
	}

	return false
}

// toFloat64 converts a value to float64
func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert to float64: %v", val)
	}
}
