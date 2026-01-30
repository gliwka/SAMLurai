package saml

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/beevik/etree"
	"github.com/crewjam/saml/xmlenc"
)

// Decryptor handles decryption of encrypted SAML assertions
type Decryptor struct {
	privateKey *rsa.PrivateKey
}

// NewDecryptor creates a new Decryptor with the given private key file
func NewDecryptor(keyPath string) (*Decryptor, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	return NewDecryptorFromPEM(keyData)
}

// NewDecryptorFromPEM creates a new Decryptor from PEM-encoded key data
func NewDecryptorFromPEM(pemData []byte) (*Decryptor, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	var privateKey *rsa.PrivateKey
	var err error

	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", parseErr)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not an RSA key")
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Decryptor{
		privateKey: privateKey,
	}, nil
}

// Decrypt decrypts an encrypted SAML assertion
func (d *Decryptor) Decrypt(encryptedXML []byte) ([]byte, error) {
	// Parse the XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(encryptedXML); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Find the EncryptedData element
	encryptedDataEl := doc.FindElement("//EncryptedData")
	if encryptedDataEl == nil {
		// Try to find EncryptedAssertion containing EncryptedData
		encryptedAssertionEl := doc.FindElement("//EncryptedAssertion")
		if encryptedAssertionEl != nil {
			encryptedDataEl = encryptedAssertionEl.FindElement("EncryptedData")
		}
	}

	if encryptedDataEl == nil {
		return nil, fmt.Errorf("no EncryptedData element found in XML")
	}

	// Decrypt the element
	decrypted, err := xmlenc.Decrypt(d.privateKey, encryptedDataEl)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return decrypted, nil
}

// DecryptString decrypts an encrypted SAML assertion from a string
func (d *Decryptor) DecryptString(encryptedXML string) ([]byte, error) {
	return d.Decrypt([]byte(encryptedXML))
}

// IsEncrypted checks if the given XML contains encrypted SAML data
func IsEncrypted(xmlData []byte) bool {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(xmlData); err != nil {
		return false
	}

	// Check for EncryptedAssertion or EncryptedData elements
	if doc.FindElement("//EncryptedAssertion") != nil {
		return true
	}
	if doc.FindElement("//EncryptedData") != nil {
		return true
	}

	return false
}

// IsEncryptedString checks if the given XML string contains encrypted SAML data
func IsEncryptedString(xmlData string) bool {
	return IsEncrypted([]byte(xmlData))
}
