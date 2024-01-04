package lex

import (
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
	return &l.Tokens[l.position]
}

func (l *Lexer) Next() *Token {
	head := l.Peek()
	l.position++
	return head
}

func (l *Lexer) IsEOF() bool {
	return l.position >= len(l.Tokens)
}

func (l *Lexer) Reset() {
	l.position = 0
}

func (l *Lexer) nextToken() Token {
	char := l.c.Peek()
	switch char {
	case '\t':
		return NewToken(l.c.Next(), TAB, l.states.Get())
	case ' ':
		return NewToken(l.c.Next(), SPACE, l.states.Get())
	case '\n':
		l.states.PushOrPopTokenType(NEWLINE)
		return NewToken(l.c.Next(), NEWLINE, l.states.Get())
	case '\r':
		l.states.PushOrPopTokenType(NEWLINE)
		return NewToken(l.c.Next(), NEWLINE, l.states.Get())
	case ';':
		return NewToken(l.c.Next(), NEWLINE, l.states.Get())
	case ':':
		return NewToken(l.c.Next(), COLON, l.states.Get())
	case '=':
		return NewToken(l.c.Next(), ASSIGN, l.states.Get())
	case '@':
		return NewToken(l.c.Next(), SIGN, l.states.Get())
	case '(':
		return NewToken(l.c.Next(), LP, l.states.Get())
	case ')':
		prevState := l.states.Get()
		l.states.PushOrPopTokenType(RP)
		return NewToken(l.c.Next(), RP, prevState)
	case '{':
		l.states.PushOrPopTokenType(LCB)
		return NewToken(l.c.Next(), LCB, l.states.Get())
	case '}':
		prevState := l.states.Get()
		l.states.PushOrPopTokenType(RCB)
		return NewToken(l.c.Next(), RCB, prevState)
	case '[':
		return NewToken(l.c.Next(), LSB, l.states.Get())
	case ']':
		return NewToken(l.c.Next(), RSB, l.states.Get())
	case '+':
		return NewToken(l.c.Next(), PLUS, l.states.Get())
	case '-':
		return NewToken(l.c.Next(), MINUS, l.states.Get())
	case '*':
		return NewToken(l.c.Next(), TIMES, l.states.Get())
	case '%':
		return NewToken(l.c.Next(), MOD, l.states.Get())
	case '^':
		return NewToken(l.c.Next(), POW, l.states.Get())
	case '>':
		return NewToken(l.c.Next(), GT, l.states.Get())
	case '<':
		return NewToken(l.c.Next(), LT, l.states.Get())
	case '!':
		return NewToken(l.c.Next(), EXCL, l.states.Get())
	case '?':
		return NewToken(l.c.Next(), QM, l.states.Get())
	case ',':
		return NewToken(l.c.Next(), COMMA, l.states.Get())
	case '#':
		return NewToken(l.c.Next(), HASH, l.states.Get())
	case '.':
		return NewToken(l.c.Next(), DOT, l.states.Get())
	case '/':
		chars := l.c.PeekN(2)
		if chars[1] == '/' {
			return l.lexSingleLineComment()
		} else if chars[1] == '*' {
			return l.lexMultiLineComment()
		} else {
			return NewToken(l.c.Next(), FS, l.states.Get())
		}
	case '"':
		return l.lexString()
	default:
		if isLetter(l.c.Peek()) {
			s := l.lexIdent()
			tokenType := LookupKeywords(s)

			l.states.PushOrPopTokenType(tokenType)

			return NewToken(s, tokenType, l.states.Get())
		} else if isDigit(l.c.Peek()) {
			s, isDouble := l.lexNumber()
			if isDouble {
				return NewToken(s, DOUBLE, l.states.Get())
			} else {
				return NewToken(s, INTEGER, l.states.Get())
			}
		} else {
			return NewToken(l.c.Next(), ILLEGAL, l.states.Get())
		}
	}
	// This code should always remain unreachable
	// panic(fmt.Sprintf("Illegal character? %s", string(char)))
}

func isNewline(r byte) bool {
	return r == '\r' || r == '\n'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) lexSingleLineComment() Token {
	s := ""
	for !l.c.IsEOF() {
		if isNewline(l.c.Peek()) {
			return NewToken(s, SINGLE_LINE_COMMENT, l.states.Get())
		}
		s += string(l.c.Next())
	}
	return NewToken(s, ILLEGAL, l.states.Get())
}

func (l *Lexer) lexMultiLineComment() Token {
	s := ""
	for !l.c.IsEOF() {
		nextTwo := l.c.PeekN(2)
		if nextTwo[0] == '*' && nextTwo[1] == '/' {
			s += string(l.c.NextN(2))
			return NewToken(s, MULTI_LINE_COMMENT, l.states.Get())
		}
		s += string(l.c.Next())
	}
	return NewToken(s, ILLEGAL, l.states.Get())
}

func (l *Lexer) lexString() Token {
	s := ""
	for !l.c.IsEOF() {
		if l.c.Peek() == '"' {
			s += string(l.c.Next())
			return NewToken(s, STRING, l.states.Get())
		}
		s += string(l.c.Next())
	}
	return NewToken(s, ILLEGAL, l.states.Get())
}

// lexNumber - reads and returns a number.
func (l *Lexer) lexNumber() (string, bool) {
	s := ""
	seenDot := false
	for isDigit(l.c.Peek()) || (!seenDot && l.c.Peek() == '.') {
		if l.c.Peek() == '.' {
			seenDot = true
		}
		s += string(l.c.Next())
	}
	return s, seenDot
}

// An identifier is a sequence of letters (upper and lowercase) digits and underscores.
// it can't start with a digit
func (l *Lexer) lexIdent() string {
	s := ""
	for isLetter(l.c.Peek()) || isDigit(l.c.Peek()) {
		s += string(l.c.Next())
	}

	return s
}
