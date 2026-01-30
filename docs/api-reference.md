---
layout: default
title: API Reference
nav_order: 6
---

# API Reference
{: .no_toc }

Reference documentation for SAMLurai's internal packages.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Package Overview

SAMLurai is organized into the following packages:

| Package | Import Path | Description |
|:--------|:------------|:------------|
| `saml` | `github.com/gliwka/SAMLurai/internal/saml` | Core SAML processing |
| `output` | `github.com/gliwka/SAMLurai/internal/output` | Output formatting |
| `cmd` | `github.com/gliwka/SAMLurai/cmd` | CLI commands |

{: .note }
The `internal/` packages are not intended for external use and may change without notice. Use the CLI directly or fork the repository if you need programmatic access.

---

## saml Package

### Types

#### SAMLInfo

Primary data structure containing parsed SAML information.

```go
type SAMLInfo struct {
    // Type indicates if this is a Response or Assertion
    Type string `json:"type"`

    // Response-level fields
    ID           string     `json:"id,omitempty"`
    IssueInstant *time.Time `json:"issue_instant,omitempty"`
    Destination  string     `json:"destination,omitempty"`
    InResponseTo string     `json:"in_response_to,omitempty"`

    // Status (for responses)
    Status *Status `json:"status,omitempty"`

    // Issuer
    Issuer string `json:"issuer,omitempty"`

    // Subject
    Subject *Subject `json:"subject,omitempty"`

    // Conditions
    Conditions *Conditions `json:"conditions,omitempty"`

    // Authentication Statement
    AuthnStatement *AuthnStatement `json:"authn_statement,omitempty"`

    // Attributes
    Attributes []Attribute `json:"attributes,omitempty"`

    // Signature info
    Signature *SignatureInfo `json:"signature,omitempty"`

    // Nested assertion (for responses)
    Assertion *SAMLInfo `json:"assertion,omitempty"`
}
```

#### Subject

```go
type Subject struct {
    NameID          string `json:"name_id,omitempty"`
    NameIDFormat    string `json:"name_id_format,omitempty"`
    SPNameQualifier string `json:"sp_name_qualifier,omitempty"`
}
```

#### Conditions

```go
type Conditions struct {
    NotBefore           *time.Time `json:"not_before,omitempty"`
    NotOnOrAfter        *time.Time `json:"not_on_or_after,omitempty"`
    AudienceRestriction []string   `json:"audience_restriction,omitempty"`
}
```

#### Attribute

```go
type Attribute struct {
    Name         string   `json:"name"`
    FriendlyName string   `json:"friendly_name,omitempty"`
    NameFormat   string   `json:"name_format,omitempty"`
    Values       []string `json:"values"`
}
```

#### AuthnStatement

```go
type AuthnStatement struct {
    AuthnInstant         *time.Time `json:"authn_instant,omitempty"`
    SessionIndex         string     `json:"session_index,omitempty"`
    SessionNotOnOrAfter  *time.Time `json:"session_not_on_or_after,omitempty"`
    AuthnContextClassRef string     `json:"authn_context_class_ref,omitempty"`
}
```

#### SignatureInfo

```go
type SignatureInfo struct {
    Signed          bool             `json:"signed"`
    SignatureMethod string           `json:"signature_method,omitempty"`
    DigestMethod    string           `json:"digest_method,omitempty"`
    CertificateInfo *CertificateInfo `json:"certificate_info,omitempty"`
}
```

#### Status

```go
type Status struct {
    StatusCode    string `json:"status_code"`
    StatusMessage string `json:"status_message,omitempty"`
}
```

### Decoder

Handles base64 and deflate decoding.

```go
// Create a new decoder
decoder := saml.NewDecoder()

// Standard base64 decode
xmlData, err := decoder.Decode(base64String)

// Deflate + base64 decode (HTTP-Redirect binding)
xmlData, err := decoder.DecodeDeflate(deflatedBase64String)

// Auto-detect and decode
xmlData, err := decoder.SmartDecode(input)
```

### Decryptor

Handles encrypted SAML assertions.

```go
// Create decryptor with private key file
decryptor, err := saml.NewDecryptor("/path/to/private.pem")
if err != nil {
    // Handle error
}

// Decrypt XML data
decrypted, err := decryptor.Decrypt(encryptedXML)
```

### Parser

Parses SAML XML into structured data.

```go
parser := saml.NewParser()

info, err := parser.Parse(xmlData)
if err != nil {
    // Handle error
}

// Access parsed information
fmt.Println(info.Issuer)
fmt.Println(info.Subject.NameID)
for _, attr := range info.Attributes {
    fmt.Printf("%s: %v\n", attr.Name, attr.Values)
}
```

### Utility Functions

```go
// Check if XML contains encrypted assertion
isEncrypted := saml.IsEncrypted(xmlData)
```

---

## output Package

### Formatter

Formats output in various styles.

```go
// Create formatter with desired format
formatter := output.NewFormatter("pretty") // or "json", "xml"

// Format XML data
formatted, err := formatter.FormatXML(xmlData)

// Format SAMLInfo struct
formatted, err := formatter.FormatInfo(samlInfo)
```

### Supported Formats

| Format | Description |
|:-------|:------------|
| `pretty` | Colored, human-readable output |
| `json` | JSON serialization |
| `xml` | Formatted/indented XML |

---

## CLI Architecture

### Command Structure

```
samlurai (root)
├── decode     - Base64 decode
├── decrypt    - Decrypt with key
└── inspect    - Full pipeline
```

### Adding Custom Commands

If you fork SAMLurai, you can add custom commands:

```go
package cmd

import "github.com/spf13/cobra"

var customCmd = &cobra.Command{
    Use:   "custom",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
        return nil
    },
}

func init() {
    rootCmd.AddCommand(customCmd)
}
```

---

## Error Handling

SAMLurai uses wrapped errors for context:

```go
// Errors are wrapped with context
if err != nil {
    return fmt.Errorf("failed to decode SAML: %w", err)
}

// Check for specific error types
if errors.Is(err, base64.CorruptInputError(0)) {
    // Handle invalid base64
}
```

Common error scenarios:

| Error Message | Cause |
|:--------------|:------|
| `failed to decode SAML` | Invalid base64 encoding |
| `failed to load private key` | Invalid PEM file |
| `failed to decrypt` | Wrong key or unsupported algorithm |
| `failed to parse SAML` | Malformed XML |
| `not a valid SAML document` | XML is not SAML |

---

## JSON Output Schema

When using `-o json`, the output conforms to this schema:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "type": {
      "type": "string",
      "enum": ["Response", "Assertion"]
    },
    "id": { "type": "string" },
    "issue_instant": { "type": "string", "format": "date-time" },
    "destination": { "type": "string" },
    "in_response_to": { "type": "string" },
    "issuer": { "type": "string" },
    "status": {
      "type": "object",
      "properties": {
        "status_code": { "type": "string" },
        "status_message": { "type": "string" }
      }
    },
    "assertion": { "$ref": "#" },
    "subject": {
      "type": "object",
      "properties": {
        "name_id": { "type": "string" },
        "name_id_format": { "type": "string" }
      }
    },
    "conditions": {
      "type": "object",
      "properties": {
        "not_before": { "type": "string", "format": "date-time" },
        "not_on_or_after": { "type": "string", "format": "date-time" },
        "audience_restriction": {
          "type": "array",
          "items": { "type": "string" }
        }
      }
    },
    "attributes": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "friendly_name": { "type": "string" },
          "values": {
            "type": "array",
            "items": { "type": "string" }
          }
        }
      }
    }
  }
}
```
