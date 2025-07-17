package lex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStates(t *testing.T) {
	states := NewStates()
	assert.Equal(t, STATE_NONE, states.Get())

	states.Push(STATE_TYPE)
	assert.Equal(t, STATE_TYPE, states.Get())

	states.Push(STATE_FIELD)
	assert.Equal(t, STATE_FIELD, states.Get())

	assert.Equal(t, STATE_FIELD, states.Pop())
	assert.Equal(t, STATE_TYPE, states.Get())

	assert.Equal(t, STATE_TYPE, states.Pop())
	assert.Equal(t, STATE_NONE, states.Get())

	assert.Equal(t, STATE_NONE, states.Pop())
	assert.Equal(t, STATE_NONE, states.Get())
}

func TestPushOrPop(t *testing.T) {
	states := NewStates()
	states.PushOrPopTokenType(TYPE)
	assert.Equal(t, STATE_TYPE, states.Get())

	states.PushOrPopTokenType(LCB)
	assert.Equal(t, STATE_FIELD, states.Get())

	states.PushOrPopTokenType(RCB)
	assert.Equal(t, STATE_TYPE, states.Get())
}
