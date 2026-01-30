package saml

import (
	"encoding/base64"
	"testing"
)

func TestHARExtractor_ExtractFromHAR(t *testing.T) {
	extractor := NewHARExtractor()

	// Sample SAML Response (single line to avoid JSON encoding issues)
	samlResponse := `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="_response123"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">https://idp.example.com</saml:Issuer></samlp:Response>`

	encodedSAML := base64.StdEncoding.EncodeToString([]byte(samlResponse))

	tests := []struct {
		name          string
		har           string
		wantCount     int
		wantType      string
		wantSource    string
		wantParamName string
	}{
		{
			name: "extract from POST body params",
			har: `{
				"log": {
					"entries": [{
						"request": {
							"method": "POST",
							"url": "https://sp.example.com/acs",
							"postData": {
								"mimeType": "application/x-www-form-urlencoded",
								"params": [
									{"name": "SAMLResponse", "value": "` + encodedSAML + `"}
								]
							}
						},
						"response": {
							"content": {"mimeType": "text/html", "text": ""}
						}
					}]
				}
			}`,
			wantCount:     1,
			wantType:      "Response",
			wantSource:    "request-body",
			wantParamName: "SAMLResponse",
		},
		{
			name: "extract from query string",
			har: `{
				"log": {
					"entries": [{
						"request": {
							"method": "GET",
							"url": "https://sp.example.com/acs?SAMLResponse=` + encodedSAML + `",
							"queryString": []
						},
						"response": {
							"content": {"mimeType": "text/html", "text": ""}
						}
					}]
				}
			}`,
			wantCount:     1,
			wantType:      "Response",
			wantSource:    "request-query",
			wantParamName: "SAMLResponse",
		},
		{
			name: "extract from HTML response with hidden field",
			har: `{
				"log": {
					"entries": [{
						"request": {
							"method": "GET",
							"url": "https://idp.example.com/sso"
						},
						"response": {
							"content": {
								"mimeType": "text/html",
								"text": "<html><body><form><input type=\"hidden\" name=\"SAMLResponse\" value=\"` + encodedSAML + `\"/></form></body></html>"
							}
						}
					}]
				}
			}`,
			wantCount:     1,
			wantType:      "Response",
			wantSource:    "response-body",
			wantParamName: "SAMLResponse",
		},
		{
			name: "no SAML data",
			har: `{
				"log": {
					"entries": [{
						"request": {
							"method": "GET",
							"url": "https://example.com/"
						},
						"response": {
							"content": {"mimeType": "text/html", "text": "<html></html>"}
						}
					}]
				}
			}`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := extractor.ExtractFromHAR([]byte(tt.har))
			if err != nil {
				t.Fatalf("ExtractFromHAR() error = %v", err)
			}

			if len(results) != tt.wantCount {
				t.Errorf("ExtractFromHAR() got %d results, want %d", len(results), tt.wantCount)
			}

			if tt.wantCount > 0 && len(results) > 0 {
				result := results[0]
				if result.Type != tt.wantType {
					t.Errorf("Type = %q, want %q", result.Type, tt.wantType)
				}
				if result.Source != tt.wantSource {
					t.Errorf("Source = %q, want %q", result.Source, tt.wantSource)
				}
				if tt.wantParamName != "" && result.ParameterName != tt.wantParamName {
					t.Errorf("ParameterName = %q, want %q", result.ParameterName, tt.wantParamName)
				}
			}
		})
	}
}

func TestHARExtractor_ExtractMultiple(t *testing.T) {
	extractor := NewHARExtractor()

	// Single line SAML to avoid JSON encoding issues
	samlResponse := `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="_resp1"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">https://idp.example.com</saml:Issuer></samlp:Response>`
	samlRequest := `<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="_req1"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">https://sp.example.com</saml:Issuer></samlp:AuthnRequest>`

	encodedResponse := base64.StdEncoding.EncodeToString([]byte(samlResponse))
	encodedRequest := base64.StdEncoding.EncodeToString([]byte(samlRequest))

	har := `{
		"log": {
			"entries": [
				{
					"request": {
						"method": "GET",
						"url": "https://idp.example.com/sso",
						"queryString": [{"name": "SAMLRequest", "value": "` + encodedRequest + `"}]
					},
					"response": {
						"content": {"mimeType": "text/html", "text": ""}
					}
				},
				{
					"request": {
						"method": "POST",
						"url": "https://sp.example.com/acs",
						"postData": {
							"mimeType": "application/x-www-form-urlencoded",
							"params": [{"name": "SAMLResponse", "value": "` + encodedResponse + `"}]
						}
					},
					"response": {
						"content": {"mimeType": "text/html", "text": ""}
					}
				}
			]
		}
	}`

	results, err := extractor.ExtractFromHAR([]byte(har))
	if err != nil {
		t.Fatalf("ExtractFromHAR() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Check first result (AuthnRequest)
	if results[0].Type != "AuthnRequest" {
		t.Errorf("First result Type = %q, want AuthnRequest", results[0].Type)
	}
	if results[0].Index != 1 {
		t.Errorf("First result Index = %d, want 1", results[0].Index)
	}

	// Check second result (Response)
	if results[1].Type != "Response" {
		t.Errorf("Second result Type = %q, want Response", results[1].Type)
	}
	if results[1].Index != 2 {
		t.Errorf("Second result Index = %d, want 2", results[1].Index)
	}
}

func TestHARExtractor_GenerateFilename(t *testing.T) {
	extractor := NewHARExtractor()

	tests := []struct {
		extracted ExtractedSAML
		want      string
	}{
		{
			extracted: ExtractedSAML{Index: 1, Type: "Response", Source: "request-body"},
			want:      "saml_001_response_request_body.xml",
		},
		{
			extracted: ExtractedSAML{Index: 2, Type: "AuthnRequest", Source: "request-query"},
			want:      "saml_002_authnrequest_request_query.xml",
		},
		{
			extracted: ExtractedSAML{Index: 10, Type: "Assertion", Source: "response-body"},
			want:      "saml_010_assertion_response_body.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := extractor.GenerateFilename(tt.extracted)
			if got != tt.want {
				t.Errorf("GenerateFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHARExtractor_DetectSAMLType(t *testing.T) {
	extractor := NewHARExtractor()

	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "SAML 2.0 Response",
			xml:  `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"/>`,
			want: "Response",
		},
		{
			name: "SAML 2.0 AuthnRequest",
			xml:  `<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"/>`,
			want: "AuthnRequest",
		},
		{
			name: "SAML 2.0 Assertion",
			xml:  `<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"/>`,
			want: "Assertion",
		},
		{
			name: "LogoutRequest",
			xml:  `<samlp:LogoutRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"/>`,
			want: "LogoutRequest",
		},
		{
			name: "Unknown",
			xml:  `<something/>`,
			want: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.detectSAMLType([]byte(tt.xml))
			if got != tt.want {
				t.Errorf("detectSAMLType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHARExtractor_InvalidHAR(t *testing.T) {
	extractor := NewHARExtractor()

	_, err := extractor.ExtractFromHAR([]byte("not valid json"))
	if err == nil {
		t.Error("Expected error for invalid HAR, got nil")
	}
}

func TestHARExtractor_ExtractFromBase64(t *testing.T) {
	extractor := NewHARExtractor()

	samlResponse := `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="_response123"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">https://idp.example.com</saml:Issuer></samlp:Response>`

	encoded := base64.StdEncoding.EncodeToString([]byte(samlResponse))

	result, err := extractor.ExtractFromBase64(encoded)
	if err != nil {
		t.Fatalf("ExtractFromBase64() error = %v", err)
	}

	if result.Type != "Response" {
		t.Errorf("Type = %q, want Response", result.Type)
	}

	if result.Source != "direct-input" {
		t.Errorf("Source = %q, want direct-input", result.Source)
	}
}
