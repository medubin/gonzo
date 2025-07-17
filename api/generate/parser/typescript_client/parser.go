package typescript_client

import (
	"fmt"
	"strings"

	lex "github.com/medubin/gonzo/api/generate/lexer"
)

type Parser struct {
	tokens   []lex.Token
	position int
}

func NewParser(lexer lex.Lexer) Parser {
	return Parser{
		tokens:   lexer.Tokens,
		position: 0,
	}
}

func (p *Parser) isEOF() bool {
	return p.position >= len(p.tokens)
}

func (p *Parser) peek() *lex.Token {
	if p.isEOF() {
		return nil
	}
	return &p.tokens[p.position]
}

func (p *Parser) next() *lex.Token {
	if p.isEOF() {
		return nil
	}
	token := p.tokens[p.position]
	p.position++
	return &token
}

func (p *Parser) skipWhitespace() {
	for !p.isEOF() {
		tokenType := p.peek().Type
		if tokenType == lex.SPACE || tokenType == lex.TAB || tokenType == lex.NEWLINE || tokenType == lex.SINGLE_LINE_COMMENT || tokenType == lex.MULTI_LINE_COMMENT {
			p.next()
		} else {
			break
		}
	}
}

var typeMappings = map[string]string{
	"string": "string",
	"int":    "number",
	"int32":  "number",
	"int64":  "number",
	"float":  "number",
	"float32":"number",
	"float64":"number",
	"bool":   "boolean",
}

func (p *Parser) mapType(goType string) string {
	if tsType, ok := typeMappings[goType]; ok {
		return tsType
	}
	return goType
}

func (p *Parser) Parse() (string, error) {
	var result strings.Builder
	for !p.isEOF() {
		p.skipWhitespace()
		if p.isEOF() {
			break
		}
		
		token := p.peek()
		switch token.Type {
		case lex.TYPE:
			p.next() // consume 'type'
			result.WriteString(p.parseTypeDeclaration())
		case lex.SERVER:
			p.next() // consume 'server'
			result.WriteString(p.parseServerDeclaration())
		default:
			// Ignore unexpected tokens at the top level
			p.next()
		}
	}
	return result.String(), nil
}

func (p *Parser) parseTypeDeclaration() string {
	p.skipWhitespace()
	typeNameToken := p.next()
	if typeNameToken == nil {
		return ""
	}
	typeName := typeNameToken.Chars

	p.skipWhitespace()
	if p.peek() != nil && p.peek().Type == lex.LCB {
		return p.parseInterface(typeName)
	}
	return p.parseTypeAlias(typeName)
}

func (p *Parser) parseInterface(name string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export interface %s {\n", name))
	p.next() // consume LCB '{'

	for !p.isEOF() && p.peek().Type != lex.RCB {
		p.skipWhitespace()
		if p.isEOF() || p.peek().Type == lex.RCB {
			break
		}

		fieldNameToken := p.next()
		fieldName := fieldNameToken.Chars
		
		p.skipWhitespace()
		fieldType := p.parseFieldType()
		sb.WriteString(fmt.Sprintf("  %s: %s;\n", fieldName, fieldType))
	}
	p.next() // consume RCB '}'
	sb.WriteString("}\n\n")
	return sb.String()
}

func (p *Parser) parseTypeAlias(name string) string {
	fieldType := p.parseFieldType()
	return fmt.Sprintf("export type %s = %s;\n\n", name, fieldType)
}

func (p *Parser) parseFieldType() string {
	p.skipWhitespace()
	if p.peek().Type == lex.REPEATED {
		p.next() // consume 'repeated'
		p.next() // consume '('
		p.skipWhitespace()
		baseType := p.mapType(p.next().Chars)
		p.skipWhitespace()
		p.next() // consume ')'
		return baseType + "[]"
	}

	if p.peek().Type == lex.MAP {
		p.next() // consume 'map'
		p.next() // consume '('
		p.skipWhitespace()
		keyType := p.mapType(p.next().Chars)
		p.skipWhitespace()
		p.next() // consume ':'
		p.skipWhitespace()
		valueType := p.mapType(p.next().Chars)
		p.skipWhitespace()
		p.next() // consume ')'
		return fmt.Sprintf("Record<%s, %s>", keyType, valueType)
	}

	return p.mapType(p.next().Chars)
}

func (p *Parser) parseServerDeclaration() string {
	p.skipWhitespace()
	serverName := p.next().Chars
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// API client for %s\n", serverName))
	
    p.skipWhitespace()
    if p.peek() != nil && p.peek().Type == lex.LCB {
	    p.next()
    }

	for !p.isEOF() && p.peek().Type != lex.RCB {
		p.skipWhitespace()
		if p.isEOF() || p.peek().Type == lex.RCB {
			break
		}
		sb.WriteString(p.parseEndpoint())
	}
    if p.peek() != nil && p.peek().Type == lex.RCB {
	    p.next()
    }
	return sb.String()
}

func (p *Parser) parseEndpoint() string {
	p.skipWhitespace()
	name := p.next().Chars
	p.skipWhitespace()
	method := p.next().Chars
	p.skipWhitespace()
	path := p.next().Chars

	bodyType := ""
	returnType := "any"

	for !p.isEOF() && p.peek().Type != lex.NEWLINE && p.peek().Type != lex.RCB {
		p.skipWhitespace()
		if p.isEOF() || p.peek().Type == lex.NEWLINE || p.peek().Type == lex.RCB {
			break
		}
		
		keywordToken := p.peek()
		if keywordToken.Type == lex.BODY {
			p.next() // consume 'body'
			p.next() // consume '('
			p.skipWhitespace()
			bodyType = p.next().Chars
			p.skipWhitespace()
			p.next() // consume ')'
		} else if keywordToken.Type == lex.RETURNS {
			p.next() // consume 'returns'
			p.next() // consume '('
			p.skipWhitespace()
			returnType = p.next().Chars
			p.skipWhitespace()
			p.next() // consume ')'
		} else {
			p.next() // consume unexpected token
		}
	}

	pathParams := []string{}
	cleanPath := path
	if strings.Contains(path, "<") {
		var currentParam strings.Builder
		inParam := false
		for _, r := range path {
			if r == '<' {
				inParam = true
			} else if r == '>' {
				inParam = false
				paramName := currentParam.String()
				pathParams = append(pathParams, paramName)
				cleanPath = strings.Replace(cleanPath, fmt.Sprintf("<%s>", paramName), fmt.Sprintf("${%s}", paramName), 1)
				currentParam.Reset()
			} else if inParam {
				currentParam.WriteRune(r)
			}
		}
	}

	var args []string
	for _, param := range pathParams {
		args = append(args, fmt.Sprintf("%s: string", param))
	}
	if bodyType != "" {
		args = append(args, fmt.Sprintf("body: %s", bodyType))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export const %s = async (%s): Promise<%s> => {\n", name, strings.Join(args, ", "), returnType))
	sb.WriteString(fmt.Sprintf("  const response = await fetch(`%s`, {\n", cleanPath))
	sb.WriteString(fmt.Sprintf("    method: '%s',\n", method))
	if bodyType != "" {
		sb.WriteString("    headers: { 'Content-Type': 'application/json' },\n")
		sb.WriteString("    body: JSON.stringify(body),\n")
	}
	sb.WriteString("  });\n")
	sb.WriteString("  if (!response.ok) {\n")
	sb.WriteString("    throw new Error(`HTTP error! status: ${response.status}`);\n")
	sb.WriteString("  }\n")
	sb.WriteString(fmt.Sprintf("  return response.json() as Promise<%s>;\n", returnType))
	sb.WriteString("};\n\n")

	return sb.String()
}
