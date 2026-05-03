package generator

import (
	"fmt"
	"os"
	"path/filepath"
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
	"HEAD":       {},
	"OPTIONS":    {},
	"body":       {},
	"returns":    {},
	"parameters": {},
	"import":     {},
	"as":         {},
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
	if strings.ContainsRune("()=:,/{}.", rune(char)) {
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

	// Handle hyphen (not part of a negative number — that case handled above)
	if char == '-' {
		l.position++
		return Token{Type: TokenSymbol, Value: "-", Line: l.line}
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
	comments     []Comment         // Buffer for accumulated comments
	baseDir      string            // Directory of the file being parsed (for resolving imports)
	visited      map[string]string // absPath → namespace alias used ("" = flat); shared across recursive parsers
	namespaces   map[string]string // namespace alias → capitalized prefix (e.g. "common" → "Common")
}

// NewParser creates a parser for the given input text.
// filePath (optional) is the absolute (or relative) path of the source file.
// It is used to resolve imports relative to the file's directory, and to mark
// the file as visited so that circular imports back to the root are skipped.
func NewParser(input string, filePath ...string) *Parser {
	lexer := NewLexer(input)
	parser := &Parser{
		lexer:      lexer,
		visited:    make(map[string]string),
		namespaces: make(map[string]string),
	}
	if len(filePath) > 0 {
		if abs, err := filepath.Abs(filePath[0]); err == nil {
			parser.baseDir = filepath.Dir(abs)
			parser.visited[abs] = "" // root file is flat; prevents circular imports back to it
		}
	}
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
	typeNames := make(map[string]bool)
	enumNames := make(map[string]bool)
	serverNames := make(map[string]bool)

	for p.currentToken.Type != TokenEOF {
		// Check for error tokens
		if p.currentToken.Type == TokenError {
			return nil, fmt.Errorf("lexer error: %s", p.currentToken.Value)
		}

		switch p.currentToken.Value {
		case "import":
			if err := p.parseImport(api, typeNames, enumNames, serverNames); err != nil {
				return nil, err
			}
		case "type":
			typeDef, err := p.parseTypeDef()
			if err != nil {
				return nil, err
			}
			if typeNames[typeDef.Name] {
				return nil, fmt.Errorf("type %q is already defined", typeDef.Name)
			}
			typeNames[typeDef.Name] = true
			api.Types = append(api.Types, *typeDef)
		case "enum":
			enumDef, err := p.parseEnumDef()
			if err != nil {
				return nil, err
			}
			if enumNames[enumDef.Name] {
				return nil, fmt.Errorf("enum %q is already defined", enumDef.Name)
			}
			enumNames[enumDef.Name] = true
			api.Enums = append(api.Enums, *enumDef)
		case "server":
			serverDef, err := p.parseServerDef()
			if err != nil {
				return nil, err
			}
			if serverNames[serverDef.Name] {
				return nil, fmt.Errorf("server %q is already defined", serverDef.Name)
			}
			serverNames[serverDef.Name] = true
			api.Servers = append(api.Servers, *serverDef)
		default:
			return nil, fmt.Errorf("unexpected token '%s' at line %d", p.currentToken.Value, p.currentToken.Line)
		}
	}

	if err := validateMapKeys(api); err != nil {
		return nil, err
	}

	return api, nil
}

// validateMapKeys rejects map types whose key type is not comparable. The Go
// language constraint is that map keys must be comparable; using slices or
// maps as keys produces invalid generated Go that fails to compile. Catching
// it at parse time gives a much clearer error.
func validateMapKeys(api *APIDefinition) error {
	typesByName := make(map[string]*TypeDef, len(api.Types))
	for i := range api.Types {
		typesByName[api.Types[i].Name] = &api.Types[i]
	}
	enumsByName := make(map[string]bool, len(api.Enums))
	for _, e := range api.Enums {
		enumsByName[e.Name] = true
	}

	var checkExpr func(expr *TypeExpr, where string) error
	checkExpr = func(expr *TypeExpr, where string) error {
		if expr == nil {
			return nil
		}
		switch expr.Kind {
		case "map":
			if !isComparableType(expr.KeyType, typesByName, enumsByName, map[string]bool{}) {
				return fmt.Errorf("map key in %s is not a comparable type: %s", where, describeType(expr.KeyType))
			}
			if err := checkExpr(expr.KeyType, where); err != nil {
				return err
			}
			return checkExpr(expr.ValueType, where)
		case "repeated":
			return checkExpr(expr.ElementType, where)
		}
		return nil
	}

	for _, td := range api.Types {
		switch td.Kind {
		case "map":
			if !isComparableType(td.KeyType, typesByName, enumsByName, map[string]bool{}) {
				return fmt.Errorf("map key in type %q is not a comparable type: %s", td.Name, describeType(td.KeyType))
			}
			if err := checkExpr(td.KeyType, fmt.Sprintf("type %q", td.Name)); err != nil {
				return err
			}
			if err := checkExpr(td.ValueType, fmt.Sprintf("type %q", td.Name)); err != nil {
				return err
			}
		case "repeated":
			if err := checkExpr(td.ElementType, fmt.Sprintf("type %q", td.Name)); err != nil {
				return err
			}
		case "struct":
			for _, f := range td.Fields {
				if err := checkExpr(f.Type, fmt.Sprintf("type %q field %q", td.Name, f.Name)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

var primitiveTypes = map[string]bool{
	"string": true, "int32": true, "int64": true,
	"float32": true, "float64": true, "bool": true,
	// "file" is a multipart marker, not a real value type, so it is never a valid map key
}

// isComparableType reports whether expr resolves to a Go-comparable type.
// Aliases and structs are followed recursively; visiting tracks names already
// seen so a cyclic alias chain doesn't blow the stack.
func isComparableType(expr *TypeExpr, typesByName map[string]*TypeDef, enumsByName map[string]bool, visiting map[string]bool) bool {
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case "repeated", "map":
		return false
	case "reference":
		if primitiveTypes[expr.Name] {
			return true
		}
		if enumsByName[expr.Name] {
			return true
		}
		if visiting[expr.Name] {
			return true
		}
		td, ok := typesByName[expr.Name]
		if !ok {
			// Unknown reference — let later stages report it; treat as comparable
			// here so we don't double-flag the same problem.
			return true
		}
		visiting[expr.Name] = true
		defer delete(visiting, expr.Name)
		switch td.Kind {
		case "alias":
			if primitiveTypes[td.Target] {
				return true
			}
			return isComparableType(&TypeExpr{Kind: "reference", Name: td.Target}, typesByName, enumsByName, visiting)
		case "struct":
			for _, f := range td.Fields {
				if !isComparableType(f.Type, typesByName, enumsByName, visiting) {
					return false
				}
			}
			return true
		case "repeated", "map":
			return false
		}
	}
	return false
}

func describeType(expr *TypeExpr) string {
	if expr == nil {
		return "<nil>"
	}
	switch expr.Kind {
	case "reference":
		return expr.Name
	case "repeated":
		return "repeated(" + describeType(expr.ElementType) + ")"
	case "map":
		return "map(" + describeType(expr.KeyType) + ": " + describeType(expr.ValueType) + ")"
	}
	return expr.Kind
}

// parseImport handles `import "path/to/file.api"` and `import "path" as "ns"` statements.
// Without `as`, definitions are flat-merged. With `as`, all imported names are prefixed
// with capitalize(ns) and type references within the imported file are rewritten accordingly.
// Circular imports are silently skipped.
func (p *Parser) parseImport(api *APIDefinition, typeNames, enumNames, serverNames map[string]bool) error {
	p.nextToken() // consume "import"

	if p.currentToken.Type != TokenString {
		return fmt.Errorf("expected string path after import at line %d", p.currentToken.Line)
	}

	importPath := p.currentToken.Value
	p.nextToken() // consume the path string

	// Optional: as "namespace"
	var nsAlias, nsPrefix string
	if p.currentToken.Type == TokenKeyword && p.currentToken.Value == "as" {
		p.nextToken() // consume "as"
		if p.currentToken.Type != TokenString {
			return fmt.Errorf("expected string namespace after 'as' at line %d", p.currentToken.Line)
		}
		nsAlias = p.currentToken.Value
		p.nextToken() // consume namespace string
		nsPrefix = capitalizeFirst(nsAlias)
		// Register so this file's type expressions can use namespace.Type syntax
		p.namespaces[nsAlias] = nsPrefix
	}

	// Resolve relative to the current file's directory
	absPath := importPath
	if !filepath.IsAbs(importPath) && p.baseDir != "" {
		absPath = filepath.Join(p.baseDir, importPath)
	}

	var err error
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		return fmt.Errorf("failed to resolve import %q: %v", importPath, err)
	}

	// Check if already imported
	if storedAlias, seen := p.visited[absPath]; seen {
		if storedAlias == nsAlias {
			return nil // same import (circular or diamond with identical namespace) — skip
		}
		if storedAlias == "" {
			return fmt.Errorf("import %q: already imported without a namespace, cannot also import as %q", importPath, nsAlias)
		}
		if nsAlias == "" {
			return fmt.Errorf("import %q: already imported as namespace %q, cannot also import without a namespace", importPath, storedAlias)
		}
		return fmt.Errorf("import %q: already imported as namespace %q, cannot also import as %q", importPath, storedAlias, nsAlias)
	}
	p.visited[absPath] = nsAlias

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read import %q: %v", importPath, err)
	}

	child := &Parser{
		lexer:      NewLexer(string(data)),
		baseDir:    filepath.Dir(absPath),
		visited:    p.visited, // share visited set for circular import detection
		namespaces: make(map[string]string), // each file has its own namespace scope
	}
	child.nextToken()

	imported, err := child.Parse()
	if err != nil {
		return fmt.Errorf("error in import %q: %v", importPath, err)
	}

	// If a namespace was given, rewrite all names and internal type references
	if nsPrefix != "" {
		imported = applyNamespacePrefix(imported, nsPrefix)
	}

	// Flat-merge, checking for conflicts using the (possibly prefixed) names
	for _, t := range imported.Types {
		if typeNames[t.Name] {
			return fmt.Errorf("import %q: type %q is already defined", importPath, t.Name)
		}
		typeNames[t.Name] = true
		api.Types = append(api.Types, t)
	}
	for _, e := range imported.Enums {
		if enumNames[e.Name] {
			return fmt.Errorf("import %q: enum %q is already defined", importPath, e.Name)
		}
		enumNames[e.Name] = true
		api.Enums = append(api.Enums, e)
	}
	for _, s := range imported.Servers {
		if serverNames[s.Name] {
			return fmt.Errorf("import %q: server %q is already defined", importPath, s.Name)
		}
		serverNames[s.Name] = true
		api.Servers = append(api.Servers, s)
	}

	return nil
}

// capitalizeFirst returns s with its first character uppercased.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// applyNamespacePrefix rewrites all definition names and internal type references
// in imported to have the given prefix. This makes types from `import "x.api" as "ns"`
// become e.g. NsUser, NsRole — preventing conflicts and making origin clear.
func applyNamespacePrefix(imported *APIDefinition, prefix string) *APIDefinition {
	// Collect original names so we know which references to rewrite
	definedNames := make(map[string]bool, len(imported.Types)+len(imported.Enums)+len(imported.Servers))
	for _, t := range imported.Types {
		definedNames[t.Name] = true
	}
	for _, e := range imported.Enums {
		definedNames[e.Name] = true
	}
	for _, s := range imported.Servers {
		definedNames[s.Name] = true
	}

	result := &APIDefinition{}

	for _, t := range imported.Types {
		t.Name = prefix + t.Name
		t.Target = prefixName(t.Target, prefix, definedNames)
		for i, f := range t.Fields {
			t.Fields[i].Type = prefixTypeExpr(f.Type, prefix, definedNames)
		}
		t.ElementType = prefixTypeExpr(t.ElementType, prefix, definedNames)
		t.KeyType = prefixTypeExpr(t.KeyType, prefix, definedNames)
		t.ValueType = prefixTypeExpr(t.ValueType, prefix, definedNames)
		result.Types = append(result.Types, t)
	}
	for _, e := range imported.Enums {
		e.Name = prefix + e.Name
		result.Enums = append(result.Enums, e)
	}
	for _, s := range imported.Servers {
		s.Name = prefix + s.Name
		for i, ep := range s.Endpoints {
			s.Endpoints[i].Body = prefixTypeExpr(ep.Body, prefix, definedNames)
			s.Endpoints[i].Returns = prefixTypeExpr(ep.Returns, prefix, definedNames)
			s.Endpoints[i].Parameters = prefixTypeExpr(ep.Parameters, prefix, definedNames)
		}
		result.Servers = append(result.Servers, s)
	}

	return result
}

// prefixName prefixes name if it is a user-defined name from the imported file.
func prefixName(name, prefix string, definedNames map[string]bool) string {
	if definedNames[name] {
		return prefix + name
	}
	return name
}

// prefixTypeExpr rewrites type references that refer to imported definitions.
func prefixTypeExpr(expr *TypeExpr, prefix string, definedNames map[string]bool) *TypeExpr {
	if expr == nil {
		return nil
	}
	result := *expr
	switch expr.Kind {
	case "reference":
		result.Name = prefixName(expr.Name, prefix, definedNames)
	case "repeated":
		result.ElementType = prefixTypeExpr(expr.ElementType, prefix, definedNames)
	case "map":
		result.KeyType = prefixTypeExpr(expr.KeyType, prefix, definedNames)
		result.ValueType = prefixTypeExpr(expr.ValueType, prefix, definedNames)
	}
	return &result
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
		name, err := p.expectIdentifier()
		if err != nil {
			return nil, err
		}
		// Check for namespace.TypeName (e.g. common.User)
		if p.currentToken.Type == TokenSymbol && p.currentToken.Value == "." {
			p.nextToken() // consume "."
			typeName, err := p.expectIdentifier()
			if err != nil {
				return nil, err
			}
			prefix, ok := p.namespaces[name]
			if !ok {
				return nil, fmt.Errorf("unknown namespace %q", name)
			}
			return &TypeExpr{Kind: "reference", Name: prefix + typeName}, nil
		}
		return &TypeExpr{Kind: "reference", Name: name}, nil
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

	// afterSeparator tracks whether we just consumed a "/" or "-", meaning the
	// next token is unambiguously a path segment even if it's a keyword.
	afterSeparator := false

	// Build path by consuming tokens until we hit a non-path token
	for {
		if p.currentToken.Value == "/" {
			pathParts = append(pathParts, "/")
			p.nextToken()
			afterSeparator = true
		} else if afterSeparator && (p.currentToken.Type == TokenIdentifier || p.currentToken.Type == TokenKeyword) {
			pathParts = append(pathParts, p.currentToken.Value)
			p.nextToken()
			afterSeparator = false
		} else if p.currentToken.Value == "-" {
			pathParts = append(pathParts, "-")
			p.nextToken()
			afterSeparator = true
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
			afterSeparator = false
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
