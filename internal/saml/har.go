package saml

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// HAR represents the root structure of a HAR file
type HAR struct {
	Log HARLog `json:"log"`
}

// HARLog represents the log entry in a HAR file
type HARLog struct {
	Entries []HAREntry `json:"entries"`
}

// HAREntry represents a single HTTP request/response entry
type HAREntry struct {
	Request  HARRequest  `json:"request"`
	Response HARResponse `json:"response"`
}

// HARRequest represents an HTTP request
type HARRequest struct {
	Method      string         `json:"method"`
	URL         string         `json:"url"`
	PostData    *HARPostData   `json:"postData,omitempty"`
	QueryString []HARNameValue `json:"queryString,omitempty"`
}

// HARResponse represents an HTTP response
type HARResponse struct {
	Content HARContent `json:"content"`
}

// HARPostData represents POST data
type HARPostData struct {
	MimeType string         `json:"mimeType"`
	Text     string         `json:"text"`
	Params   []HARNameValue `json:"params,omitempty"`
}

// HARContent represents response content
type HARContent struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// HARNameValue represents a name-value pair (query params, form params)
type HARNameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ExtractedSAML represents an extracted SAML assertion with metadata
type ExtractedSAML struct {
	// Index is the sequential number of this extraction
	Index int `json:"index"`

	// Type indicates the SAML message type (Response, Request, Assertion, etc.)
	Type string `json:"type"`

	// Source indicates where the SAML was found (request-body, request-query, response-body)
	Source string `json:"source"`

	// URL is the request URL where this SAML was found
	URL string `json:"url"`

	// ParameterName is the form/query parameter name (e.g., SAMLResponse, SAMLRequest)
	ParameterName string `json:"parameter_name,omitempty"`

	// RawValue is the original encoded value
	RawValue string `json:"raw_value"`

	// DecodedXML is the decoded SAML XML
	DecodedXML []byte `json:"decoded_xml"`

	// WasDeflated indicates if deflate decompression was applied
	WasDeflated bool `json:"was_deflated"`
}

// HARExtractor extracts SAML assertions from HAR files
type HARExtractor struct {
	decoder *Decoder
}

// NewHARExtractor creates a new HAR extractor
func NewHARExtractor() *HARExtractor {
	return &HARExtractor{
		decoder: NewDecoder(),
	}
}

// ExtractFromHAR extracts all SAML assertions from a HAR file
func (e *HARExtractor) ExtractFromHAR(data []byte) ([]ExtractedSAML, error) {
	var har HAR
	if err := json.Unmarshal(data, &har); err != nil {
		return nil, fmt.Errorf("failed to parse HAR file: %w", err)
	}

	var results []ExtractedSAML
	index := 1

	for _, entry := range har.Log.Entries {
		// Check request query parameters
		extracted := e.extractFromQueryParams(entry.Request.QueryString, entry.Request.URL, &index)
		results = append(results, extracted...)

		// Check request POST data
		if entry.Request.PostData != nil {
			extracted = e.extractFromPostData(entry.Request.PostData, entry.Request.URL, &index)
			results = append(results, extracted...)
		}

		// Check response body for SAML content
		extracted = e.extractFromResponseBody(entry.Response.Content, entry.Request.URL, &index)
		results = append(results, extracted...)
	}

	return results, nil
}

// extractFromQueryParams extracts SAML from URL query parameters
func (e *HARExtractor) extractFromQueryParams(params []HARNameValue, requestURL string, index *int) []ExtractedSAML {
	var results []ExtractedSAML

	// Also parse the URL itself for query params not in the array
	if parsedURL, err := url.Parse(requestURL); err == nil {
		for key, values := range parsedURL.Query() {
			for _, value := range values {
				if e.isSAMLParameter(key) {
					if extracted := e.tryExtractSAML(value, key, requestURL, "request-query", index); extracted != nil {
						results = append(results, *extracted)
					}
				}
			}
		}
	}

	for _, param := range params {
		if e.isSAMLParameter(param.Name) {
			if extracted := e.tryExtractSAML(param.Value, param.Name, requestURL, "request-query", index); extracted != nil {
				results = append(results, *extracted)
			}
		}
	}

	return results
}

// extractFromPostData extracts SAML from POST body
func (e *HARExtractor) extractFromPostData(postData *HARPostData, requestURL string, index *int) []ExtractedSAML {
	var results []ExtractedSAML

	// Check form params
	for _, param := range postData.Params {
		if e.isSAMLParameter(param.Name) {
			if extracted := e.tryExtractSAML(param.Value, param.Name, requestURL, "request-body", index); extracted != nil {
				results = append(results, *extracted)
			}
		}
	}

	// Parse URL-encoded body
	if strings.Contains(postData.MimeType, "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(postData.Text)
		if err == nil {
			for key, vals := range values {
				if e.isSAMLParameter(key) {
					for _, val := range vals {
						if extracted := e.tryExtractSAML(val, key, requestURL, "request-body", index); extracted != nil {
							results = append(results, *extracted)
						}
					}
				}
			}
		}
	}

	// Try to extract SAML from raw body (might be base64 encoded SAML directly)
	if extracted := e.tryExtractSAML(postData.Text, "", requestURL, "request-body", index); extracted != nil {
		results = append(results, *extracted)
	}

	return results
}

// extractFromResponseBody extracts SAML from response body
func (e *HARExtractor) extractFromResponseBody(content HARContent, requestURL string, index *int) []ExtractedSAML {
	var results []ExtractedSAML

	if content.Text == "" {
		return results
	}

	// Check for SAML in HTML form (common for POST binding)
	samlMatches := e.extractSAMLFromHTML(content.Text)
	for paramName, value := range samlMatches {
		if extracted := e.tryExtractSAML(value, paramName, requestURL, "response-body", index); extracted != nil {
			results = append(results, *extracted)
		}
	}

	// Try direct extraction if content looks like SAML or base64
	if extracted := e.tryExtractSAML(content.Text, "", requestURL, "response-body", index); extracted != nil {
		results = append(results, *extracted)
	}

	return results
}

// extractSAMLFromHTML extracts SAML values from hidden form fields in HTML
func (e *HARExtractor) extractSAMLFromHTML(html string) map[string]string {
	results := make(map[string]string)

	// Pattern to match hidden input fields with SAML data
	// Matches: <input type="hidden" name="SAMLResponse" value="..."/>
	patterns := []string{
		`<input[^>]*name=["']?(SAMLResponse|SAMLRequest|SAMLAssertion)["']?[^>]*value=["']([^"']+)["']`,
		`<input[^>]*value=["']([^"']+)["'][^>]*name=["']?(SAMLResponse|SAMLRequest|SAMLAssertion)["']?`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(html, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				// Order depends on which pattern matched
				if strings.HasPrefix(match[1], "SAML") {
					results[match[1]] = match[2]
				} else {
					results[match[2]] = match[1]
				}
			}
		}
	}

	return results
}

// isSAMLParameter checks if a parameter name is a known SAML parameter
func (e *HARExtractor) isSAMLParameter(name string) bool {
	lowerName := strings.ToLower(name)
	samlParams := []string{
		"samlresponse",
		"samlrequest",
		"samlassertion",
		"samlart",
		"logoutrequest",
		"logoutresponse",
	}

	for _, param := range samlParams {
		if lowerName == param {
			return true
		}
	}
	return false
}

// tryExtractSAML attempts to extract and decode SAML from a value
func (e *HARExtractor) tryExtractSAML(value, paramName, requestURL, source string, index *int) *ExtractedSAML {
	if value == "" {
		return nil
	}

	// URL decode first if necessary
	decoded, err := url.QueryUnescape(value)
	if err == nil && decoded != value {
		value = decoded
	}

	var xmlData []byte
	var wasDeflated bool

	// Try regular base64 decode first
	xmlData, err = e.decoder.Decode(value)
	if err != nil {
		return nil
	}

	// Check if it looks like XML
	if !e.looksLikeXML(xmlData) {
		// Try deflate decompression
		xmlData, err = e.decoder.DecodeDeflate(value)
		if err != nil {
			return nil
		}
		wasDeflated = true
	}

	// Validate it's actually SAML
	if !e.isSAMLXML(xmlData) {
		return nil
	}

	samlType := e.detectSAMLType(xmlData)

	result := &ExtractedSAML{
		Index:         *index,
		Type:          samlType,
		Source:        source,
		URL:           requestURL,
		ParameterName: paramName,
		RawValue:      value,
		DecodedXML:    xmlData,
		WasDeflated:   wasDeflated,
	}

	*index++
	return result
}

// looksLikeXML checks if data appears to be XML
func (e *HARExtractor) looksLikeXML(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	return strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<")
}

// isSAMLXML checks if XML data is SAML
func (e *HARExtractor) isSAMLXML(data []byte) bool {
	content := string(data)
	samlIndicators := []string{
		"samlp:Response",
		"saml2p:Response",
		"samlp:AuthnRequest",
		"saml2p:AuthnRequest",
		"saml:Assertion",
		"saml2:Assertion",
		"urn:oasis:names:tc:SAML",
		"<Response",
		"<AuthnRequest",
		"<Assertion",
		"<LogoutRequest",
		"<LogoutResponse",
	}

	for _, indicator := range samlIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	return false
}

// detectSAMLType determines the type of SAML message
// Order matters: check Response/Request types before Assertion since
// responses contain assertions
func (e *HARExtractor) detectSAMLType(data []byte) string {
	content := string(data)

	// Check in order of specificity - Response/Request wrappers first
	// since they contain Assertions inside them
	typeChecks := []struct {
		typeName   string
		indicators []string
	}{
		{
			"Response",
			[]string{"samlp:Response", "saml2p:Response", "<Response "},
		},
		{
			"AuthnRequest",
			[]string{"samlp:AuthnRequest", "saml2p:AuthnRequest", "<AuthnRequest "},
		},
		{
			"LogoutRequest",
			[]string{"samlp:LogoutRequest", "saml2p:LogoutRequest", "<LogoutRequest "},
		},
		{
			"LogoutResponse",
			[]string{"samlp:LogoutResponse", "saml2p:LogoutResponse", "<LogoutResponse "},
		},
		{
			"Assertion",
			[]string{"saml:Assertion", "saml2:Assertion", "<Assertion "},
		},
	}

	for _, check := range typeChecks {
		for _, indicator := range check.indicators {
			if strings.Contains(content, indicator) {
				return check.typeName
			}
		}
	}

	return "Unknown"
}

// GenerateFilename generates a descriptive filename for an extracted SAML
func (e *HARExtractor) GenerateFilename(extracted ExtractedSAML) string {
	// Format: saml_<index>_<type>_<source>.xml
	safeType := strings.ToLower(strings.ReplaceAll(extracted.Type, " ", "_"))
	safeSource := strings.ReplaceAll(extracted.Source, "-", "_")
	return fmt.Sprintf("saml_%03d_%s_%s.xml", extracted.Index, safeType, safeSource)
}

// ExtractFromBase64 extracts SAML from a raw base64 string (for direct input)
func (e *HARExtractor) ExtractFromBase64(value string) (*ExtractedSAML, error) {
	var xmlData []byte
	var wasDeflated bool
	var err error

	// Try regular base64 decode first
	xmlData, err = e.decoder.Decode(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check if it looks like XML
	if !e.looksLikeXML(xmlData) {
		// Try deflate decompression
		xmlData, err = e.decoder.DecodeDeflate(value)
		if err != nil {
			return nil, fmt.Errorf("failed to decode with deflate: %w", err)
		}
		wasDeflated = true
	}

	// Validate it's actually SAML
	if !e.isSAMLXML(xmlData) {
		return nil, fmt.Errorf("decoded content is not valid SAML XML")
	}

	return &ExtractedSAML{
		Index:       1,
		Type:        e.detectSAMLType(xmlData),
		Source:      "direct-input",
		RawValue:    value,
		DecodedXML:  xmlData,
		WasDeflated: wasDeflated,
	}, nil
}
