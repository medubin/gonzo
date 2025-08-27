package generator_test

import (
	"encoding/json"
	"testing"

	"github.com/medubin/gonzo/api/code_generator/fileio"
	"github.com/medubin/gonzo/api/code_generator/generator"
	"github.com/stretchr/testify/assert"
)

func TestJSONGenerate(t *testing.T) {
	input, err := fileio.ParseFile("test_data/test_server.api")
	assert.NoError(t, err)

	expectedOutput, err := fileio.ParseFile("test_data/json_expected_output.json")
	assert.NoError(t, err)

	parser := generator.NewParser(string(input))
	api, err := parser.Parse()
	assert.NoError(t, err)

	jsonOutput, err := json.MarshalIndent(api, "", "  ")
	assert.NoError(t, err)

	assert.Equal(t, string(expectedOutput), string(jsonOutput))
	println(string(jsonOutput))
}
