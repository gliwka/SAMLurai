package cmd

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractCommand(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "samlurai-extract-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Sample SAML Response (single line to avoid JSON encoding issues)
	samlResponse := `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="_response123"><saml:Subject xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"><saml:NameID>user@example.com</saml:NameID></saml:Subject></samlp:Response>`

	encodedSAML := base64.StdEncoding.EncodeToString([]byte(samlResponse))

	// Create test HAR file
	harContent := `{
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
	}`

	harFile := filepath.Join(tmpDir, "test.har")
	if err := os.WriteFile(harFile, []byte(harContent), 0644); err != nil {
		t.Fatalf("Failed to create HAR file: %v", err)
	}

	t.Run("list mode", func(t *testing.T) {
		// Reset flags
		extractFile = ""
		extractOutputDir = "."
		extractList = false

		cmd := GetRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		cmd.SetArgs([]string{"extract", "-f", harFile, "--list"})
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Found 1 SAML assertion") {
			t.Errorf("Expected to find 1 SAML assertion, got: %s", output)
		}
		if !strings.Contains(output, "Response") {
			t.Errorf("Expected to see Response type, got: %s", output)
		}
	})

	t.Run("extract mode", func(t *testing.T) {
		// Reset flags
		extractFile = ""
		extractOutputDir = "."
		extractList = false

		outputDir := filepath.Join(tmpDir, "extracted")

		cmd := GetRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		cmd.SetArgs([]string{"extract", "-f", harFile, "-d", outputDir})
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		// Check that file was created
		expectedFile := filepath.Join(outputDir, "saml_001_response_request_body.xml")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", expectedFile)
		}

		// Check file content contains SAML
		content, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("Failed to read extracted file: %v", err)
		}
		contentStr := string(content)
		// Check for either namespaced or non-namespaced format (formatter may modify)
		if !strings.Contains(contentStr, "Response") || !strings.Contains(contentStr, "SAML") {
			t.Errorf("Extracted file doesn't look like SAML Response, got: %s", contentStr)
		}
	})

	t.Run("no SAML found", func(t *testing.T) {
		// Reset flags
		extractFile = ""
		extractOutputDir = "."
		extractList = false

		// Create empty HAR file
		emptyHAR := `{"log": {"entries": []}}`
		emptyHarFile := filepath.Join(tmpDir, "empty.har")
		if err := os.WriteFile(emptyHarFile, []byte(emptyHAR), 0644); err != nil {
			t.Fatalf("Failed to create empty HAR file: %v", err)
		}

		cmd := GetRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		cmd.SetArgs([]string{"extract", "-f", emptyHarFile})
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "No SAML assertions found") {
			t.Errorf("Expected 'No SAML assertions found' message, got: %s", output)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		// Reset flags
		extractFile = ""
		extractOutputDir = "."
		extractList = false

		cmd := GetRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		cmd.SetArgs([]string{"extract", "-f", "/nonexistent/file.har"})
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("invalid HAR file", func(t *testing.T) {
		// Reset flags
		extractFile = ""
		extractOutputDir = "."
		extractList = false

		// Create invalid HAR file
		invalidHarFile := filepath.Join(tmpDir, "invalid.har")
		if err := os.WriteFile(invalidHarFile, []byte("not json"), 0644); err != nil {
			t.Fatalf("Failed to create invalid HAR file: %v", err)
		}

		cmd := GetRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		cmd.SetArgs([]string{"extract", "-f", invalidHarFile})
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for invalid HAR file")
		}
	})
}

func TestExtractMultipleSAML(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "samlurai-extract-multi-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Single-line SAML to avoid JSON encoding issues
	samlResponse := `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="_resp1"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">idp</saml:Issuer></samlp:Response>`
	samlRequest := `<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="_req1"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">sp</saml:Issuer></samlp:AuthnRequest>`

	encodedResponse := base64.StdEncoding.EncodeToString([]byte(samlResponse))
	encodedRequest := base64.StdEncoding.EncodeToString([]byte(samlRequest))

	harContent := `{
		"log": {
			"entries": [
				{
					"request": {
						"method": "GET",
						"url": "https://idp.example.com/sso?SAMLRequest=` + encodedRequest + `"
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

	harFile := filepath.Join(tmpDir, "multi.har")
	if err := os.WriteFile(harFile, []byte(harContent), 0644); err != nil {
		t.Fatalf("Failed to create HAR file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "extracted")

	// Reset flags
	extractFile = ""
	extractOutputDir = "."
	extractList = false

	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"extract", "-f", harFile, "-d", outputDir})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Check that both files were created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	output := buf.String()
	if !strings.Contains(output, "Extracted 2 SAML assertion") {
		t.Errorf("Expected to extract 2 SAML assertions, got: %s", output)
	}
}

func TestTruncateURL(t *testing.T) {
	tests := []struct {
		url    string
		maxLen int
		want   string
	}{
		{"https://example.com", 50, "https://example.com"},
		{"https://example.com/very/long/path/that/exceeds/limit", 30, "https://example.com/very/lo..."},
		{"short", 10, "short"},
	}

	for _, tt := range tests {
		got := truncateURL(tt.url, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateURL(%q, %d) = %q, want %q", tt.url, tt.maxLen, got, tt.want)
		}
	}
}
