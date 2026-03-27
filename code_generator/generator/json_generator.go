package generator

import (
	"fmt"
	"strings"
)

// Token types
type TokenType int

const (
	TokenIdentifier TokenType = iota
	TokenKeyword
	TokenString
	TokenNumber
	TokenSymbol
	TokenComment
	TokenError
	TokenEOF
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
}

// Comment represents a comment in the source
type Comment struct {
	Type    string `json:"type"`    // "single" or "multi"
	Content string `json:"content"` // Comment text without // or /* */
	Line    int    `json:"line"`
}

// Keywords defined at package level for better maintenance and performance
var keywords = map[string]struct{}{
	"type":       {},
	"enum":       {},
	"server":     {},
	"required":   {},
	"repeated":   {},
	"map":        {},
	"GET":        {},
	"POST":       {},
	"PUT":        {},
	"DELETE":     {},
	"PATCH":      {},
	"body":       {},
	"returns":    {},
	"parameters": {},
}

// Helper function for keyword checking
func isKeyword(word string) bool {
	_, exists := keywords[word]
	return exists
}

// Lexer
type Lexer struct {
	input    string
	position int
	line     int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		line:  1,
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.position >= len(l.input) {
		return Token{Type: TokenEOF, Line: l.line}
	}

	char := l.input[l.position]

	// Handle comments
	if char == '/' && l.position+1 < len(l.input) {
		nextChar := l.input[l.position+1]
		if nextChar == '/' {
			return l.readSingleLineComment()
		} else if nextChar == '*' {
			return l.readMultiLineComment()
		}
	}

	// Handle symbols
	if strings.ContainsRune("()=:,/{}", rune(char)) {
		l.position++
		return Token{Type: TokenSymbol, Value: string(char), Line: l.line}
	}

	// Handle strings
	if char == '"' {
		return l.readString()
	}

	// Handle numbers
	if isDigit(char) || (char == '-' && l.position+1 < len(l.input) && isDigit(l.input[l.position+1])) {
		return l.readNumber()
	}

	// Handle identifiers and keywords
	if isAlpha(char) {
		return l.readIdentifier()
	}

	// Skip unknown characters
	l.position++
	return l.NextToken()
}

func (l *Lexer) skipWhitespace() {
	for l.position < len(l.input) {
		char := l.input[l.position]
		if char == '\n' {
			l.line++
		}
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			l.position++
		} else {
			break
		}
	}
}

func (l *Lexer) readSingleLineComment() Token {
	startLine := l.line
	l.position += 2 // Skip //
	start := l.position

	// Read until end of line
	for l.position < len(l.input) && l.input[l.position] != '\n' {
		l.position++
	}

	content := strings.TrimSpace(l.input[start:l.position])
	return Token{Type: TokenComment, Value: fmt.Sprintf("single:%s", content), Line: startLine}
}

func (l *Lexer) readMultiLineComment() Token {
	startLine := l.line
	l.position += 2 // Skip /*
	start := l.position

	// Read until */
	for l.position < len(l.input)-1 {
		if l.input[l.position] == '*' && l.input[l.position+1] == '/' {
			content := strings.TrimSpace(l.input[start:l.position])
			l.position += 2 // Skip */
			return Token{Type: TokenComment, Value: fmt.Sprintf("multi:%s", content), Line: startLine}
		}
		if l.input[l.position] == '\n' {
			l.line++
		}
		l.position++
	}

	// Unterminated comment
	return Token{
		Type:  TokenError,
		Value: fmt.Sprintf("unterminated comment starting at line %d", startLine),
		Line:  startLine,
	}
}

func (l *Lexer) readString() Token {
	startLine := l.line
	l.position++ // Skip opening quote
	var result strings.Builder

	for l.position < len(l.input) {
		char := l.input[l.position]

		// Check for unterminated string at newline
		if char == '\n' {
			return Token{
				Type:  TokenError,
				Value: fmt.Sprintf("unterminated string at line %d", startLine),
				Line:  startLine,
			}
		}

		// Handle escaped characters
		if char == '\\' && l.position+1 < len(l.input) {
			l.position++ // Skip the backslash
			nextChar := l.input[l.position]
			switch nextChar {
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			default:
				// Unknown escape sequence - pass through with backslash
				result.WriteByte('\\')
				result.WriteByte(nextChar)
			}
			l.position++
			continue
		}

		// Found closing quote
		if char == '"' {
			l.position++ // Skip closing quote
			return Token{Type: TokenString, Value: result.String(), Line: startLine}
		}

		result.WriteByte(char)
		l.position++
	}

	// Reached end of input without finding closing quote
	return Token{
		Type:  TokenError,
		Value: fmt.Sprintf("unterminated string starting at line %d", startLine),
		Line:  startLine,
	}
}

func (l *Lexer) readNumber() Token {
	start := l.position
	hasDecimal := false

	if l.input[l.position] == '-' {
		l.position++
	}

	for l.position < len(l.input) {
		char := l.input[l.position]
		if isDigit(char) {
			l.position++
		} else if char == '.' && !hasDecimal {
			hasDecimal = true
			l.position++
		} else {
			break
		}
	}

	value := l.input[start:l.position]

	// Validate the number format
	if strings.HasSuffix(value, ".") || strings.Contains(value, "..") {
		return Token{
			Type:  TokenError,
			Value: fmt.Sprintf("invalid number format '%s'", value),
			Line:  l.line,
		}
	}

	return Token{Type: TokenNumber, Value: value, Line: l.line}
}

func (l *Lexer) readIdentifier() Token {
	start := l.position

	for l.position < len(l.input) && (isAlpha(l.input[l.position]) || isDigit(l.input[l.position]) || l.input[l.position] == '_') {
		l.position++
	}

	value := l.input[start:l.position]
	tokenType := TokenIdentifier

	// Check if it's a keyword using the improved approach
	if isKeyword(value) {
		tokenType = TokenKeyword
	}

	return Token{Type: tokenType, Value: value, Line: l.line}
}

func isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

// Helper function to parse comment content
func parseCommentContent(tokenValue string) Comment {
	parts := strings.SplitN(tokenValue, ":", 2)
	if len(parts) != 2 {
		return Comment{Type: "single", Content: tokenValue}
	}
	return Comment{Type: parts[0], Content: parts[1]}
}

// AST Node types for JSON output with comments
type APIDefinition struct {
	Types   []TypeDef   `json:"types"`
	Enums   []EnumDef   `json:"enums"`
	Servers []ServerDef `json:"servers"`
}

type TypeDef struct {
	Name        string     `json:"name"`
	Kind        string     `json:"kind"`                  // "alias", "struct", "repeated", "map"
	Target      string     `json:"target,omitempty"`      // For aliases
	Fields      []FieldDef `json:"fields,omitempty"`      // For structs
	ElementType *TypeExpr  `json:"elementType,omitempty"` // For repeated
	KeyType     *TypeExpr  `json:"keyType,omitempty"`     // For maps
	ValueType   *TypeExpr  `json:"valueType,omitempty"`   // For maps
	Comments    []Comment  `json:"comments,omitempty"`    // Associated comments
}

type TypeExpr struct {
	Kind        string    `json:"kind"`                  // "reference", "repeated", "map"
	Name        string    `json:"name,omitempty"`        // For type references
	ElementType *TypeExpr `json:"elementType,omitempty"` // For repeated
	KeyType     *TypeExpr `json:"keyType,omitempty"`     // For maps
	ValueType   *TypeExpr `json:"valueType,omitempty"`   // For maps
}

type FieldDef struct {
	Name     string    `json:"name"`
	Type     *TypeExpr `json:"type"`
	Required bool      `json:"required"`
	Comments []Comment `json:"comments,omitempty"` // Associated comments
}

type EnumDef struct {
	Name     string            `json:"name"`
	BaseType string            `json:"baseType"`
	Values   map[string]string `json:"values"`
	Comments []Comment         `json:"comments,omitempty"` // Associated comments
}

type ServerDef struct {
	Name      string        `json:"name"`
	Endpoints []EndpointDef `json:"endpoints"`
	Comments  []Comment     `json:"comments,omitempty"` // Associated comments
}

type EndpointDef struct {
	Name       string     `json:"name"`
	Method     string     `json:"method"`
	Path       string     `json:"path"`
	PathParams []ParamDef `json:"pathParams,omitempty"`
	Parameters *TypeExpr  `json:"parameters,omitempty"`
	Body       *TypeExpr  `json:"body,omitempty"`
	Returns    *TypeExpr  `json:"returns,omitempty"`
	Comments   []Comment  `json:"comments,omitempty"` // Associated comments
}

type ParamDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Parser with comment handling
type Parser struct {
	lexer        *Lexer
	currentToken Token
	comments     []Comment // Buffer for accumulated comments
}

func NewParser(input string) *Parser {
	lexer := NewLexer(input)
	parser := &Parser{lexer: lexer}
	parser.nextToken()
	return parser
}

func (p *Parser) nextToken() {
	p.currentToken = p.lexer.NextToken()

	// Collect comments
	for p.currentToken.Type == TokenComment {
		comment := parseCommentContent(p.currentToken.Value)
		comment.Line = p.currentToken.Line
		p.comments = append(p.comments, comment)
		p.currentToken = p.lexer.NextToken()
	}
}

// Get and clear accumulated comments
func (p *Parser) getComments() []Comment {
	comments := p.comments
	p.comments = nil
	return comments
}

func (p *Parser) expect(value string) error {
	if p.currentToken.Value != value {
		return fmt.Errorf("expected '%s', got '%s' at line %d", value, p.currentToken.Value, p.currentToken.Line)
	}
	p.nextToken()
	return nil
}

func (p *Parser) expectIdentifier() (string, error) {
	if p.currentToken.Type != TokenIdentifier && p.currentToken.Type != TokenKeyword {
		return "", fmt.Errorf("expected identifier, got '%s' at line %d", p.currentToken.Value, p.currentToken.Line)
	}
	value := p.currentToken.Value
	p.nextToken()
	return value, nil
}

func (p *Parser) Parse() (*APIDefinition, error) {
	api := &APIDefinition{}

	for p.currentToken.Type != TokenEOF {
		// Check for error tokens
		if p.currentToken.Type == TokenError {
			return nil, fmt.Errorf("lexer error: %s", p.currentToken.Value)
		}

		switch p.currentToken.Value {
		case "type":
			typeDef, err := p.parseTypeDef()
			if err != nil {
				return nil, err
			}
			api.Types = append(api.Types, *typeDef)
		case "enum":
			enumDef, err := p.parseEnumDef()
			if err != nil {
				return nil, err
			}
			api.Enums = append(api.Enums, *enumDef)
		case "server":
			serverDef, err := p.parseServerDef()
			if err != nil {
				return nil, err
			}
			api.Servers = append(api.Servers, *serverDef)
		default:
			return nil, fmt.Errorf("unexpected token '%s' at line %d", p.currentToken.Value, p.currentToken.Line)
		}
	}

	return api, nil
}

func (p *Parser) parseTypeDef() (*TypeDef, error) {
	// Get comments that appeared before 'type'
	comments := p.getComments()

	if err := p.expect("type"); err != nil {
		return nil, err
	}

	name, err := p.expectIdentifier()
	if err != nil {
		return nil, err
	}

	typeDef := &TypeDef{Name: name, Comments: comments}

	// Check if it's a struct
	if p.currentToken.Value == "{" {
		typeDef.Kind = "struct"
		fields, err := p.parseStruct()
		if err != nil {
			return nil, err
		}
		typeDef.Fields = fields
	} else {
		// Parse type expression
		typeExpr, err := p.parseTypeExpression()
		if err != nil {
			return nil, err
		}

		if typeExpr.Kind == "reference" {
			typeDef.Kind = "alias"
			typeDef.Target = typeExpr.Name
		} else {
			typeDef.Kind = typeExpr.Kind
			typeDef.ElementType = typeExpr.ElementType
			typeDef.KeyType = typeExpr.KeyType
			typeDef.ValueType = typeExpr.ValueType
		}
	}

	return typeDef, nil
}

func (p *Parser) parseTypeExpression() (*TypeExpr, error) {
	switch p.currentToken.Value {
	case "repeated":
		return p.parseRepeatedExpr()
	case "map":
		return p.parseMapExpr()
	default:
		// Simple type reference
		typeName, err := p.expectIdentifier()
		if err != nil {
			return nil, err
		}
		return &TypeExpr{Kind: "reference", Name: typeName}, nil
	}
}

func (p *Parser) parseRepeatedExpr() (*TypeExpr, error) {
	if err := p.expect("repeated"); err != nil {
		return nil, err
	}
	if err := p.expect("("); err != nil {
		return nil, err
	}

	elementType, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expect(")"); err != nil {
		return nil, err
	}

	return &TypeExpr{Kind: "repeated", ElementType: elementType}, nil
}

func (p *Parser) parseMapExpr() (*TypeExpr, error) {
	if err := p.expect("map"); err != nil {
		return nil, err
	}
	if err := p.expect("("); err != nil {
		return nil, err
	}

	keyType, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expect(":"); err != nil {
		return nil, err
	}

	valueType, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expect(")"); err != nil {
		return nil, err
	}

	return &TypeExpr{Kind: "map", KeyType: keyType, ValueType: valueType}, nil
}

func (p *Parser) parseStruct() ([]FieldDef, error) {
	if err := p.expect("{"); err != nil {
		return nil, err
	}

	var fields []FieldDef

	for p.currentToken.Value != "}" {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, *field)
	}

	if err := p.expect("}"); err != nil {
		return nil, err
	}

	return fields, nil
}

func (p *Parser) parseField() (*FieldDef, error) {
	// Get comments that appeared before this field
	comments := p.getComments()

	field := &FieldDef{Comments: comments}

	// Check for required keyword
	if p.currentToken.Value == "required" {
		field.Required = true
		p.nextToken()
	}

	name, err := p.expectIdentifier()
	if err != nil {
		return nil, err
	}
	field.Name = name

	typeExpr, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	field.Type = typeExpr

	return field, nil
}

func (p *Parser) parseEnumDef() (*EnumDef, error) {
	// Get comments that appeared before 'enum'
	comments := p.getComments()

	if err := p.expect("enum"); err != nil {
		return nil, err
	}

	name, err := p.expectIdentifier()
	if err != nil {
		return nil, err
	}

	baseType, err := p.expectIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.expect("{"); err != nil {
		return nil, err
	}

	values := make(map[string]string)

	for p.currentToken.Value != "}" {
		// Skip comments inside enum (could be enhanced to associate with enum values)
		p.getComments()

		key, err := p.expectIdentifier()
		if err != nil {
			return nil, err
		}

		if err := p.expect("="); err != nil {
			return nil, err
		}

		var value string
		if p.currentToken.Type == TokenString {
			value = p.currentToken.Value
			p.nextToken()
		} else if p.currentToken.Type == TokenNumber {
			value = p.currentToken.Value
			p.nextToken()
		} else {
			return nil, fmt.Errorf("expected string or number value at line %d", p.currentToken.Line)
		}

		values[key] = value
	}

	if err := p.expect("}"); err != nil {
		return nil, err
	}

	return &EnumDef{
		Name:     name,
		BaseType: baseType,
		Values:   values,
		Comments: comments,
	}, nil
}

func (p *Parser) parseServerDef() (*ServerDef, error) {
	// Get comments that appeared before 'server'
	comments := p.getComments()

	if err := p.expect("server"); err != nil {
		return nil, err
	}

	name, err := p.expectIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.expect("{"); err != nil {
		return nil, err
	}

	var endpoints []EndpointDef

	for p.currentToken.Value != "}" {
		endpoint, err := p.parseEndpoint()
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, *endpoint)
	}

	if err := p.expect("}"); err != nil {
		return nil, err
	}

	return &ServerDef{
		Name:      name,
		Endpoints: endpoints,
		Comments:  comments,
	}, nil
}

func (p *Parser) parseEndpoint() (*EndpointDef, error) {
	// Get comments that appeared before this endpoint
	comments := p.getComments()

	name, err := p.expectIdentifier()
	if err != nil {
		return nil, err
	}

	method, err := p.expectIdentifier()
	if err != nil {
		return nil, err
	}

	path, pathParams, err := p.parsePath()
	if err != nil {
		return nil, err
	}

	endpoint := &EndpointDef{
		Name:       name,
		Method:     method,
		Path:       path,
		PathParams: pathParams,
		Comments:   comments,
	}

	// Parse optional clauses
	for p.currentToken.Value == "parameters" || p.currentToken.Value == "body" || p.currentToken.Value == "returns" {
		switch p.currentToken.Value {
		case "parameters":
			p.nextToken()
			if err := p.expect("("); err != nil {
				return nil, err
			}
			paramType, err := p.parseTypeExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expect(")"); err != nil {
				return nil, err
			}
			endpoint.Parameters = paramType
		case "body":
			p.nextToken()
			if err := p.expect("("); err != nil {
				return nil, err
			}
			bodyType, err := p.parseTypeExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expect(")"); err != nil {
				return nil, err
			}
			endpoint.Body = bodyType
		case "returns":
			p.nextToken()
			if err := p.expect("("); err != nil {
				return nil, err
			}
			returnType, err := p.parseReturnTypeExpr()
			if err != nil {
				return nil, err
			}
			if err := p.expect(")"); err != nil {
				return nil, err
			}
			endpoint.Returns = returnType
}
	}

	return endpoint, nil
}

func (p *Parser) parsePath() (string, []ParamDef, error) {
	// Expect path to start with /
	if p.currentToken.Value != "/" {
		return "", nil, fmt.Errorf("expected path starting with '/' at line %d", p.currentToken.Line)
	}

	var pathParts []string
	var params []ParamDef

	// Build path by consuming tokens until we hit a non-path token
	for {
		if p.currentToken.Value == "/" {
			pathParts = append(pathParts, "/")
			p.nextToken()
		} else if p.currentToken.Type == TokenIdentifier {
			pathParts = append(pathParts, p.currentToken.Value)
			p.nextToken()
		} else if p.currentToken.Value == "{" {
			// Parse path parameter
			p.nextToken()
			paramName, err := p.expectIdentifier()
			if err != nil {
				return "", nil, err
			}

			paramType, err := p.expectIdentifier()
			if err != nil {
				return "", nil, err
			}

			if err := p.expect("}"); err != nil {
				return "", nil, err
			}

			params = append(params, ParamDef{
				Name: paramName,
				Type: paramType,
			})

			// Add parameter placeholder using {paramName} format
			pathParts = append(pathParts, fmt.Sprintf("{%s}", paramName))
		} else {
			// End of path
			break
		}
	}

	path := strings.Join(pathParts, "")
	return path, params, nil
}

func (p *Parser) parseReturnTypeExpr() (*TypeExpr, error) {
	return p.parseTypeExpression()
}
