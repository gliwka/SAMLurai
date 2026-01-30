package saml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_ParseResponse(t *testing.T) {
	parser := NewParser()

	responseXML, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "assertions", "response.xml"))
	require.NoError(t, err)

	info, err := parser.Parse(responseXML)
	require.NoError(t, err)

	assert.Equal(t, "Response", info.Type)
	assert.Equal(t, "_response123", info.ID)
	assert.Equal(t, "https://idp.example.com", info.Issuer)
	assert.Equal(t, "https://sp.example.com/acs", info.Destination)
	assert.Equal(t, "_request456", info.InResponseTo)

	// Check status
	require.NotNil(t, info.Status)
	assert.Equal(t, "Success", info.Status.StatusCode)

	// Check nested assertion
	require.NotNil(t, info.Assertion)
	assert.Equal(t, "Assertion", info.Assertion.Type)
	assert.Equal(t, "_assertion789", info.Assertion.ID)

	// Check subject in assertion
	require.NotNil(t, info.Assertion.Subject)
	assert.Equal(t, "user@example.com", info.Assertion.Subject.NameID)
	assert.Contains(t, info.Assertion.Subject.NameIDFormat, "emailAddress")

	// Check conditions
	require.NotNil(t, info.Assertion.Conditions)
	assert.NotNil(t, info.Assertion.Conditions.NotBefore)
	assert.NotNil(t, info.Assertion.Conditions.NotOnOrAfter)
	assert.Contains(t, info.Assertion.Conditions.AudienceRestriction, "https://sp.example.com")

	// Check authn statement
	require.NotNil(t, info.Assertion.AuthnStatement)
	assert.Equal(t, "_session123", info.Assertion.AuthnStatement.SessionIndex)
	assert.Contains(t, info.Assertion.AuthnStatement.AuthnContextClassRef, "PasswordProtectedTransport")

	// Check attributes
	require.Len(t, info.Assertion.Attributes, 4)

	emailAttr := findAttribute(info.Assertion.Attributes, "email")
	require.NotNil(t, emailAttr)
	assert.Equal(t, "Email", emailAttr.FriendlyName)
	assert.Contains(t, emailAttr.Values, "user@example.com")

	groupsAttr := findAttribute(info.Assertion.Attributes, "groups")
	require.NotNil(t, groupsAttr)
	assert.Len(t, groupsAttr.Values, 2)
	assert.Contains(t, groupsAttr.Values, "admins")
	assert.Contains(t, groupsAttr.Values, "users")
}

func TestParser_ParseAssertion(t *testing.T) {
	parser := NewParser()

	assertionXML, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "assertions", "assertion.xml"))
	require.NoError(t, err)

	info, err := parser.Parse(assertionXML)
	require.NoError(t, err)

	assert.Equal(t, "Assertion", info.Type)
	assert.Equal(t, "_assertion789", info.ID)
	assert.Equal(t, "https://idp.example.com", info.Issuer)

	// Check subject
	require.NotNil(t, info.Subject)
	assert.Equal(t, "user@example.com", info.Subject.NameID)

	// Check conditions
	require.NotNil(t, info.Conditions)
	assert.NotNil(t, info.Conditions.NotBefore)

	// Check attributes
	require.Len(t, info.Attributes, 1)
	assert.Equal(t, "email", info.Attributes[0].Name)
}

func TestParser_ParseInvalidXML(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "not XML",
			input: "this is not XML",
		},
		{
			name:  "malformed XML",
			input: "<saml><broken>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse([]byte(tt.input))
			assert.Error(t, err)
		})
	}
}

func TestParser_ExtractStatusCode(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "urn:oasis:names:tc:SAML:2.0:status:Success",
			expected: "Success",
		},
		{
			input:    "urn:oasis:names:tc:SAML:2.0:status:Requester",
			expected: "Requester",
		},
		{
			input:    "Success",
			expected: "Success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parser.extractStatusCode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_ParseMinimalAssertion(t *testing.T) {
	parser := NewParser()

	minimal := `<?xml version="1.0"?>
<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="_min123">
	<saml:Issuer>https://idp.example.com</saml:Issuer>
</saml:Assertion>`

	info, err := parser.Parse([]byte(minimal))
	require.NoError(t, err)

	assert.Equal(t, "Assertion", info.Type)
	assert.Equal(t, "_min123", info.ID)
	assert.Equal(t, "https://idp.example.com", info.Issuer)
	assert.Nil(t, info.Subject)
	assert.Nil(t, info.Conditions)
}

func TestParser_ParseAuthnRequest(t *testing.T) {
	parser := NewParser()

	authnRequest := `<?xml version="1.0"?>
<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
    ID="_req123"
    Version="2.0"
    IssueInstant="2024-01-15T10:00:00Z"
    Destination="https://idp.example.com/sso"
    AssertionConsumerServiceURL="https://sp.example.com/acs"
    ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
    ForceAuthn="true"
    IsPassive="false">
    <saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">https://sp.example.com</saml:Issuer>
    <samlp:NameIDPolicy Format="urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress" AllowCreate="true"/>
</samlp:AuthnRequest>`

	info, err := parser.Parse([]byte(authnRequest))
	require.NoError(t, err)

	assert.Equal(t, "AuthnRequest", info.Type)
	assert.Equal(t, "_req123", info.ID)
	assert.Equal(t, "https://sp.example.com", info.Issuer)
	assert.Equal(t, "https://idp.example.com/sso", info.Destination)
	assert.Equal(t, "https://sp.example.com/acs", info.AssertionConsumerServiceURL)
	assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST", info.ProtocolBinding)
	
	require.NotNil(t, info.ForceAuthn)
	assert.True(t, *info.ForceAuthn)
	
	require.NotNil(t, info.IsPassive)
	assert.False(t, *info.IsPassive)

	require.NotNil(t, info.NameIDPolicy)
	assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress", info.NameIDPolicy.Format)
	require.NotNil(t, info.NameIDPolicy.AllowCreate)
	assert.True(t, *info.NameIDPolicy.AllowCreate)
}

func TestParser_ParseAuthnRequestWithExtensions(t *testing.T) {
	parser := NewParser()

	authnRequest := `<?xml version="1.0"?>
<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
    xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"
    ID="_req456"
    Version="2.0"
    Destination="https://idp.example.com/sso"
    AssertionConsumerServiceURL="https://sp.example.com/acs">
    <saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">https://sp.example.com</saml:Issuer>
    <samlp:Extensions>
        <md:RequestedAttribute Name="urn:oid:1.2.3.4" FriendlyName="email" isRequired="true"/>
        <md:RequestedAttribute Name="urn:oid:5.6.7.8" FriendlyName="name" isRequired="false"/>
    </samlp:Extensions>
</samlp:AuthnRequest>`

	info, err := parser.Parse([]byte(authnRequest))
	require.NoError(t, err)

	assert.Equal(t, "AuthnRequest", info.Type)
	assert.Equal(t, "_req456", info.ID)
	
	require.Len(t, info.RequestedAttributes, 2)
	
	assert.Equal(t, "urn:oid:1.2.3.4", info.RequestedAttributes[0].Name)
	assert.Equal(t, "email", info.RequestedAttributes[0].FriendlyName)
	require.NotNil(t, info.RequestedAttributes[0].IsRequired)
	assert.True(t, *info.RequestedAttributes[0].IsRequired)
	
	assert.Equal(t, "urn:oid:5.6.7.8", info.RequestedAttributes[1].Name)
	assert.Equal(t, "name", info.RequestedAttributes[1].FriendlyName)
	require.NotNil(t, info.RequestedAttributes[1].IsRequired)
	assert.False(t, *info.RequestedAttributes[1].IsRequired)
}

// Helper function to find an attribute by name
func findAttribute(attrs []Attribute, name string) *Attribute {
	for _, attr := range attrs {
		if attr.Name == name {
			return &attr
		}
	}
	return nil
}
