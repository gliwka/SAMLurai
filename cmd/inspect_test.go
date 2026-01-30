package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspectCmd_FromFile(t *testing.T) {
	resetInspectFlags()

	fixtureDir := filepath.Join("..", "testdata", "fixtures", "assertions")
	responsePath := filepath.Join(fixtureDir, "response.xml")

	output, err := executeCommand(rootCmd, "inspect", "-f", responsePath)
	require.NoError(t, err)

	// Should contain key SAML elements
	assert.Contains(t, output, "Response")
	assert.Contains(t, output, "https://idp.example.com")
	assert.Contains(t, output, "user@example.com")
}

func TestInspectCmd_JSONOutput(t *testing.T) {
	resetInspectFlags()

	fixtureDir := filepath.Join("..", "testdata", "fixtures", "assertions")
	responsePath := filepath.Join(fixtureDir, "response.xml")

	output, err := executeCommand(rootCmd, "inspect", "-f", responsePath, "-o", "json")
	require.NoError(t, err)

	// Should be JSON
	assert.Contains(t, output, "{")
	assert.Contains(t, output, `"type"`)
	assert.Contains(t, output, `"issuer"`)
}

func TestInspectCmd_NoInput(t *testing.T) {
	resetInspectFlags()

	_, err := executeCommand(rootCmd, "inspect")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
}

func TestInspectCmd_FileNotFound(t *testing.T) {
	resetInspectFlags()

	_, err := executeCommand(rootCmd, "inspect", "-f", "/nonexistent/file.xml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestInspectCmd_InvalidXML(t *testing.T) {
	resetInspectFlags()

	tmpFile := createTempFile(t, "not valid XML at all")
	defer os.Remove(tmpFile)

	_, err := executeCommand(rootCmd, "inspect", "-f", tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse SAML")
}

func TestInspectCmd_Golden_Response(t *testing.T) {
	resetInspectFlags()

	fixtureDir := filepath.Join("..", "testdata", "fixtures", "assertions")

	responsePath := filepath.Join(fixtureDir, "response.xml")

	output, err := executeCommand(rootCmd, "inspect", "-f", responsePath, "-o", "json")
	require.NoError(t, err)

	// Verify JSON output contains expected fields
	assert.Contains(t, output, `"type": "Response"`)
	assert.Contains(t, output, `"id": "_response123"`)
	assert.Contains(t, output, `"issuer": "https://idp.example.com"`)
	assert.Contains(t, output, `"status_code": "Success"`)
	assert.Contains(t, output, `"name_id": "user@example.com"`)
}

func TestInspectCmd_Golden_Assertion(t *testing.T) {
	resetInspectFlags()

	fixtureDir := filepath.Join("..", "testdata", "fixtures", "assertions")

	assertionPath := filepath.Join(fixtureDir, "assertion.xml")

	output, err := executeCommand(rootCmd, "inspect", "-f", assertionPath, "-o", "json")
	require.NoError(t, err)

	// Verify JSON output contains expected fields
	assert.Contains(t, output, `"type": "Assertion"`)
	assert.Contains(t, output, `"id": "_assertion789"`)
	assert.Contains(t, output, `"issuer": "https://idp.example.com"`)
	assert.Contains(t, output, `"name_id": "user@example.com"`)
}

func TestInspectCmd_HelpText(t *testing.T) {
	resetInspectFlags()

	output, err := executeCommand(rootCmd, "help", "inspect")
	require.NoError(t, err)

	assert.Contains(t, output, "Inspect")
	assert.Contains(t, output, "--file")
	assert.Contains(t, output, "--output")
}

func resetInspectFlags() {
	inspectFile = ""
	outputFormat = "pretty"
}
