package generator_test

import (
	"strings"
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseAPI(t *testing.T, src string) (*generator.APIDefinition, error) {
	t.Helper()
	return generator.NewParser(src).Parse()
}

func TestValidateMapKeys_RejectsRepeatedKey(t *testing.T) {
	_, err := parseAPI(t, `type Bad map(repeated(string): int32)`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a comparable")
	assert.Contains(t, err.Error(), "repeated")
}

func TestValidateMapKeys_RejectsMapKey(t *testing.T) {
	_, err := parseAPI(t, `type Bad map(map(string: int32): int32)`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a comparable")
}

func TestValidateMapKeys_RejectsRepeatedKeyOnField(t *testing.T) {
	_, err := parseAPI(t, `type Bag {
  Tags map(repeated(string): int32)
}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field \"Tags\"")
}

func TestValidateMapKeys_RejectsAliasOfRepeated(t *testing.T) {
	src := `type Tags repeated(string)
type Bad map(Tags: int32)`
	_, err := parseAPI(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a comparable")
}

func TestValidateMapKeys_RejectsStructWithRepeatedField(t *testing.T) {
	src := `type Composite {
  Tags repeated(string)
}
type Bad map(Composite: int32)`
	_, err := parseAPI(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a comparable")
}

func TestValidateMapKeys_AllowsPrimitiveKey(t *testing.T) {
	_, err := parseAPI(t, `type Counts map(string: int32)`)
	assert.NoError(t, err)
}

func TestValidateMapKeys_AllowsAliasKey(t *testing.T) {
	src := `type UserID int64
type Counts map(UserID: int32)`
	_, err := parseAPI(t, src)
	assert.NoError(t, err)
}

func TestValidateMapKeys_AllowsEnumKey(t *testing.T) {
	src := `enum Role string {
  ADMIN = "admin"
  USER = "user"
}
type Counts map(Role: int32)`
	_, err := parseAPI(t, src)
	assert.NoError(t, err)
}

func TestValidateMapKeys_AllowsComparableStructKey(t *testing.T) {
	src := `type Pair {
  A string
  B int32
}
type Counts map(Pair: int32)`
	_, err := parseAPI(t, src)
	assert.NoError(t, err)
}

func TestValidateMapKeys_AllowsNestedMapWithComparableKeys(t *testing.T) {
	_, err := parseAPI(t, `type Nested map(string: map(int64: bool))`)
	assert.NoError(t, err)
}

func TestValidateMapKeys_ErrorMentionsTypeAndOffender(t *testing.T) {
	_, err := parseAPI(t, `type WantsList map(repeated(int32): bool)`)
	require.Error(t, err)
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "WantsList"), "error should name the offending type, got: %s", msg)
}
