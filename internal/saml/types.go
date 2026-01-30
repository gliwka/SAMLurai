package saml

import "time"

// SAMLInfo contains parsed information from a SAML assertion or response
type SAMLInfo struct {
	// Type indicates if this is a Response, Assertion, or AuthnRequest
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

	// Raw assertion (for responses containing assertions)
	Assertion *SAMLInfo `json:"assertion,omitempty"`

	// AuthnRequest-specific fields
	AssertionConsumerServiceURL string `json:"assertion_consumer_service_url,omitempty"`
	ProtocolBinding             string `json:"protocol_binding,omitempty"`
	ForceAuthn                  *bool  `json:"force_authn,omitempty"`
	IsPassive                   *bool  `json:"is_passive,omitempty"`
	NameIDPolicy                *NameIDPolicy `json:"name_id_policy,omitempty"`
	RequestedAttributes         []RequestedAttribute `json:"requested_attributes,omitempty"`
}

// NameIDPolicy contains the NameID policy for AuthnRequests
type NameIDPolicy struct {
	Format          string `json:"format,omitempty"`
	AllowCreate     *bool  `json:"allow_create,omitempty"`
	SPNameQualifier string `json:"sp_name_qualifier,omitempty"`
}

// RequestedAttribute contains information about a requested attribute in AuthnRequest
type RequestedAttribute struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name,omitempty"`
	NameFormat   string `json:"name_format,omitempty"`
	IsRequired   *bool  `json:"is_required,omitempty"`
}

// Status represents the SAML response status
type Status struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_message,omitempty"`
}

// Subject contains the subject information
type Subject struct {
	NameID          string `json:"name_id,omitempty"`
	NameIDFormat    string `json:"name_id_format,omitempty"`
	SPNameQualifier string `json:"sp_name_qualifier,omitempty"`
}

// Conditions contains the assertion conditions
type Conditions struct {
	NotBefore           *time.Time `json:"not_before,omitempty"`
	NotOnOrAfter        *time.Time `json:"not_on_or_after,omitempty"`
	AudienceRestriction []string   `json:"audience_restriction,omitempty"`
}

// AuthnStatement contains authentication statement information
type AuthnStatement struct {
	AuthnInstant         *time.Time `json:"authn_instant,omitempty"`
	SessionIndex         string     `json:"session_index,omitempty"`
	SessionNotOnOrAfter  *time.Time `json:"session_not_on_or_after,omitempty"`
	AuthnContextClassRef string     `json:"authn_context_class_ref,omitempty"`
}

// Attribute represents a SAML attribute
type Attribute struct {
	Name         string   `json:"name"`
	FriendlyName string   `json:"friendly_name,omitempty"`
	NameFormat   string   `json:"name_format,omitempty"`
	Values       []string `json:"values"`
}

// SignatureInfo contains information about the XML signature
type SignatureInfo struct {
	Signed          bool   `json:"signed"`
	SignatureMethod string `json:"signature_method,omitempty"`
	DigestMethod    string `json:"digest_method,omitempty"`
	CertificateInfo *CertificateInfo `json:"certificate_info,omitempty"`
}

// CertificateInfo contains information about the signing certificate
type CertificateInfo struct {
	Subject    string    `json:"subject,omitempty"`
	Issuer     string    `json:"issuer,omitempty"`
	NotBefore  time.Time `json:"not_before,omitempty"`
	NotAfter   time.Time `json:"not_after,omitempty"`
	Serial     string    `json:"serial,omitempty"`
}
