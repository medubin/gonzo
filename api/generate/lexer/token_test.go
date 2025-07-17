package lex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	tokenByte := NewTokenFromByte('a', IDENT, STATE_NONE)
	assert.Equal(t, "a", tokenByte.Chars)
	assert.Equal(t, IDENT, tokenByte.Type)
	assert.Equal(t, STATE_NONE, tokenByte.State)

	tokenString := NewTokenFromString("abc", IDENT, STATE_NONE)
	assert.Equal(t, "abc", tokenString.Chars)
	assert.Equal(t, IDENT, tokenString.Type)
	assert.Equal(t, STATE_NONE, tokenString.State)
}
