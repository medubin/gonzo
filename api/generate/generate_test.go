package generatev2_test

import (
	"testing"

	generate "github.com/medubin/gonzo/api/generate"
	"github.com/medubin/gonzo/api/generate/fileio"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	file, _ := fileio.ParseFile("server.api")
	types, endpoints, err := generate.Generate(file, "server")
	assert.NoError(t, err)
	assert.NotNil(t, types)
	assert.NotNil(t, endpoints)
}
