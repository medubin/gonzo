package lex

type Token struct {
	Chars string
	Type  Type
	State State
}

// NewTokenFromByte creates a new token from a byte.
func NewTokenFromByte(char byte, typ Type, state State) Token {
	return Token{Chars: string(char), Type: typ, State: state}
}

// NewTokenFromString creates a new token from a string.
func NewTokenFromString(s string, typ Type, state State) Token {
	return Token{Chars: s, Type: typ, State: state}
}
