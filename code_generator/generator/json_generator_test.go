package generator_test

import (
	"encoding/json"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/medubin/gonzo/code_generator/fileio"
	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/require"
)

func TestJSONGenerate(t *testing.T) {
	input, err := fileio.ParseFile("test_data/test_server.api")
	require.NoError(t, err)

	parser := generator.NewParser(string(input))
	api, err := parser.Parse()
	require.NoError(t, err)

	jsonOutput, err := json.MarshalIndent(api, "", "  ")
	require.NoError(t, err)

	// Use JSON snapshot for better readability and diffs
	snaps.MatchJSON(t, jsonOutput)
}
