package lex

type State string

const (
	STATE_NONE     State = "NONE"
	STATE_TYPE     State = "TYPE"
	STATE_FIELD    State = "FIELD"
	STATE_SERVER   State = "SERVER"
	STATE_REPEATED State = "REPEATED"
	STATE_MAP      State = "MAP"
	STATE_BODY     State = "BODY"
	STATE_RETURNS  State = "RETURNS"
)

type States struct {
	states []State
}

func NewStates() States {
	return States{
		states: []State{STATE_NONE},
	}
}

func (s *States) Get() State {
	if len(s.states) == 0 {
		return STATE_NONE
	}
	return s.states[len(s.states)-1]
}

func (s *States) Pop() State {
	if len(s.states) <= 1 {
		return STATE_NONE
	}
	last := s.Get()
	s.states = s.states[:len(s.states)-1]
	return last
}

func (s *States) Push(state State) {
	s.states = append(s.states, state)
}

var statePushTokens = map[Type]State{
	TYPE:     STATE_TYPE,
	SERVER:   STATE_SERVER,
	REPEATED: STATE_REPEATED,
	MAP:      STATE_MAP,
	BODY:     STATE_BODY,
	RETURNS:  STATE_RETURNS,
}

var statePopTokens = map[Type]State{
	RP:  STATE_MAP,
	RCB: STATE_FIELD,
}

func (s *States) PushOrPopTokenType(tokenType Type) {
	if state, ok := statePushTokens[tokenType]; ok {
		s.Push(state)
		return
	}

	if tokenType == LCB {
		if s.Get() == STATE_TYPE || s.Get() == STATE_FIELD {
			s.Push(STATE_FIELD)
		}
		return
	}

	if state, ok := statePopTokens[tokenType]; ok {
		if s.Get() == state || s.Get() == STATE_REPEATED || s.Get() == STATE_BODY || s.Get() == STATE_RETURNS || s.Get() == STATE_TYPE || s.Get() == STATE_SERVER {
			s.Pop()
		}
		return
	}

	if tokenType == NEWLINE && s.Get() == STATE_TYPE {
		s.Pop()
	}
}
