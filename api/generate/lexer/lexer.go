package lex

import (
	"strings"

	cr "github.com/medubin/gonzo/api/generate/character_reader"
)

type Lexer struct {
	c        *cr.CharacterReader
	Tokens   []Token
	position int
	states   States
}

func NewLexer(c cr.CharacterReader) Lexer {
	return Lexer{
		c:      &c,
		states: NewStates(),
	}
}

func (l *Lexer) Lex() error {
	for !l.c.IsEOF() {
		token := l.nextToken()
		l.Tokens = append(l.Tokens, token)
	}
	return nil
}

func (l *Lexer) Peek() *Token {
	if l.position >= len(l.Tokens) {
		return nil
	}
	return &l.Tokens[l.position]
}

func (l *Lexer) Next() *Token {
	if l.position >= len(l.Tokens) {
		return nil
	}
	head := &l.Tokens[l.position]
	l.position++
	return head
}

func (l *Lexer) IsEOF() bool {
	return l.position >= len(l.Tokens)
}

func (l *Lexer) Reset() {
	l.position = 0
}

var singleCharTokens = map[byte]Type{
	'\t': TAB,
	' ':  SPACE,
	'\n': NEWLINE,
	'\r': NEWLINE,
	';':  NEWLINE,
	':':  COLON,
	'=':  ASSIGN,
	'@':  SIGN,
	'(':  LP,
	'[':  LSB,
	']':  RSB,
	'+':  PLUS,
	'-':  MINUS,
	'*':  TIMES,
	'%':  MOD,
	'^':  POW,
	'>':  GT,
	'<':  LT,
	'!':  EXCL,
	'?':  QM,
	',':  COMMA,
	'#':  HASH,
	'.':  DOT,
}

func (l *Lexer) nextToken() Token {
	char := l.c.Peek()

	if tokenType, ok := singleCharTokens[char]; ok {
		if tokenType == NEWLINE {
			l.states.PushOrPopTokenType(NEWLINE)
		}
		return NewTokenFromByte(l.c.Next(), tokenType, l.states.Get())
	}

	switch char {
	case ')':
		prevState := l.states.Get()
		l.states.PushOrPopTokenType(RP)
		return NewTokenFromByte(l.c.Next(), RP, prevState)
	case '{':
		l.states.PushOrPopTokenType(LCB)
		return NewTokenFromByte(l.c.Next(), LCB, l.states.Get())
	case '}':
		prevState := l.states.Get()
		l.states.PushOrPopTokenType(RCB)
		return NewTokenFromByte(l.c.Next(), RCB, prevState)
	case '/':
		if len(l.c.PeekN(2)) > 1 && l.c.PeekN(2)[1] == '/' {
			return l.lexSingleLineComment()
		} else if len(l.c.PeekN(2)) > 1 && l.c.PeekN(2)[1] == '*' {
			return l.lexMultiLineComment()
		}
		if l.states.Get() == STATE_SERVER {
			return l.lexUrl()
		}
		return NewTokenFromByte(l.c.Next(), FS, l.states.Get())
	case '"':
		return l.lexString()
	default:
		if isLetter(char) {
			ident := l.lexIdent()
			tokenType := LookupKeywords(ident)
			l.states.PushOrPopTokenType(tokenType)
			return NewTokenFromString(ident, tokenType, l.states.Get())
		} else if isDigit(char) {
			s, isDouble := l.lexNumber()
			tokenType := Type(INTEGER)
			if isDouble {
				tokenType = Type(DOUBLE)
			}
			return NewTokenFromString(s, tokenType, l.states.Get())
		}
		return NewTokenFromByte(l.c.Next(), ILLEGAL, l.states.Get())
	}
}

func isNewline(r byte) bool {
	return r == '\r' || r == '\n'
}

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) lexSingleLineComment() Token {
	var sb strings.Builder
	for !l.c.IsEOF() {
		if isNewline(l.c.Peek()) {
			return NewTokenFromString(sb.String(), SINGLE_LINE_COMMENT, l.states.Get())
		}
		sb.WriteByte(l.c.Next())
	}
	return NewTokenFromString(sb.String(), ILLEGAL, l.states.Get())
}

func (l *Lexer) lexMultiLineComment() Token {
	var sb strings.Builder
	for !l.c.IsEOF() {
		nextTwo := l.c.PeekN(2)
		if len(nextTwo) > 1 && nextTwo[0] == '*' && nextTwo[1] == '/' {
			sb.Write(l.c.NextN(2))
			return NewTokenFromString(sb.String(), MULTI_LINE_COMMENT, l.states.Get())
		}
		sb.WriteByte(l.c.Next())
	}
	return NewTokenFromString(sb.String(), ILLEGAL, l.states.Get())
}

func (l *Lexer) lexString() Token {
	var sb strings.Builder
	for !l.c.IsEOF() {
		if l.c.Peek() == '"' {
			sb.WriteByte(l.c.Next())
			return NewTokenFromString(sb.String(), STRING, l.states.Get())
		}
		sb.WriteByte(l.c.Next())
	}
	return NewTokenFromString(sb.String(), ILLEGAL, l.states.Get())
}

func (l *Lexer) lexNumber() (string, bool) {
	var sb strings.Builder
	seenDot := false
	for !l.c.IsEOF() {
		peek := l.c.Peek()
		if peek == '.' && !seenDot {
			seenDot = true
			sb.WriteByte(l.c.Next())
		} else if isDigit(peek) {
			sb.WriteByte(l.c.Next())
		} else {
			break
		}
	}
	return sb.String(), seenDot
}

func (l *Lexer) lexUrl() Token {
	var sb strings.Builder
	for !l.c.IsEOF() {
		peek := l.c.Peek()
		if peek == ' ' || isNewline(peek) {
			break
		}
		sb.WriteByte(l.c.Next())
	}
	return NewTokenFromString(sb.String(), URL, l.states.Get())
}

func (l *Lexer) lexIdent() string {
	var sb strings.Builder
	for !l.c.IsEOF() {
		peek := l.c.Peek()
		if isLetter(peek) || isDigit(peek) {
			sb.WriteByte(l.c.Next())
		} else {
			break
		}
	}
	return sb.String()
}
