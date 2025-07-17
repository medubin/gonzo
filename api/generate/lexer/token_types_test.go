package lex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupKeywords(t *testing.T) {
	assert.Equal(t, TYPE, LookupKeywords("type"))
	assert.Equal(t, SERVER, LookupKeywords("server"))
	assert.Equal(t, REPEATED, LookupKeywords("repeated"))
	assert.Equal(t, MAP, LookupKeywords("map"))
	assert.Equal(t, RETURNS, LookupKeywords("returns"))
	assert.Equal(t, BODY, LookupKeywords("body"))
	assert.Equal(t, BOOLEAN, LookupKeywords("true"))
	assert.Equal(t, BOOLEAN, LookupKeywords("false"))
	assert.Equal(t, IDENT, LookupKeywords("unknown"))
}
