package saml

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

func TestNewDecryptor_ValidPKCS1Key(t *testing.T) {
	keyPath := createTestKey(t, "RSA PRIVATE KEY")
	defer os.Remove(keyPath)

	decryptor, err := NewDecryptor(keyPath)
	require.NoError(t, err)
	assert.NotNil(t, decryptor)
	assert.NotNil(t, decryptor.privateKey)
}

func TestNewDecryptor_ValidPKCS8Key(t *testing.T) {
	keyPath := createTestKeyPKCS8(t)
	defer os.Remove(keyPath)

	decryptor, err := NewDecryptor(keyPath)
	require.NoError(t, err)
	assert.NotNil(t, decryptor)
	assert.NotNil(t, decryptor.privateKey)
}

func TestNewDecryptor_InvalidKeyPath(t *testing.T) {
	_, err := NewDecryptor("/nonexistent/path/to/key.pem")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read private key file")
}

func TestNewDecryptor_InvalidPEM(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("not a valid PEM file")
	require.NoError(t, err)
	tmpFile.Close()

	_, err = NewDecryptor(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse PEM block")
}

func TestNewDecryptor_UnsupportedKeyType(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unsupported_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write a PEM with unsupported type
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "UNKNOWN KEY TYPE",
		Bytes: []byte("fake key data"),
	})
	_, err = tmpFile.Write(pemData)
	require.NoError(t, err)
	tmpFile.Close()

	_, err = NewDecryptor(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported key type")
}

func TestNewDecryptorFromPEM(t *testing.T) {
	// Generate a test key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	decryptor, err := NewDecryptorFromPEM(pemData)
	require.NoError(t, err)
	assert.NotNil(t, decryptor)
}

// Helper function to create a test RSA private key file (PKCS#1)
func createTestKey(t *testing.T, pemType string) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key.pem")

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  pemType,
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	err = os.WriteFile(keyPath, pemData, 0600)
	require.NoError(t, err)

	return keyPath
}

// Helper function to create a test RSA private key file (PKCS#8)
func createTestKeyPKCS8(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key_pkcs8.pem")

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	err = os.WriteFile(keyPath, pemData, 0600)
	require.NoError(t, err)

	return keyPath
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		expected bool
	}{
		{
			name: "encrypted assertion",
			xml: `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">
				<saml:EncryptedAssertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
					<xenc:EncryptedData xmlns:xenc="http://www.w3.org/2001/04/xmlenc#"/>
				</saml:EncryptedAssertion>
			</samlp:Response>`,
			expected: true,
		},
		{
			name: "encrypted data only",
			xml: `<root>
				<xenc:EncryptedData xmlns:xenc="http://www.w3.org/2001/04/xmlenc#"/>
			</root>`,
			expected: true,
		},
		{
			name: "unencrypted assertion",
			xml: `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">
				<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"/>
			</samlp:Response>`,
			expected: false,
		},
		{
			name: "plain assertion",
			xml: `<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
				<saml:Issuer>https://idp.example.com</saml:Issuer>
			</saml:Assertion>`,
			expected: false,
		},
		{
			name:     "invalid XML",
			xml:      "not valid xml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEncrypted([]byte(tt.xml))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsEncryptedString(t *testing.T) {
	encrypted := `<saml:EncryptedAssertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
		<xenc:EncryptedData xmlns:xenc="http://www.w3.org/2001/04/xmlenc#"/>
	</saml:EncryptedAssertion>`

	assert.True(t, IsEncryptedString(encrypted))

	unencrypted := `<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"/>`
	assert.False(t, IsEncryptedString(unencrypted))
}
