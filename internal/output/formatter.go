package output

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/gliwka/SAMLurai/internal/saml"
)

// Formatter handles output formatting for different formats
type Formatter struct {
	format string
	noColor bool
}

// NewFormatter creates a new formatter with the specified format
func NewFormatter(format string) *Formatter {
	return &Formatter{
		format: strings.ToLower(format),
	}
}

// NewFormatterWithOptions creates a formatter with additional options
func NewFormatterWithOptions(format string, noColor bool) *Formatter {
	return &Formatter{
		format:  strings.ToLower(format),
		noColor: noColor,
	}
}

// FormatXML formats XML data according to the configured format
func (f *Formatter) FormatXML(data []byte) (string, error) {
	switch f.format {
	case "json":
		return f.xmlToJSON(data)
	case "xml", "raw":
		return f.prettyXML(data)
	case "pretty":
		return f.prettyXML(data)
	default:
		return f.prettyXML(data)
	}
}

// FormatSAMLInfo formats SAMLInfo according to the configured format
func (f *Formatter) FormatSAMLInfo(info *saml.SAMLInfo) (string, error) {
	switch f.format {
	case "json":
		return f.toJSON(info)
	case "xml":
		return f.toXML(info)
	case "pretty":
		return f.toPretty(info)
	default:
		return f.toPretty(info)
	}
}

func (f *Formatter) prettyXML(data []byte) (string, error) {
	var buf bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewReader(data))
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		if err := encoder.EncodeToken(token); err != nil {
			return "", fmt.Errorf("failed to encode token: %w", err)
		}
	}

	if err := encoder.Flush(); err != nil {
		return "", fmt.Errorf("failed to flush encoder: %w", err)
	}

	return buf.String() + "\n", nil
}

func (f *Formatter) xmlToJSON(data []byte) (string, error) {
	// For XML to JSON, we'll parse as SAML first
	parser := saml.NewParser()
	info, err := parser.Parse(data)
	if err != nil {
		// If it fails, return a simple structure
		return f.toJSON(map[string]string{"raw_xml": string(data)})
	}
	return f.toJSON(info)
}

func (f *Formatter) toJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data) + "\n", nil
}

func (f *Formatter) toXML(v interface{}) (string, error) {
	data, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}
	return xml.Header + string(data) + "\n", nil
}

func (f *Formatter) toPretty(info *saml.SAMLInfo) (string, error) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	// Color setup
	headerColor := color.New(color.FgCyan, color.Bold)
	labelColor := color.New(color.FgYellow)
	valueColor := color.New(color.FgWhite)
	successColor := color.New(color.FgGreen)
	warnColor := color.New(color.FgRed)

	if f.noColor {
		color.NoColor = true
	}

	// Header
	headerColor.Fprintf(w, "═══════════════════════════════════════════════════════════════\n")
	headerColor.Fprintf(w, " SAML %s\n", info.Type)
	headerColor.Fprintf(w, "═══════════════════════════════════════════════════════════════\n\n")

	// Basic Info
	f.printSection(w, headerColor, "Basic Information")
	f.printField(w, labelColor, valueColor, "ID", info.ID)
	f.printField(w, labelColor, valueColor, "Issuer", info.Issuer)
	if info.IssueInstant != nil {
		f.printField(w, labelColor, valueColor, "Issue Instant", info.IssueInstant.Format(time.RFC3339))
	}
	if info.Destination != "" {
		f.printField(w, labelColor, valueColor, "Destination", info.Destination)
	}
	if info.InResponseTo != "" {
		f.printField(w, labelColor, valueColor, "In Response To", info.InResponseTo)
	}
	fmt.Fprintln(w)

	// Status (for responses)
	if info.Status != nil {
		f.printSection(w, headerColor, "Status")
		statusColor := successColor
		if info.Status.StatusCode != "Success" {
			statusColor = warnColor
		}
		statusColor.Fprintf(w, "  Status Code:\t%s\n", info.Status.StatusCode)
		if info.Status.StatusMessage != "" {
			f.printField(w, labelColor, valueColor, "Message", info.Status.StatusMessage)
		}
		fmt.Fprintln(w)
	}

	// AuthnRequest-specific fields
	if info.AssertionConsumerServiceURL != "" {
		f.printSection(w, headerColor, "Request Details")
		f.printField(w, labelColor, valueColor, "ACS URL", info.AssertionConsumerServiceURL)
		if info.ProtocolBinding != "" {
			f.printField(w, labelColor, valueColor, "Protocol Binding", f.shortenURI(info.ProtocolBinding))
		}
		if info.ForceAuthn != nil {
			f.printField(w, labelColor, valueColor, "Force Authn", fmt.Sprintf("%v", *info.ForceAuthn))
		}
		if info.IsPassive != nil {
			f.printField(w, labelColor, valueColor, "Is Passive", fmt.Sprintf("%v", *info.IsPassive))
		}
		fmt.Fprintln(w)
	}

	// NameID Policy (for AuthnRequest)
	if info.NameIDPolicy != nil {
		f.printSection(w, headerColor, "NameID Policy")
		if info.NameIDPolicy.Format != "" {
			f.printField(w, labelColor, valueColor, "Format", f.shortenURI(info.NameIDPolicy.Format))
		}
		if info.NameIDPolicy.AllowCreate != nil {
			f.printField(w, labelColor, valueColor, "Allow Create", fmt.Sprintf("%v", *info.NameIDPolicy.AllowCreate))
		}
		if info.NameIDPolicy.SPNameQualifier != "" {
			f.printField(w, labelColor, valueColor, "SP Name Qualifier", info.NameIDPolicy.SPNameQualifier)
		}
		fmt.Fprintln(w)
	}

	// Requested Attributes (for AuthnRequest)
	if len(info.RequestedAttributes) > 0 {
		f.printSection(w, headerColor, "Requested Attributes")
		for _, attr := range info.RequestedAttributes {
			name := attr.FriendlyName
			if name == "" {
				name = f.shortenURI(attr.Name)
			}
			required := ""
			if attr.IsRequired != nil && *attr.IsRequired {
				required = " (required)"
			}
			f.printField(w, labelColor, valueColor, name, f.shortenURI(attr.Name)+required)
		}
		fmt.Fprintln(w)
	}

	// Subject
	if info.Subject != nil {
		f.printSection(w, headerColor, "Subject")
		f.printField(w, labelColor, valueColor, "NameID", info.Subject.NameID)
		if info.Subject.NameIDFormat != "" {
			f.printField(w, labelColor, valueColor, "Format", f.shortenURI(info.Subject.NameIDFormat))
		}
		if info.Subject.SPNameQualifier != "" {
			f.printField(w, labelColor, valueColor, "SP Name Qualifier", info.Subject.SPNameQualifier)
		}
		fmt.Fprintln(w)
	}

	// Conditions
	if info.Conditions != nil {
		f.printSection(w, headerColor, "Conditions")
		if info.Conditions.NotBefore != nil {
			f.printField(w, labelColor, valueColor, "Not Before", info.Conditions.NotBefore.Format(time.RFC3339))
		}
		if info.Conditions.NotOnOrAfter != nil {
			f.printField(w, labelColor, valueColor, "Not On Or After", info.Conditions.NotOnOrAfter.Format(time.RFC3339))
		}
		if len(info.Conditions.AudienceRestriction) > 0 {
			f.printField(w, labelColor, valueColor, "Audiences", strings.Join(info.Conditions.AudienceRestriction, ", "))
		}
		fmt.Fprintln(w)
	}

	// Authentication Statement
	if info.AuthnStatement != nil {
		f.printSection(w, headerColor, "Authentication")
		if info.AuthnStatement.AuthnInstant != nil {
			f.printField(w, labelColor, valueColor, "Auth Instant", info.AuthnStatement.AuthnInstant.Format(time.RFC3339))
		}
		if info.AuthnStatement.SessionIndex != "" {
			f.printField(w, labelColor, valueColor, "Session Index", info.AuthnStatement.SessionIndex)
		}
		if info.AuthnStatement.AuthnContextClassRef != "" {
			f.printField(w, labelColor, valueColor, "Auth Context", f.shortenURI(info.AuthnStatement.AuthnContextClassRef))
		}
		fmt.Fprintln(w)
	}

	// Attributes
	if len(info.Attributes) > 0 {
		f.printSection(w, headerColor, "Attributes")
		for _, attr := range info.Attributes {
			name := attr.Name
			if attr.FriendlyName != "" {
				name = attr.FriendlyName + " (" + f.shortenURI(attr.Name) + ")"
			}
			f.printField(w, labelColor, valueColor, name, strings.Join(attr.Values, ", "))
		}
		fmt.Fprintln(w)
	}

	// Signature
	if info.Signature != nil {
		f.printSection(w, headerColor, "Signature")
		if info.Signature.Signed {
			successColor.Fprintf(w, "  Signed:\tYes\n")
		} else {
			warnColor.Fprintf(w, "  Signed:\tNo\n")
		}
		if info.Signature.SignatureMethod != "" {
			f.printField(w, labelColor, valueColor, "Signature Method", f.shortenURI(info.Signature.SignatureMethod))
		}
		if info.Signature.DigestMethod != "" {
			f.printField(w, labelColor, valueColor, "Digest Method", f.shortenURI(info.Signature.DigestMethod))
		}
		if info.Signature.CertificateInfo != nil {
			fmt.Fprintln(w)
			f.printField(w, labelColor, valueColor, "Cert Subject", info.Signature.CertificateInfo.Subject)
			f.printField(w, labelColor, valueColor, "Cert Issuer", info.Signature.CertificateInfo.Issuer)
			f.printField(w, labelColor, valueColor, "Cert Valid From", info.Signature.CertificateInfo.NotBefore.Format(time.RFC3339))
			f.printField(w, labelColor, valueColor, "Cert Valid Until", info.Signature.CertificateInfo.NotAfter.Format(time.RFC3339))
		}
		fmt.Fprintln(w)
	}

	// Nested Assertion
	if info.Assertion != nil {
		headerColor.Fprintf(w, "───────────────────────────────────────────────────────────────\n")
		headerColor.Fprintf(w, " Embedded Assertion\n")
		headerColor.Fprintf(w, "───────────────────────────────────────────────────────────────\n")
		
		nested, _ := f.toPretty(info.Assertion)
		fmt.Fprint(w, nested)
	}

	w.Flush()
	return buf.String(), nil
}

func (f *Formatter) printSection(w *tabwriter.Writer, c *color.Color, title string) {
	c.Fprintf(w, "▸ %s\n", title)
}

func (f *Formatter) printField(w *tabwriter.Writer, labelColor, valueColor *color.Color, label, value string) {
	labelColor.Fprintf(w, "  %s:\t", label)
	valueColor.Fprintf(w, "%s\n", value)
}

func (f *Formatter) shortenURI(uri string) string {
	// Shorten common SAML URIs for readability
	replacements := map[string]string{
		"urn:oasis:names:tc:SAML:2.0:nameid-format:": "",
		"urn:oasis:names:tc:SAML:2.0:ac:classes:":    "",
		"urn:oasis:names:tc:SAML:2.0:attrname-format:": "",
		"http://www.w3.org/2001/04/xmldsig-more#":    "",
		"http://www.w3.org/2000/09/xmldsig#":         "",
		"http://www.w3.org/2001/04/xmlenc#":          "",
	}

	for prefix, replacement := range replacements {
		if strings.HasPrefix(uri, prefix) {
			return replacement + strings.TrimPrefix(uri, prefix)
		}
	}

	return uri
}
