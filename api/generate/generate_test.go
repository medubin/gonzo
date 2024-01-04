package generatev2_test

import (
	"testing"

	generate "github.com/medubin/gonzo/api/generate"
	"github.com/medubin/gonzo/api/generate/fileio"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	file, _ := fileio.ParseFile("server.api")
	generate.Generate(file)
	assert.True(t, false)
}
