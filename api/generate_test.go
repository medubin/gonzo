package api_test

import (
	"testing"

	"github.com/medubin/gonzo/api"
	"github.com/stretchr/testify/assert"
)

const expectdOutput = `type UserID string

type User struct {
	Id   UserID
	Name string
}

type TaskRequest struct {
	Message string
	Count   int
	Many    []string
	User    User
	Users   []User
}

`

func TestMain(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		output, err := api.GenerateAPI("test.api")
		println(output)
		assert.NoError(t, err)
		assert.Equal(t, expectdOutput, output)
	})
	t.Run("Nonexistent file", func(t *testing.T) {
		output, err := api.GenerateAPI("bleh.api")
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
