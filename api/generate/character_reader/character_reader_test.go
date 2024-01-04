package cr_test

import (
	"testing"

	cr "github.com/medubin/gonzo/api/generate/character_reader"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	testFile := []byte("abcdefghijk\nlmnopqrstuvwxyz")
	t.Run("Next", func(t *testing.T) {
		charReader := cr.NewCharacterReader(testFile)

		idx := 0
		for !charReader.IsEOF() {
			assert.Equal(t, testFile[idx], charReader.Next())
			idx++
		}
	})
}
