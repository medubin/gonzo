package cr

type CharacterReader struct {
	input    []byte
	position int
}

func NewCharacterReader(file []byte) CharacterReader {
	return CharacterReader{
		input: file,
	}
}

func (c *CharacterReader) Peek() byte {
	if c.IsEOF() {
		return 0
	}
	return c.input[c.position]
}

func (c *CharacterReader) PeekN(n int) []byte {
	return c.input[c.position : c.position+n]
}

func (c *CharacterReader) Next() byte {
	consumed := c.Peek()
	c.position++
	return consumed
}

func (c *CharacterReader) NextN(n int) []byte {
	consumed := c.PeekN(n)
	c.position += n
	return consumed
}

func (c *CharacterReader) Pos() int {
	return c.position
}

func (c *CharacterReader) IsEOF() bool {
	return c.position >= len(c.input)
}
