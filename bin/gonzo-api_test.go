package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoRoot changes the working directory to the repository root for the
// duration of the test, then restores it. This is needed because config paths
// returned by GetLanguageStackConfig are relative to the repo root.
func repoRoot(t *testing.T) {
	t.Helper()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(".."))
	t.Cleanup(func() { os.Chdir(orig) })
}

// --- root() ---

func TestRoot_NoArgs_ReturnsError(t *testing.T) {
	err := root([]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sub-command")
}

func TestRoot_UnknownSubcommand_ReturnsError(t *testing.T) {
	err := root([]string{"notacommand"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown subcommand")
	assert.Contains(t, err.Error(), "notacommand")
}

func TestRoot_GenerateSubcommand_Recognised(t *testing.T) {
	// Passes the subcommand through to GenerateCommand; missing flags produce a
	// validation error, not "unknown subcommand".
	err := root([]string{"generate"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown subcommand")
}

// --- GenerateCommand ---

func TestGenerateCommand_MissingInput(t *testing.T) {
	cmd := NewGenerateCommand()
	require.NoError(t, cmd.Init([]string{
		"-language", "go",
		"-stack", "server",
		"-output", t.TempDir(),
		"-package", "mypkg",
	}))
	err := cmd.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input required")
}

func TestGenerateCommand_MissingOutput(t *testing.T) {
	cmd := NewGenerateCommand()
	require.NoError(t, cmd.Init([]string{
		"-language", "go",
		"-stack", "server",
		"-input", "somefile.api",
		"-package", "mypkg",
	}))
	err := cmd.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "output required")
}

func TestGenerateCommand_MissingPackage(t *testing.T) {
	cmd := NewGenerateCommand()
	require.NoError(t, cmd.Init([]string{
		"-language", "go",
		"-stack", "server",
		"-input", "somefile.api",
		"-output", t.TempDir(),
	}))
	err := cmd.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "package name required")
}

func TestGenerateCommand_UnsupportedLanguageStack(t *testing.T) {
	cmd := NewGenerateCommand()
	require.NoError(t, cmd.Init([]string{
		"-language", "cobol",
		"-stack", "server",
		"-input", "somefile.api",
		"-output", t.TempDir(),
		"-package", "mypkg",
	}))
	err := cmd.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported language stack combination")
}

func TestGenerateCommand_FullRun_Go(t *testing.T) {
	repoRoot(t)
	cmd := NewGenerateCommand()
	require.NoError(t, cmd.Init([]string{
		"-language", "go",
		"-stack", "server",
		"-input", "code_generator/generator/test_data/test_server.api",
		"-output", t.TempDir(),
		"-package", "server",
	}))
	err := cmd.Run()
	require.NoError(t, err)
}

func TestGenerateCommand_FullRun_TypeScript(t *testing.T) {
	repoRoot(t)
	cmd := NewGenerateCommand()
	require.NoError(t, cmd.Init([]string{
		"-language", "typescript",
		"-stack", "client",
		"-input", "code_generator/generator/test_data/test_server.api",
		"-output", t.TempDir(),
		"-package", "client",
	}))
	err := cmd.Run()
	require.NoError(t, err)
}
