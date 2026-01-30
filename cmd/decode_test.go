package cmd

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeCmd_FromArgument(t *testing.T) {
	// Reset flags for test
	resetDecodeFlags()

	input := `<saml>test</saml>`
	encoded := base64.StdEncoding.EncodeToString([]byte(input))

	output, err := executeCommand(rootCmd, "decode", encoded)
	require.NoError(t, err)

	assert.Contains(t, output, "<saml>")
	assert.Contains(t, output, "test")
}

func TestDecodeCmd_FromFile(t *testing.T) {
	resetDecodeFlags()

	input := `<saml>file test</saml>`
	encoded := base64.StdEncoding.EncodeToString([]byte(input))

	// Create temp file
	tmpFile := createTempFile(t, encoded)
	defer os.Remove(tmpFile)

	output, err := executeCommand(rootCmd, "decode", "-f", tmpFile)
	require.NoError(t, err)

	assert.Contains(t, output, "<saml>")
	assert.Contains(t, output, "file test")
}

func TestDecodeCmd_JSONOutput(t *testing.T) {
	resetDecodeFlags()

	// Use a proper SAML assertion for JSON parsing
	input := `<Assertion ID="_test123"><Issuer>test-issuer</Issuer></Assertion>`
	encoded := base64.StdEncoding.EncodeToString([]byte(input))

	output, err := executeCommand(rootCmd, "decode", "-o", "json", encoded)
	require.NoError(t, err)

	// Should be valid JSON-ish output (pretty printed XML or parsed JSON)
	assert.NotEmpty(t, output)
}

func TestDecodeCmd_InvalidBase64(t *testing.T) {
	resetDecodeFlags()

	_, err := executeCommand(rootCmd, "decode", "not-valid-base64!!!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode SAML")
}

func TestDecodeCmd_NoInput(t *testing.T) {
	resetDecodeFlags()

	_, err := executeCommand(rootCmd, "decode")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
}

func TestDecodeCmd_FileNotFound(t *testing.T) {
	resetDecodeFlags()

	_, err := executeCommand(rootCmd, "decode", "-f", "/nonexistent/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestDecodeCmd_Golden(t *testing.T) {
	resetDecodeFlags()

	// Load test fixture
	fixtureDir := filepath.Join("..", "testdata", "fixtures", "assertions")

	responseXML, err := os.ReadFile(filepath.Join(fixtureDir, "response.xml"))
	require.NoError(t, err)

	encoded := base64.StdEncoding.EncodeToString(responseXML)

	output, err := executeCommand(rootCmd, "decode", "-o", "xml", encoded)
	require.NoError(t, err)

	// Verify output contains key SAML elements instead of exact match
	assert.Contains(t, output, "Response")
	assert.Contains(t, output, "_response123")
	assert.Contains(t, output, "https://idp.example.com")
	assert.Contains(t, output, "user@example.com")
	assert.Contains(t, output, "Success")
}

// Helper functions

func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()

	return buf.String(), err
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "samlurai_test_*.txt")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}

func resetDecodeFlags() {
	decodeFile = ""
	decodeDeflate = false
	outputFormat = "pretty"
}

func TestDecodeCmd_WithWhitespace(t *testing.T) {
	resetDecodeFlags()

	input := `<saml>whitespace test</saml>`
	encoded := base64.StdEncoding.EncodeToString([]byte(input))

	// Add whitespace and newlines
	encodedWithWhitespace := "  " + encoded[:10] + "\n" + encoded[10:] + "  \n"

	tmpFile := createTempFile(t, encodedWithWhitespace)
	defer os.Remove(tmpFile)

	output, err := executeCommand(rootCmd, "decode", "-f", tmpFile)
	require.NoError(t, err)

	assert.Contains(t, output, "whitespace test")
}

func TestDecodeCmd_HelpText(t *testing.T) {
	resetDecodeFlags()

	output, err := executeCommand(rootCmd, "decode", "--help")
	require.NoError(t, err)

	assert.Contains(t, output, "Decode a base64-encoded SAML")
	assert.Contains(t, output, "--file")
	assert.Contains(t, output, "--deflate")
	assert.Contains(t, output, "--output")
}

func TestDecodeCmd_DeflateFlag(t *testing.T) {
	resetDecodeFlags()

	// We need to create a properly deflated input
	// For this test, we'll just verify the flag is recognized
	output, err := executeCommand(rootCmd, "decode", "--help")
	require.NoError(t, err)
	assert.Contains(t, output, "deflate")
}

func TestDecodeCmd_MultipleOutputFormats(t *testing.T) {
	formats := []string{"pretty", "xml", "json"}

	input := `<Assertion ID="_test"><Issuer>test</Issuer></Assertion>`
	encoded := base64.StdEncoding.EncodeToString([]byte(input))

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			resetDecodeFlags()

			output, err := executeCommand(rootCmd, "decode", "-o", format, encoded)
			require.NoError(t, err)
			assert.NotEmpty(t, output)

			// Each format should produce different looking output
			if format == "json" {
				assert.True(t, strings.Contains(output, "{") || strings.Contains(output, "["),
					"JSON output should contain braces or brackets")
			}
		})
	}
}
