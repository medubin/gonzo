package generator_test

import (
	"testing"

	"github.com/medubin/gonzo/api/code_generator/fileio"
	"github.com/medubin/gonzo/api/code_generator/generator"
	"github.com/stretchr/testify/assert"
)

func TestCoreGenerate(t *testing.T) {
	input, err := fileio.ParseFile("test_data/test_server.api")
	assert.NoError(t, err)

	parser := generator.NewParser(string(input))
	api, err := parser.Parse()
	assert.NoError(t, err)

	g, err := generator.NewTemplateGenerator("languages/go/server/config.yaml")
	assert.NoError(t, err)
	results, err := g.Generate(api, "testapi")
	for name, result := range results {
		println(name)
		println(result)
	}
	assert.NoError(t, err)
	println(len(results))
	assert.NotNil(t, nil)
}
