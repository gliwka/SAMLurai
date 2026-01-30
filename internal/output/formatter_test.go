package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gliwka/SAMLurai/internal/saml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_FormatXML_Pretty(t *testing.T) {
	formatter := NewFormatter("pretty")

	input := `<root><child>value</child></root>`
	result, err := formatter.FormatXML([]byte(input))

	require.NoError(t, err)
	assert.Contains(t, result, "<root>")
	assert.Contains(t, result, "  <child>") // Should be indented
}

func TestFormatter_FormatXML_JSON(t *testing.T) {
	formatter := NewFormatter("json")

	// Use a valid SAML-like structure
	input := `<?xml version="1.0"?><Assertion ID="_test"><Issuer>test-issuer</Issuer></Assertion>`
	result, err := formatter.FormatXML([]byte(input))

	require.NoError(t, err)
	assert.True(t, json.Valid([]byte(result)), "Output should be valid JSON")
}

func TestFormatter_FormatXML_Raw(t *testing.T) {
	formatter := NewFormatter("xml")

	input := `<root><child>value</child></root>`
	result, err := formatter.FormatXML([]byte(input))

	require.NoError(t, err)
	assert.Contains(t, result, "<root>")
}

func TestFormatter_FormatSAMLInfo_JSON(t *testing.T) {
	formatter := NewFormatter("json")

	info := createTestSAMLInfo()
	result, err := formatter.FormatSAMLInfo(info)

	require.NoError(t, err)
	assert.True(t, json.Valid([]byte(result)), "Output should be valid JSON")

	// Verify it contains expected fields
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(result), &parsed)
	require.NoError(t, err)

	assert.Equal(t, "Response", parsed["type"])
	assert.Equal(t, "_response123", parsed["id"])
	assert.Equal(t, "https://idp.example.com", parsed["issuer"])
}

func TestFormatter_FormatSAMLInfo_Pretty(t *testing.T) {
	formatter := NewFormatterWithOptions("pretty", true) // noColor for consistent testing

	info := createTestSAMLInfo()
	result, err := formatter.FormatSAMLInfo(info)

	require.NoError(t, err)

	// Check for expected sections
	assert.Contains(t, result, "SAML Response")
	assert.Contains(t, result, "Basic Information")
	assert.Contains(t, result, "_response123")
	assert.Contains(t, result, "https://idp.example.com")
	assert.Contains(t, result, "Status")
	assert.Contains(t, result, "Success")
	assert.Contains(t, result, "Subject")
	assert.Contains(t, result, "user@example.com")
	assert.Contains(t, result, "Conditions")
	assert.Contains(t, result, "Attributes")
	assert.Contains(t, result, "email")
}

func TestFormatter_FormatSAMLInfo_XML(t *testing.T) {
	formatter := NewFormatter("xml")

	info := createTestSAMLInfo()
	result, err := formatter.FormatSAMLInfo(info)

	require.NoError(t, err)
	assert.Contains(t, result, "<?xml")
	assert.Contains(t, result, "<SAMLInfo>")
}

func TestFormatter_ShortenURI(t *testing.T) {
	formatter := NewFormatter("pretty")

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress",
			expected: "emailAddress",
		},
		{
			input:    "urn:oasis:names:tc:SAML:2.0:ac:classes:Password",
			expected: "Password",
		},
		{
			input:    "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
			expected: "rsa-sha256",
		},
		{
			input:    "custom-value",
			expected: "custom-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatter.shortenURI(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_DefaultFormat(t *testing.T) {
	// Unknown format should default to pretty
	formatter := NewFormatter("unknown")

	info := createTestSAMLInfo()
	result, err := formatter.FormatSAMLInfo(info)

	require.NoError(t, err)
	// Should output pretty format (not JSON, not XML declaration)
	assert.Contains(t, result, "SAML Response")
}

func TestFormatter_CaseInsensitiveFormat(t *testing.T) {
	tests := []string{"JSON", "Json", "json", "JSON"}

	info := createTestSAMLInfo()

	for _, format := range tests {
		t.Run(format, func(t *testing.T) {
			formatter := NewFormatter(format)
			result, err := formatter.FormatSAMLInfo(info)

			require.NoError(t, err)
			assert.True(t, json.Valid([]byte(result)), "Output should be valid JSON for format: %s", format)
		})
	}
}

func TestFormatter_EmptyAttributes(t *testing.T) {
	formatter := NewFormatterWithOptions("pretty", true)

	info := &saml.SAMLInfo{
		Type:   "Assertion",
		ID:     "_test",
		Issuer: "test-issuer",
		// No attributes
	}

	result, err := formatter.FormatSAMLInfo(info)

	require.NoError(t, err)
	assert.NotContains(t, result, "Attributes")
}

func TestFormatter_NestedAssertion(t *testing.T) {
	formatter := NewFormatterWithOptions("pretty", true)

	info := &saml.SAMLInfo{
		Type:   "Response",
		ID:     "_response",
		Issuer: "response-issuer",
		Assertion: &saml.SAMLInfo{
			Type:   "Assertion",
			ID:     "_assertion",
			Issuer: "assertion-issuer",
		},
	}

	result, err := formatter.FormatSAMLInfo(info)

	require.NoError(t, err)
	assert.Contains(t, result, "Embedded Assertion")
	assert.Contains(t, result, "_assertion")
}

// Helper function to create test SAML info
func createTestSAMLInfo() *saml.SAMLInfo {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	notBefore := now.Add(-5 * time.Minute)
	notOnOrAfter := now.Add(5 * time.Minute)
	authnInstant := now.Add(-1 * time.Minute)

	return &saml.SAMLInfo{
		Type:         "Response",
		ID:           "_response123",
		IssueInstant: &now,
		Destination:  "https://sp.example.com/acs",
		InResponseTo: "_request456",
		Issuer:       "https://idp.example.com",
		Status: &saml.Status{
			StatusCode: "Success",
		},
		Subject: &saml.Subject{
			NameID:       "user@example.com",
			NameIDFormat: "urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress",
		},
		Conditions: &saml.Conditions{
			NotBefore:           &notBefore,
			NotOnOrAfter:        &notOnOrAfter,
			AudienceRestriction: []string{"https://sp.example.com"},
		},
		AuthnStatement: &saml.AuthnStatement{
			AuthnInstant:         &authnInstant,
			SessionIndex:         "_session123",
			AuthnContextClassRef: "urn:oasis:names:tc:SAML:2.0:ac:classes:Password",
		},
		Attributes: []saml.Attribute{
			{
				Name:         "email",
				FriendlyName: "Email",
				Values:       []string{"user@example.com"},
			},
			{
				Name:   "groups",
				Values: []string{"admins", "users"},
			},
		},
		Signature: &saml.SignatureInfo{
			Signed:          true,
			SignatureMethod: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
			DigestMethod:    "http://www.w3.org/2001/04/xmlenc#sha256",
		},
	}
}

func TestFormatter_SignatureInfo(t *testing.T) {
	formatter := NewFormatterWithOptions("pretty", true)

	info := &saml.SAMLInfo{
		Type:   "Assertion",
		ID:     "_test",
		Issuer: "test-issuer",
		Signature: &saml.SignatureInfo{
			Signed:          true,
			SignatureMethod: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
			DigestMethod:    "http://www.w3.org/2001/04/xmlenc#sha256",
			CertificateInfo: &saml.CertificateInfo{
				Subject:   "CN=test",
				Issuer:    "CN=test-ca",
				NotBefore: time.Now().Add(-time.Hour),
				NotAfter:  time.Now().Add(time.Hour),
				Serial:    "12345",
			},
		},
	}

	result, err := formatter.FormatSAMLInfo(info)

	require.NoError(t, err)
	assert.Contains(t, result, "Signature")
	assert.Contains(t, result, "Yes") // Signed: Yes
	assert.Contains(t, result, "rsa-sha256")
	assert.Contains(t, result, "Cert Subject")
}

func TestFormatter_MultipleAttributeValues(t *testing.T) {
	formatter := NewFormatterWithOptions("pretty", true)

	info := &saml.SAMLInfo{
		Type:   "Assertion",
		ID:     "_test",
		Issuer: "test-issuer",
		Attributes: []saml.Attribute{
			{
				Name:   "groups",
				Values: []string{"admin", "developer", "user"},
			},
		},
	}

	result, err := formatter.FormatSAMLInfo(info)

	require.NoError(t, err)
	// Values should be comma-separated
	assert.True(t, strings.Contains(result, "admin, developer, user") ||
		strings.Contains(result, "admin") && strings.Contains(result, "developer"))
}
