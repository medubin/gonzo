package generatev2

type State int32

type Lexer struct {
	tokens []string
}

func (l *Lexer) Initialize(c CharacterReader) error {
	currentToken := ""

	for !c.IsEOF() {
		currentChar := c.Consume()
		switch currentChar {
		case " ":
			if currentToken != "" {
				l.tokens = append(l.tokens, currentToken)
				currentToken = ""
			}
		case ")":
			fallthrough
		case "(":
			fallthrough
		case "{":
			fallthrough
		case "}":
			fallthrough
		case ":":
			fallthrough
		case "\n":
			if currentToken != "" {
				l.tokens = append(l.tokens, currentToken)
				currentToken = ""
			}
			l.tokens = append(l.tokens, currentChar)
		default:
			currentToken += currentChar
		}
	}

	return nil
}

func (l *Lexer) Peek() string {
	return l.tokens[0]
}

func (l *Lexer) PeekN(n int) []string {
	return l.tokens[0:n]
}

func (l *Lexer) Consume() string {
	consumed := l.Peek()
	l.tokens = l.tokens[1:]
	return consumed
}

func (l *Lexer) ConsumeN(n int) []string {
	consumed := l.PeekN(n)
	l.tokens = l.tokens[n:]
	return consumed
}

func (l *Lexer) IsEOF() bool {
	return len(l.tokens) == 0
}
