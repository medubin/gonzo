package lex

import (
	"testing"

	cr "github.com/medubin/gonzo/api/generate/character_reader"
	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	input := `
// this is a comment
type UserID int32
`
	c := cr.NewCharacterReader([]byte(input))
	l := NewLexer(c)
	l.Lex()

	expectedTokens := []Token{
		NewTokenFromByte('\n', NEWLINE, STATE_NONE),
		NewTokenFromString("// this is a comment", SINGLE_LINE_COMMENT, STATE_NONE),
		NewTokenFromByte('\n', NEWLINE, STATE_NONE),
		NewTokenFromString("type", TYPE, STATE_TYPE),
		NewTokenFromByte(' ', SPACE, STATE_TYPE),
		NewTokenFromString("UserID", IDENT, STATE_TYPE),
		NewTokenFromByte(' ', SPACE, STATE_TYPE),
		NewTokenFromString("int32", IDENT, STATE_TYPE),
		NewTokenFromByte('\n', NEWLINE, STATE_NONE),
	}

	assert.Equal(t, len(expectedTokens), len(l.Tokens))
	for i, token := range l.Tokens {
		assert.Equal(t, expectedTokens[i].Type, token.Type)
		assert.Equal(t, expectedTokens[i].Chars, token.Chars)
		assert.Equal(t, expectedTokens[i].State, token.State)
	}
}
