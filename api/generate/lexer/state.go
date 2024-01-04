package lex

type State string

const (
	STATE_NONE     = "NONE"
	STATE_TYPE     = "TYPE"
	STATE_FIELD    = "FIELD"
	STATE_SERVER   = "SERVER"
	STATE_REPEATED = "REPEATED"
	STATE_MAP      = "MAP"
	STATE_BODY     = "BODY"
	STATE_RETURNS  = "RETURNS"
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
	return s.states[len(s.states)-1]
}

func (s *States) Pop() State {
	last := s.Get()
	s.states = s.states[:len(s.states)-1]
	return last
}

func (s *States) Push(state State) {
	s.states = append(s.states, state)
}

func (s *States) PushOrPopTokenType(tokenType Type) {
	switch tokenType {
	case TYPE:
		s.Push(STATE_TYPE)
	case SERVER:
		s.Push(STATE_SERVER)
	case LCB:
		if s.Get() == STATE_TYPE || s.Get() == STATE_FIELD {
			s.Push(STATE_FIELD)
		}
	case REPEATED:
		s.Push(STATE_REPEATED)
	case MAP:
		s.Push(STATE_MAP)
	case BODY:
		s.Push(STATE_BODY)
	case RETURNS:
		s.Push(STATE_RETURNS)
	case RP:
		if s.Get() == STATE_MAP || s.Get() == STATE_REPEATED {
			s.Pop()
		}
		if s.Get() == STATE_BODY || s.Get() == STATE_RETURNS {
			s.Pop()
		}
	case RCB:
		if s.Get() == STATE_FIELD || s.Get() == STATE_TYPE || s.Get() == STATE_SERVER {
			s.Pop()
		}
	case NEWLINE:
		// TODO currently breaks if we have a RCB on a new line
		if s.Get() == STATE_TYPE {
			s.Pop()
		}
	}
}
