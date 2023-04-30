package generatev2_test

import (
	"testing"

	generatev2 "github.com/medubin/gonzo/api/generate_v2"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	generatev2.Generate("server.api")
	assert.True(t, false)
}
