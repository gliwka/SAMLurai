package saml

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"
	"unicode/utf8"
)

// Decoder handles base64 and deflate decoding of SAML messages
type Decoder struct{}

// NewDecoder creates a new SAML decoder
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode decodes a base64-encoded SAML message
func (d *Decoder) Decode(input string) ([]byte, error) {
	// Clean up the input - remove whitespace and newlines
	cleaned := strings.ReplaceAll(input, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.TrimSpace(cleaned)

	// Try standard base64 first (before URL decoding to preserve + characters)
	decoded, err := base64.StdEncoding.DecodeString(cleaned)
	if err == nil {
		return decoded, nil
	}

	// Try URL-safe base64
	decoded, err = base64.URLEncoding.DecodeString(cleaned)
	if err == nil {
		return decoded, nil
	}

	// Try with padding adjustment
	decoded, err = d.decodeWithPaddingFix(cleaned)
	if err == nil {
		return decoded, nil
	}

	// Try URL decoding first (in case it's URL-encoded, e.g., from query params)
	urlDecoded, urlErr := url.QueryUnescape(cleaned)
	if urlErr == nil && urlDecoded != cleaned {
		decoded, err = base64.StdEncoding.DecodeString(urlDecoded)
		if err == nil {
			return decoded, nil
		}
		decoded, err = d.decodeWithPaddingFix(urlDecoded)
		if err == nil {
			return decoded, nil
		}
	}

	return nil, fmt.Errorf("base64 decode failed: %w", err)
}

// DecodeDeflate decodes a base64-encoded, deflate-compressed SAML message
// This is typically used for HTTP-Redirect binding
func (d *Decoder) DecodeDeflate(input string) ([]byte, error) {
	// First, base64 decode
	decoded, err := d.Decode(input)
	if err != nil {
		return nil, err
	}

	// Then, inflate (decompress)
	inflated, err := d.inflate(decoded)
	if err != nil {
		return nil, fmt.Errorf("deflate decompression failed: %w", err)
	}

	return inflated, nil
}

// decodeWithPaddingFix attempts to decode base64 with automatic padding correction
func (d *Decoder) decodeWithPaddingFix(input string) ([]byte, error) {
	// Add padding if necessary
	padded := input
	switch len(input) % 4 {
	case 2:
		padded += "=="
	case 3:
		padded += "="
	}

	decoded, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		// Try RawStdEncoding (no padding required)
		decoded, err = base64.RawStdEncoding.DecodeString(input)
		if err != nil {
			return nil, err
		}
	}

	return decoded, nil
}

// inflate decompresses deflate-compressed data
func (d *Decoder) inflate(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Deflate compresses data using deflate (useful for testing)
func (d *Decoder) Deflate(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Encode encodes data to base64 (useful for testing)
func (d *Decoder) Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// EncodeDeflate compresses and encodes data (useful for testing)
func (d *Decoder) EncodeDeflate(data []byte) (string, error) {
	deflated, err := d.Deflate(data)
	if err != nil {
		return "", err
	}
	return d.Encode(deflated), nil
}

// IsBase64Encoded checks if the input appears to be base64-encoded
// rather than raw XML. It checks if the input looks like XML (starts with <)
// or if it's likely base64 encoded.
func IsBase64Encoded(input string) bool {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return false
	}

	// If it starts with '<', it's likely XML
	if strings.HasPrefix(trimmed, "<") {
		return false
	}

	// Check if the content is valid base64 characters
	// Base64 uses A-Z, a-z, 0-9, +, /, and = for padding
	// URL-safe base64 uses - and _ instead of + and /
	for _, r := range trimmed {
		if !isBase64Char(r) && r != '\n' && r != '\r' && r != ' ' {
			return false
		}
	}

	return true
}

func isBase64Char(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') ||
		r == '+' || r == '/' || r == '=' ||
		r == '-' || r == '_' // URL-safe variants
}

// SmartDecode attempts to decode the input, auto-detecting if it's base64 encoded.
// If the input is already XML, it returns it as-is.
// It also tries deflate decompression if the decoded content isn't valid UTF-8/XML.
func (d *Decoder) SmartDecode(input string) ([]byte, error) {
	trimmed := strings.TrimSpace(input)

	// If it looks like XML, return as-is
	if !IsBase64Encoded(trimmed) {
		return []byte(trimmed), nil
	}

	// Try regular base64 decode first
	decoded, err := d.Decode(trimmed)
	if err != nil {
		return nil, err
	}

	// Check if the decoded content is valid UTF-8 and looks like XML
	if utf8.Valid(decoded) && len(decoded) > 0 && decoded[0] == '<' {
		return decoded, nil
	}

	// If not valid UTF-8 or not XML, try deflate decompression
	inflated, err := d.inflate(decoded)
	if err == nil && utf8.Valid(inflated) && len(inflated) > 0 && inflated[0] == '<' {
		return inflated, nil
	}

	// Return the base64-decoded content even if it doesn't look like XML
	// (could be binary or other format)
	return decoded, nil
}
