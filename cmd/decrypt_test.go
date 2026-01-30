package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecryptCmd_NoKey(t *testing.T) {
	resetDecryptFlags()

	_, err := executeCommand(rootCmd, "decrypt")
	assert.Error(t, err)
	// Should require the key flag
}

func TestDecryptCmd_KeyFileNotFound(t *testing.T) {
	resetDecryptFlags()

	tmpFile := createTempFile(t, "<EncryptedAssertion>test</EncryptedAssertion>")
	defer os.Remove(tmpFile)

	_, err := executeCommand(rootCmd, "decrypt", "-k", "/nonexistent/key.pem", "-f", tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load private key")
}

func TestDecryptCmd_InvalidKey(t *testing.T) {
	resetDecryptFlags()

	// Create an invalid key file
	keyFile := createTempFile(t, "not a valid PEM key")
	defer os.Remove(keyFile)

	inputFile := createTempFile(t, "<EncryptedAssertion>test</EncryptedAssertion>")
	defer os.Remove(inputFile)

	_, err := executeCommand(rootCmd, "decrypt", "-k", keyFile, "-f", inputFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load private key")
}

func TestDecryptCmd_NoInput(t *testing.T) {
	resetDecryptFlags()

	keyFile := createTestKeyFile(t)
	defer os.Remove(keyFile)

	_, err := executeCommand(rootCmd, "decrypt", "-k", keyFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
}

func TestDecryptCmd_HelpText(t *testing.T) {
	resetDecryptFlags()

	output, err := executeCommand(rootCmd, "help", "decrypt")
	require.NoError(t, err)

	assert.Contains(t, output, "Decrypt")
	assert.Contains(t, output, "--key")
	assert.Contains(t, output, "--file")
	assert.Contains(t, output, "private key")
	assert.Contains(t, output, "PEM")
}

func resetDecryptFlags() {
	decryptFile = ""
	decryptKeyFile = ""
	outputFormat = "pretty"
}

// createTestKeyFile creates a temporary RSA private key file for testing
func createTestKeyFile(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key.pem")

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	err = os.WriteFile(keyPath, pemData, 0600)
	require.NoError(t, err)

	return keyPath
}
