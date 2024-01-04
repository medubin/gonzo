package lex

type Token struct {
	Chars string
	Type  Type
	State State
}

// Supports passing in a byte or a string. Give me usable union types Golang!
func NewToken(char any, typ Type, state State) Token {
	if b, ok := char.(byte); ok {
		return Token{Chars: string(b), Type: typ, State: state}
	}
	if s, ok := char.(string); ok {
		return Token{Chars: s, Type: typ, State: state}
	}
	panic("what the hell is this?")
}
