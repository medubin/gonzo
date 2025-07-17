package cr_test

import (
	"testing"

	cr "github.com/medubin/gonzo/api/generate/character_reader"
	"github.com/stretchr/testify/assert"
)

func TestCharacterReader(t *testing.T) {
	testFile := []byte("abcdefghijk\nlmnopqrstuvwxyz")
	
	t.Run("Next", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)

		idx := 0
		for !charReader.IsEOF() {
			assert.Equal(t, testFile[idx], charReader.Next())
			idx++
		}
	})

	t.Run("Peek", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)
		
		// Peek should not advance position
		first := charReader.Peek()
		assert.Equal(t, testFile[0], first)
		assert.Equal(t, 0, charReader.Pos())
		
		// Peek again should return same value
		assert.Equal(t, first, charReader.Peek())
		assert.Equal(t, 0, charReader.Pos())
	})

	t.Run("PeekN", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)
		
		// Normal case
		peeked := charReader.PeekN(5)
		expected := testFile[:5]
		assert.Equal(t, expected, peeked)
		assert.Equal(t, 0, charReader.Pos()) // Position should not change
		
		// Edge case: peek more than available
		charReader.Skip(len(testFile) - 3) // Move near end
		peeked = charReader.PeekN(10)
		expected = testFile[len(testFile)-3:]
		assert.Equal(t, expected, peeked)
		
		// Edge case: peek 0 or negative
		assert.Equal(t, []byte{}, charReader.PeekN(0))
		assert.Equal(t, []byte{}, charReader.PeekN(-1))
	})

	t.Run("NextN", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)
		
		// Normal case
		consumed := charReader.NextN(5)
		expected := testFile[:5]
		assert.Equal(t, expected, consumed)
		assert.Equal(t, 5, charReader.Pos())
		
		// Edge case: consume more than available
		charReader.Reset()
		charReader.Skip(len(testFile) - 3)
		consumed = charReader.NextN(10)
		expected = testFile[len(testFile)-3:]
		assert.Equal(t, expected, consumed)
		assert.True(t, charReader.IsEOF())
	})

	t.Run("Position and EOF", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)
		
		assert.Equal(t, 0, charReader.Pos())
		assert.False(t, charReader.IsEOF())
		assert.Equal(t, len(testFile), charReader.Remaining())
		assert.Equal(t, len(testFile), charReader.Len())
		
		// Move to end
		charReader.Skip(len(testFile))
		assert.Equal(t, len(testFile), charReader.Pos())
		assert.True(t, charReader.IsEOF())
		assert.Equal(t, 0, charReader.Remaining())
	})

	t.Run("Skip", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)
		
		// Normal skip
		skipped := charReader.Skip(5)
		assert.Equal(t, 5, skipped)
		assert.Equal(t, 5, charReader.Pos())
		
		// Skip more than available
		remaining := charReader.Remaining()
		skipped = charReader.Skip(remaining + 10)
		assert.Equal(t, remaining, skipped)
		assert.True(t, charReader.IsEOF())
		
		// Skip 0 or negative
		charReader.Reset()
		assert.Equal(t, 0, charReader.Skip(0))
		assert.Equal(t, 0, charReader.Skip(-5))
		assert.Equal(t, 0, charReader.Pos())
	})

	t.Run("Reset", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)
		
		charReader.Skip(10)
		assert.Equal(t, 10, charReader.Pos())
		
		charReader.Reset()
		assert.Equal(t, 0, charReader.Pos())
		assert.False(t, charReader.IsEOF())
	})

	t.Run("Seek", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)
		
		// Valid seek
		err := charReader.Seek(10)
		assert.NoError(t, err)
		assert.Equal(t, 10, charReader.Pos())
		
		// Seek to end
		err = charReader.Seek(len(testFile))
		assert.NoError(t, err)
		assert.True(t, charReader.IsEOF())
		
		// Invalid seeks
		err = charReader.Seek(-1)
		assert.Error(t, err)
		assert.Equal(t, cr.ErrOutOfBounds, err)
		
		err = charReader.Seek(len(testFile) + 1)
		assert.Error(t, err)
		assert.Equal(t, cr.ErrOutOfBounds, err)
	})

	t.Run("Empty input", func(t *testing.T) {
		charReader := cr.NewCharacterReader([]byte{})
		
		assert.True(t, charReader.IsEOF())
		assert.Equal(t, 0, charReader.Remaining())
		assert.Equal(t, 0, charReader.Len())
		assert.Equal(t, byte(0), charReader.Peek())
		assert.Equal(t, []byte{}, charReader.PeekN(5))
		assert.Equal(t, byte(0), charReader.Next())
		assert.Equal(t, []byte{}, charReader.NextN(5))
	})
}
