package filter

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseExpression parses a filter expression string into an Expression tree
func ParseExpression(expr string) (Expression, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	parser := &ExpressionParser{
		input: expr,
		pos:   0,
	}

	return parser.parse()
}

// ExpressionParser is a simple recursive descent parser
type ExpressionParser struct {
	input string
	pos   int
}

func (p *ExpressionParser) parse() (Expression, error) {
	return p.parseOr()
}

func (p *ExpressionParser) parseOr() (Expression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.matchKeyword("OR") {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpression{
			Operator: "OR",
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

func (p *ExpressionParser) parseAnd() (Expression, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.matchKeyword("AND") {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpression{
			Operator: "AND",
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

func (p *ExpressionParser) parseUnary() (Expression, error) {
	if p.matchKeyword("NOT") {
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpression{
			Operator: "NOT",
			Operand:  operand,
		}, nil
	}

	return p.parsePrimary()
}

func (p *ExpressionParser) parsePrimary() (Expression, error) {
	p.skipWhitespace()

	// Handle parenthesized expressions
	if p.match("(") {
		expr, err := p.parse()
		if err != nil {
			return nil, err
		}
		if !p.match(")") {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		return expr, nil
	}

	// Parse field name
	field := p.parseIdentifier()
	if field == "" {
		return nil, fmt.Errorf("expected field name at position %d", p.pos)
	}

	p.skipWhitespace()

	// Check for special operators
	if p.match("IN (") {
		values, err := p.parseValueList()
		if err != nil {
			return nil, err
		}
		return &InExpression{Field: field, Values: values, Negate: false}, nil
	}

	if p.match("NOT IN (") {
		values, err := p.parseValueList()
		if err != nil {
			return nil, err
		}
		return &InExpression{Field: field, Values: values, Negate: true}, nil
	}

	if p.match("BETWEEN ") {
		lower, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		// Match AND (without leading/trailing space check to avoid conflict with boolean AND)
		if !p.matchExact("AND") {
			return nil, fmt.Errorf("BETWEEN requires AND")
		}
		p.skipWhitespace()
		upper, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return &BetweenExpression{Field: field, Lower: lower, Upper: upper}, nil
	}

	if p.match("LIKE ") {
		pattern, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		patternStr, ok := pattern.(string)
		if !ok {
			return nil, fmt.Errorf("LIKE pattern must be a string")
		}
		return &LikeExpression{Field: field, Pattern: patternStr}, nil
	}

	if p.match("REGEXP ") {
		pattern, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		patternStr, ok := pattern.(string)
		if !ok {
			return nil, fmt.Errorf("REGEXP pattern must be a string")
		}
		return &RegexpExpression{Field: field, Pattern: patternStr}, nil
	}

	// Parse comparison operator
	var op string
	if p.match("!=") {
		op = "!="
	} else if p.match(">=") {
		op = ">="
	} else if p.match("<=") {
		op = "<="
	} else if p.match("=") {
		op = "="
	} else if p.match(">") {
		op = ">"
	} else if p.match("<") {
		op = "<"
	} else {
		return nil, fmt.Errorf("expected operator at position %d", p.pos)
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	return &ComparisonExpression{Field: field, Operator: op, Value: value}, nil
}

func (p *ExpressionParser) parseIdentifier() string {
	p.skipWhitespace()
	start := p.pos

	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			p.pos++
		} else {
			break
		}
	}

	return p.input[start:p.pos]
}

func (p *ExpressionParser) parseValue() (interface{}, error) {
	p.skipWhitespace()

	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	// String value (quoted)
	if p.input[p.pos] == '\'' {
		return p.parseQuotedString()
	}

	// Numeric value
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' {
			p.pos++
		} else {
			break
		}
	}

	if p.pos > start {
		numStr := p.input[start:p.pos]
		if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
			return int(num), nil
		}
		if num, err := strconv.ParseFloat(numStr, 64); err == nil {
			return num, nil
		}
		return numStr, nil // Return as string if not numeric
	}

	return nil, fmt.Errorf("expected value at position %d", p.pos)
}

func (p *ExpressionParser) parseQuotedString() (string, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '\'' {
		return "", fmt.Errorf("expected quoted string")
	}

	p.pos++ // Skip opening quote
	start := p.pos

	for p.pos < len(p.input) {
		if p.input[p.pos] == '\'' {
			str := p.input[start:p.pos]
			p.pos++ // Skip closing quote
			return str, nil
		}
		p.pos++
	}

	return "", fmt.Errorf("unclosed quote")
}

func (p *ExpressionParser) parseValueList() ([]interface{}, error) {
	var values []interface{}

	for {
		p.skipWhitespace()
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		values = append(values, value)

		p.skipWhitespace()
		if p.match(",") {
			continue
		}
		if p.match(")") {
			break
		}
		return nil, fmt.Errorf("expected ',' or ')' in value list")
	}

	return values, nil
}

func (p *ExpressionParser) match(s string) bool {
	p.skipWhitespace()
	if strings.HasPrefix(p.input[p.pos:], s) {
		p.pos += len(s)
		return true
	}
	return false
}

func (p *ExpressionParser) matchKeyword(keyword string) bool {
	p.skipWhitespace()
	if strings.HasPrefix(p.input[p.pos:], keyword) {
		// Check word boundary (next char should be whitespace or end of input)
		endPos := p.pos + len(keyword)
		if endPos < len(p.input) {
			nextCh := p.input[endPos]
			if !(nextCh == ' ' || nextCh == '\t' || nextCh == ')' || nextCh == '(') {
				return false
			}
		}
		p.pos = endPos
		return true
	}
	return false
}

func (p *ExpressionParser) matchExact(s string) bool {
	if strings.HasPrefix(p.input[p.pos:], s) {
		p.pos += len(s)
		return true
	}
	return false
}

func (p *ExpressionParser) skipWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t') {
		p.pos++
	}
}
