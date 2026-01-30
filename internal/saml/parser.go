package saml

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// Parser handles parsing of SAML XML documents
type Parser struct{}

// NewParser creates a new SAML parser
func NewParser() *Parser {
	return &Parser{}
}

// XML namespace constants
const (
	SAMLPNamespace   = "urn:oasis:names:tc:SAML:2.0:protocol"
	SAMLNamespace    = "urn:oasis:names:tc:SAML:2.0:assertion"
	XMLDSigNamespace = "http://www.w3.org/2000/09/xmldsig#"
)

// SAML Response structure for XML parsing
type samlResponse struct {
	XMLName      xml.Name          `xml:"Response"`
	ID           string            `xml:"ID,attr"`
	IssueInstant string            `xml:"IssueInstant,attr"`
	Destination  string            `xml:"Destination,attr"`
	InResponseTo string            `xml:"InResponseTo,attr"`
	Issuer       string            `xml:"Issuer"`
	Status       samlStatus        `xml:"Status"`
	Assertion    *samlAssertion    `xml:"Assertion"`
	Signature    *xmldsigSignature `xml:"Signature"`
}

type samlStatus struct {
	StatusCode struct {
		Value string `xml:"Value,attr"`
	} `xml:"StatusCode"`
	StatusMessage string `xml:"StatusMessage"`
}

type samlAssertion struct {
	XMLName            xml.Name                `xml:"Assertion"`
	ID                 string                  `xml:"ID,attr"`
	IssueInstant       string                  `xml:"IssueInstant,attr"`
	Issuer             string                  `xml:"Issuer"`
	Subject            *samlSubject            `xml:"Subject"`
	Conditions         *samlConditions         `xml:"Conditions"`
	AuthnStatement     *samlAuthnStatement     `xml:"AuthnStatement"`
	AttributeStatement *samlAttributeStatement `xml:"AttributeStatement"`
	Signature          *xmldsigSignature       `xml:"Signature"`
}

type samlSubject struct {
	NameID struct {
		Value           string `xml:",chardata"`
		Format          string `xml:"Format,attr"`
		SPNameQualifier string `xml:"SPNameQualifier,attr"`
	} `xml:"NameID"`
}

type samlConditions struct {
	NotBefore           string `xml:"NotBefore,attr"`
	NotOnOrAfter        string `xml:"NotOnOrAfter,attr"`
	AudienceRestriction struct {
		Audiences []string `xml:"Audience"`
	} `xml:"AudienceRestriction"`
}

type samlAuthnStatement struct {
	AuthnInstant        string `xml:"AuthnInstant,attr"`
	SessionIndex        string `xml:"SessionIndex,attr"`
	SessionNotOnOrAfter string `xml:"SessionNotOnOrAfter,attr"`
	AuthnContext        struct {
		AuthnContextClassRef string `xml:"AuthnContextClassRef"`
	} `xml:"AuthnContext"`
}

type samlAttributeStatement struct {
	Attributes []samlAttribute `xml:"Attribute"`
}

type samlAttribute struct {
	Name         string   `xml:"Name,attr"`
	FriendlyName string   `xml:"FriendlyName,attr"`
	NameFormat   string   `xml:"NameFormat,attr"`
	Values       []string `xml:"AttributeValue"`
}

type xmldsigSignature struct {
	SignedInfo struct {
		SignatureMethod struct {
			Algorithm string `xml:"Algorithm,attr"`
		} `xml:"SignatureMethod"`
		Reference struct {
			DigestMethod struct {
				Algorithm string `xml:"Algorithm,attr"`
			} `xml:"DigestMethod"`
		} `xml:"Reference"`
	} `xml:"SignedInfo"`
	KeyInfo struct {
		X509Data struct {
			X509Certificate string `xml:"X509Certificate"`
		} `xml:"X509Data"`
	} `xml:"KeyInfo"`
}

// AuthnRequest structure for XML parsing
type samlAuthnRequest struct {
	XMLName                     xml.Name          `xml:"AuthnRequest"`
	ID                          string            `xml:"ID,attr"`
	Version                     string            `xml:"Version,attr"`
	IssueInstant                string            `xml:"IssueInstant,attr"`
	Destination                 string            `xml:"Destination,attr"`
	AssertionConsumerServiceURL string            `xml:"AssertionConsumerServiceURL,attr"`
	ProtocolBinding             string            `xml:"ProtocolBinding,attr"`
	ForceAuthn                  string            `xml:"ForceAuthn,attr"`
	IsPassive                   string            `xml:"IsPassive,attr"`
	Issuer                      string            `xml:"Issuer"`
	NameIDPolicy                *samlNameIDPolicy `xml:"NameIDPolicy"`
	Signature                   *xmldsigSignature `xml:"Signature"`
	Extensions                  *samlExtensions   `xml:"Extensions"`
}

type samlNameIDPolicy struct {
	Format          string `xml:"Format,attr"`
	AllowCreate     string `xml:"AllowCreate,attr"`
	SPNameQualifier string `xml:"SPNameQualifier,attr"`
}

type samlExtensions struct {
	RequestedAttributes []samlRequestedAttribute `xml:"RequestedAttribute"`
}

type samlRequestedAttribute struct {
	Name         string `xml:"Name,attr"`
	FriendlyName string `xml:"FriendlyName,attr"`
	NameFormat   string `xml:"NameFormat,attr"`
	IsRequired   string `xml:"isRequired,attr"`
}

// Parse parses a SAML XML document and returns structured information
func (p *Parser) Parse(xmlData []byte) (*SAMLInfo, error) {
	// Try to detect the SAML message type
	trimmed := bytes.TrimSpace(xmlData)

	if bytes.Contains(trimmed, []byte("<samlp:Response")) || bytes.Contains(trimmed, []byte("<Response")) {
		return p.parseResponse(xmlData)
	}

	if bytes.Contains(trimmed, []byte("<samlp:AuthnRequest")) || bytes.Contains(trimmed, []byte("<AuthnRequest")) {
		return p.parseAuthnRequest(xmlData)
	}

	if bytes.Contains(trimmed, []byte("<saml:Assertion")) || bytes.Contains(trimmed, []byte("<Assertion")) {
		return p.parseAssertion(xmlData)
	}

	// Try parsing as Response first, then AuthnRequest, then Assertion
	info, err := p.parseResponse(xmlData)
	if err == nil {
		return info, nil
	}

	info, err = p.parseAuthnRequest(xmlData)
	if err == nil {
		return info, nil
	}

	return p.parseAssertion(xmlData)
}

func (p *Parser) parseAuthnRequest(xmlData []byte) (*SAMLInfo, error) {
	var req samlAuthnRequest
	if err := xml.Unmarshal(xmlData, &req); err != nil {
		return nil, fmt.Errorf("failed to parse SAML AuthnRequest: %w", err)
	}

	info := &SAMLInfo{
		Type:                        "AuthnRequest",
		ID:                          req.ID,
		Destination:                 req.Destination,
		Issuer:                      req.Issuer,
		AssertionConsumerServiceURL: req.AssertionConsumerServiceURL,
		ProtocolBinding:             req.ProtocolBinding,
	}

	// Parse IssueInstant
	if req.IssueInstant != "" {
		if t, err := time.Parse(time.RFC3339, req.IssueInstant); err == nil {
			info.IssueInstant = &t
		}
	}

	// Parse ForceAuthn
	if req.ForceAuthn != "" {
		val := strings.ToLower(req.ForceAuthn) == "true"
		info.ForceAuthn = &val
	}

	// Parse IsPassive
	if req.IsPassive != "" {
		val := strings.ToLower(req.IsPassive) == "true"
		info.IsPassive = &val
	}

	// Parse NameIDPolicy
	if req.NameIDPolicy != nil {
		info.NameIDPolicy = &NameIDPolicy{
			Format:          req.NameIDPolicy.Format,
			SPNameQualifier: req.NameIDPolicy.SPNameQualifier,
		}
		if req.NameIDPolicy.AllowCreate != "" {
			val := strings.ToLower(req.NameIDPolicy.AllowCreate) == "true"
			info.NameIDPolicy.AllowCreate = &val
		}
	}

	// Parse Signature
	if req.Signature != nil {
		info.Signature = p.parseSignature(req.Signature)
	}

	// Parse RequestedAttributes from Extensions
	if req.Extensions != nil {
		for _, attr := range req.Extensions.RequestedAttributes {
			ra := RequestedAttribute{
				Name:         attr.Name,
				FriendlyName: attr.FriendlyName,
				NameFormat:   attr.NameFormat,
			}
			if attr.IsRequired != "" {
				val := strings.ToLower(attr.IsRequired) == "true"
				ra.IsRequired = &val
			}
			info.RequestedAttributes = append(info.RequestedAttributes, ra)
		}
	}

	return info, nil
}

func (p *Parser) parseResponse(xmlData []byte) (*SAMLInfo, error) {
	var resp samlResponse
	if err := xml.Unmarshal(xmlData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	info := &SAMLInfo{
		Type:         "Response",
		ID:           resp.ID,
		Destination:  resp.Destination,
		InResponseTo: resp.InResponseTo,
		Issuer:       resp.Issuer,
	}

	// Parse IssueInstant
	if resp.IssueInstant != "" {
		if t, err := time.Parse(time.RFC3339, resp.IssueInstant); err == nil {
			info.IssueInstant = &t
		}
	}

	// Parse Status
	if resp.Status.StatusCode.Value != "" {
		info.Status = &Status{
			StatusCode:    p.extractStatusCode(resp.Status.StatusCode.Value),
			StatusMessage: resp.Status.StatusMessage,
		}
	}

	// Parse Signature
	if resp.Signature != nil {
		info.Signature = p.parseSignature(resp.Signature)
	}

	// Parse Assertion if present
	if resp.Assertion != nil {
		assertion, err := p.parseAssertionStruct(resp.Assertion)
		if err != nil {
			return nil, err
		}
		info.Assertion = assertion
	}

	return info, nil
}

// ParsePartial parses a SAML document and returns whatever information is available,
// even if some parts (like encrypted assertions) cannot be fully parsed.
// This is useful for showing partial information when decryption is not possible.
func (p *Parser) ParsePartial(xmlData []byte) (*SAMLInfo, error) {
	trimmed := bytes.TrimSpace(xmlData)

	// For responses with encrypted assertions, we can still show the response-level info
	if bytes.Contains(trimmed, []byte("<samlp:Response")) || bytes.Contains(trimmed, []byte("<Response")) {
		return p.parseResponsePartial(xmlData)
	}

	// For other types, use regular parsing
	return p.Parse(xmlData)
}

// parseResponsePartial parses a Response and returns available information
// even if the assertion is encrypted
func (p *Parser) parseResponsePartial(xmlData []byte) (*SAMLInfo, error) {
	var resp samlResponse
	if err := xml.Unmarshal(xmlData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	info := &SAMLInfo{
		Type:         "Response (Encrypted)",
		ID:           resp.ID,
		Destination:  resp.Destination,
		InResponseTo: resp.InResponseTo,
		Issuer:       resp.Issuer,
	}

	// Parse IssueInstant
	if resp.IssueInstant != "" {
		if t, err := time.Parse(time.RFC3339, resp.IssueInstant); err == nil {
			info.IssueInstant = &t
		}
	}

	// Parse Status
	if resp.Status.StatusCode.Value != "" {
		info.Status = &Status{
			StatusCode:    p.extractStatusCode(resp.Status.StatusCode.Value),
			StatusMessage: resp.Status.StatusMessage,
		}
	}

	// Parse Signature
	if resp.Signature != nil {
		info.Signature = p.parseSignature(resp.Signature)
	}

	return info, nil
}

func (p *Parser) parseAssertion(xmlData []byte) (*SAMLInfo, error) {
	var assertion samlAssertion
	if err := xml.Unmarshal(xmlData, &assertion); err != nil {
		return nil, fmt.Errorf("failed to parse SAML assertion: %w", err)
	}

	return p.parseAssertionStruct(&assertion)
}

func (p *Parser) parseAssertionStruct(assertion *samlAssertion) (*SAMLInfo, error) {
	info := &SAMLInfo{
		Type:   "Assertion",
		ID:     assertion.ID,
		Issuer: assertion.Issuer,
	}

	// Parse IssueInstant
	if assertion.IssueInstant != "" {
		if t, err := time.Parse(time.RFC3339, assertion.IssueInstant); err == nil {
			info.IssueInstant = &t
		}
	}

	// Parse Subject
	if assertion.Subject != nil {
		info.Subject = &Subject{
			NameID:          assertion.Subject.NameID.Value,
			NameIDFormat:    assertion.Subject.NameID.Format,
			SPNameQualifier: assertion.Subject.NameID.SPNameQualifier,
		}
	}

	// Parse Conditions
	if assertion.Conditions != nil {
		info.Conditions = &Conditions{}
		if assertion.Conditions.NotBefore != "" {
			if t, err := time.Parse(time.RFC3339, assertion.Conditions.NotBefore); err == nil {
				info.Conditions.NotBefore = &t
			}
		}
		if assertion.Conditions.NotOnOrAfter != "" {
			if t, err := time.Parse(time.RFC3339, assertion.Conditions.NotOnOrAfter); err == nil {
				info.Conditions.NotOnOrAfter = &t
			}
		}
		info.Conditions.AudienceRestriction = assertion.Conditions.AudienceRestriction.Audiences
	}

	// Parse AuthnStatement
	if assertion.AuthnStatement != nil {
		info.AuthnStatement = &AuthnStatement{
			SessionIndex:         assertion.AuthnStatement.SessionIndex,
			AuthnContextClassRef: assertion.AuthnStatement.AuthnContext.AuthnContextClassRef,
		}
		if assertion.AuthnStatement.AuthnInstant != "" {
			if t, err := time.Parse(time.RFC3339, assertion.AuthnStatement.AuthnInstant); err == nil {
				info.AuthnStatement.AuthnInstant = &t
			}
		}
		if assertion.AuthnStatement.SessionNotOnOrAfter != "" {
			if t, err := time.Parse(time.RFC3339, assertion.AuthnStatement.SessionNotOnOrAfter); err == nil {
				info.AuthnStatement.SessionNotOnOrAfter = &t
			}
		}
	}

	// Parse Attributes
	if assertion.AttributeStatement != nil {
		for _, attr := range assertion.AttributeStatement.Attributes {
			info.Attributes = append(info.Attributes, Attribute(attr))
		}
	}

	// Parse Signature
	if assertion.Signature != nil {
		info.Signature = p.parseSignature(assertion.Signature)
	}

	return info, nil
}

func (p *Parser) parseSignature(sig *xmldsigSignature) *SignatureInfo {
	sigInfo := &SignatureInfo{
		Signed:          true,
		SignatureMethod: sig.SignedInfo.SignatureMethod.Algorithm,
		DigestMethod:    sig.SignedInfo.Reference.DigestMethod.Algorithm,
	}

	// Try to parse certificate info
	if sig.KeyInfo.X509Data.X509Certificate != "" {
		certData := sig.KeyInfo.X509Data.X509Certificate
		// Clean up whitespace
		certData = strings.ReplaceAll(certData, "\n", "")
		certData = strings.ReplaceAll(certData, "\r", "")
		certData = strings.ReplaceAll(certData, " ", "")

		if certBytes, err := base64.StdEncoding.DecodeString(certData); err == nil {
			if cert, err := x509.ParseCertificate(certBytes); err == nil {
				sigInfo.CertificateInfo = &CertificateInfo{
					Subject:   cert.Subject.String(),
					Issuer:    cert.Issuer.String(),
					NotBefore: cert.NotBefore,
					NotAfter:  cert.NotAfter,
					Serial:    cert.SerialNumber.String(),
				}
			}
		}
	}

	return sigInfo
}

func (p *Parser) extractStatusCode(fullCode string) string {
	// Extract just the status code name from the full URI
	parts := strings.Split(fullCode, ":")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullCode
}
