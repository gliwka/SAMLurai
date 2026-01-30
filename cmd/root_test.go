package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Help(t *testing.T) {
	output, err := executeCommand(rootCmd, "--help")
	require.NoError(t, err)

	assert.Contains(t, output, "SAMLurai")
	assert.Contains(t, output, "decoding")
	assert.Contains(t, output, "decrypting")
	assert.Contains(t, output, "decode")
	assert.Contains(t, output, "decrypt")
	assert.Contains(t, output, "inspect")
}

func TestRootCmd_Version(t *testing.T) {
	output, err := executeCommand(rootCmd, "--version")
	require.NoError(t, err)

	assert.Contains(t, output, "samlurai")
}

func TestRootCmd_UnknownCommand(t *testing.T) {
	_, err := executeCommand(rootCmd, "unknowncommand")
	assert.Error(t, err)
}

func TestRootCmd_GlobalOutputFlag(t *testing.T) {
	output, err := executeCommand(rootCmd, "--help")
	require.NoError(t, err)

	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "pretty")
	assert.Contains(t, output, "json")
	assert.Contains(t, output, "xml")
}

func TestRootCmd_SubcommandsList(t *testing.T) {
	output, err := executeCommand(rootCmd, "--help")
	require.NoError(t, err)

	// All subcommands should be listed
	subcommands := []string{"decode", "decrypt", "inspect"}
	for _, cmd := range subcommands {
		assert.Contains(t, output, cmd, "Subcommand %s should be listed in help", cmd)
	}
}

func TestRootCmd_Examples(t *testing.T) {
	output, err := executeCommand(rootCmd, "--help")
	require.NoError(t, err)

	// Should contain example usage
	assert.Contains(t, output, "Examples:")
}
