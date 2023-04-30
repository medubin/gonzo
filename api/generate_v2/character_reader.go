package generatev2

type CharacterReader struct {
	input string
}

func (c *CharacterReader) Initialize(file string) {
	c.input = file
}

func (c *CharacterReader) Peek() string {
	return c.input[0:1]
}

func (c *CharacterReader) PeekN(n int) string {
	return c.input[0:n]
}

func (c *CharacterReader) Consume() string {
	consumed := c.Peek()
	c.input = c.input[1:]
	return consumed
}

func (c *CharacterReader) ConsumeN(n int) string {
	consumed := c.PeekN(n)
	c.input = c.input[n:]
	return consumed
}

func (c *CharacterReader) IsEOF() bool {
	return c.input == ""
}
