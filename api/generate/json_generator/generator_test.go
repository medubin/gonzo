package jsongenerator_test

import (
	"encoding/json"
	"testing"

	"github.com/medubin/gonzo/api/generate/fileio"
	jsongenerator "github.com/medubin/gonzo/api/generate/json_generator"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	input, err := fileio.ParseFile("test_server.api")
	assert.NoError(t, err)

	expectedOutput, err := fileio.ParseFile("expected_output.json")
	assert.NoError(t, err)

	parser := jsongenerator.NewParser(string(input))
	api, err := parser.Parse()
	assert.NoError(t, err)

	jsonOutput, err := json.MarshalIndent(api, "", "  ")
	assert.NoError(t, err)

	assert.Equal(t, string(expectedOutput), string(jsonOutput))
	println(string(jsonOutput))
}
