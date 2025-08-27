package cr

import (
	"errors"
)

var (
	ErrOutOfBounds = errors.New("position out of bounds")
)

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
	if n <= 0 || c.IsEOF() {
		return []byte{}
	}
	end := c.position + n
	if end > len(c.input) {
		end = len(c.input)
	}
	return c.input[c.position:end]
}

func (c *CharacterReader) Next() byte {
	consumed := c.Peek()
	c.position++
	return consumed
}

func (c *CharacterReader) NextN(n int) []byte {
	consumed := c.PeekN(n)
	actualN := len(consumed)
	c.position += actualN
	return consumed
}

func (c *CharacterReader) Pos() int {
	return c.position
}

func (c *CharacterReader) IsEOF() bool {
	return c.position >= len(c.input)
}

// Remaining returns the number of bytes left to read
func (c *CharacterReader) Remaining() int {
	remaining := len(c.input) - c.position
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Skip advances the position by n bytes without returning the data
func (c *CharacterReader) Skip(n int) int {
	if n <= 0 {
		return 0
	}
	maxSkip := c.Remaining()
	if n > maxSkip {
		n = maxSkip
	}
	c.position += n
	return n
}

// Reset resets the position to the beginning
func (c *CharacterReader) Reset() {
	c.position = 0
}

// Seek sets the position to the specified offset
func (c *CharacterReader) Seek(offset int) error {
	if offset < 0 || offset > len(c.input) {
		return ErrOutOfBounds
	}
	c.position = offset
	return nil
}

// Len returns the total length of the input
func (c *CharacterReader) Len() int {
	return len(c.input)
}
