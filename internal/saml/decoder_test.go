package saml

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoder_Decode(t *testing.T) {
	decoder := NewDecoder()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "simple base64",
			input:    base64.StdEncoding.EncodeToString([]byte("<saml>test</saml>")),
			expected: "<saml>test</saml>",
		},
		{
			name:     "base64 with whitespace",
			input:    "  " + base64.StdEncoding.EncodeToString([]byte("<saml>test</saml>")) + "  \n",
			expected: "<saml>test</saml>",
		},
		{
			name:     "base64 with newlines in middle",
			input:    "PHNhbWw+\ndGVzdDwv\nc2FtbD4=",
			expected: "<saml>test</saml>",
		},
		{
			name:        "invalid base64",
			input:       "not-valid-base64!!!",
			expectError: true,
		},
		{
			name:     "url-safe base64",
			input:    base64.URLEncoding.EncodeToString([]byte("<saml>test</saml>")),
			expected: "<saml>test</saml>",
		},
		{
			name:     "base64 without padding",
			input:    "PHNhbWw+dGVzdDwvc2FtbD4", // Missing ==
			expected: "<saml>test</saml>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.Decode(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestDecoder_DecodeDeflate(t *testing.T) {
	decoder := NewDecoder()

	// Create a deflated and encoded message
	original := "<saml>deflate test</saml>"
	encoded, err := decoder.EncodeDeflate([]byte(original))
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "valid deflate encoded",
			input:    encoded,
			expected: original,
		},
		{
			name:        "invalid base64",
			input:       "not-valid!!!",
			expectError: true,
		},
		{
			name:        "valid base64 but not deflated",
			input:       base64.StdEncoding.EncodeToString([]byte("not deflated")),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.DecodeDeflate(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestDecoder_RoundTrip(t *testing.T) {
	decoder := NewDecoder()

	original := `<?xml version="1.0"?><samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">test</samlp:AuthnRequest>`

	// Test simple encode/decode roundtrip
	t.Run("simple roundtrip", func(t *testing.T) {
		encoded := decoder.Encode([]byte(original))
		decoded, err := decoder.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, original, string(decoded))
	})

	// Test deflate encode/decode roundtrip
	t.Run("deflate roundtrip", func(t *testing.T) {
		encoded, err := decoder.EncodeDeflate([]byte(original))
		require.NoError(t, err)

		decoded, err := decoder.DecodeDeflate(encoded)
		require.NoError(t, err)
		assert.Equal(t, original, string(decoded))
	})
}

func TestDecoder_Deflate(t *testing.T) {
	decoder := NewDecoder()

	input := []byte("<saml>test data for deflate compression</saml>")

	deflated, err := decoder.Deflate(input)
	require.NoError(t, err)

	// Deflated should be smaller than or equal to input (for small inputs might not compress much)
	assert.NotNil(t, deflated)

	// Should be able to inflate back
	inflated, err := decoder.inflate(deflated)
	require.NoError(t, err)
	assert.Equal(t, input, inflated)
}

func TestDecoder_URLEncodedInput(t *testing.T) {
	decoder := NewDecoder()

	original := "<saml>test</saml>"
	base64Encoded := base64.StdEncoding.EncodeToString([]byte(original))

	// URL encode the base64 (as might come from a query parameter)
	urlEncoded := base64Encoded // In this case it doesn't need URL encoding, but test the path

	decoded, err := decoder.Decode(urlEncoded)
	require.NoError(t, err)
	assert.Equal(t, original, string(decoded))
}

func TestIsBase64Encoded(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "base64 encoded string",
			input:    "PHNhbWw+dGVzdDwvc2FtbD4=",
			expected: true,
		},
		{
			name:     "raw XML",
			input:    "<saml>test</saml>",
			expected: false,
		},
		{
			name:     "XML with whitespace",
			input:    "  <saml>test</saml>  ",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "base64 with newlines",
			input:    "PHNhbWw+\ndGVzdDwv\nc2FtbD4=",
			expected: true,
		},
		{
			name:     "URL-safe base64",
			input:    "PHNhbWw-dGVzdDwvc2FtbD4_",
			expected: true,
		},
		{
			name:     "invalid characters",
			input:    "hello world!@#$",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBase64Encoded(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecoder_SmartDecode(t *testing.T) {
	decoder := NewDecoder()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "raw XML passthrough",
			input:    "<saml>test</saml>",
			expected: "<saml>test</saml>",
		},
		{
			name:     "base64 encoded XML",
			input:    base64.StdEncoding.EncodeToString([]byte("<saml>test</saml>")),
			expected: "<saml>test</saml>",
		},
		{
			name:     "XML with leading whitespace",
			input:    "  <saml>test</saml>",
			expected: "<saml>test</saml>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.SmartDecode(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestDecoder_SmartDecode_Deflate(t *testing.T) {
	decoder := NewDecoder()

	original := "<saml>deflated content</saml>"
	encoded, err := decoder.EncodeDeflate([]byte(original))
	require.NoError(t, err)

	// SmartDecode should auto-detect deflate and decompress
	result, err := decoder.SmartDecode(encoded)
	require.NoError(t, err)
	assert.Equal(t, original, string(result))
}
