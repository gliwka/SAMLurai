---
layout: default
title: SAML Basics
parent: Guides
nav_order: 3
---

# SAML Basics
{: .no_toc }

A quick introduction to SAML concepts for developers.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## What is SAML?

**Security Assertion Markup Language (SAML)** is an XML-based standard for exchanging authentication and authorization data between parties, specifically:

- **Identity Provider (IdP)**: Authenticates users (e.g., Okta, Azure AD, OneLogin)
- **Service Provider (SP)**: Your application that needs to verify user identity

SAML enables **Single Sign-On (SSO)**—users authenticate once with the IdP and can access multiple SPs without re-authenticating.

## SAML Flow

### SP-Initiated SSO (Most Common)

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  User    │     │    SP    │     │   IdP    │
│ Browser  │     │ (Your    │     │ (Okta,   │
│          │     │  App)    │     │  Azure)  │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     │ 1. Access app  │                │
     │───────────────▶│                │
     │                │                │
     │ 2. Redirect to IdP              │
     │◀───────────────│                │
     │                                 │
     │ 3. SAML Request                 │
     │────────────────────────────────▶│
     │                                 │
     │         4. User authenticates   │
     │◀───────────────────────────────▶│
     │                                 │
     │ 5. SAML Response (POST)         │
     │◀────────────────────────────────│
     │                │                │
     │ 6. POST to ACS │                │
     │───────────────▶│                │
     │                │                │
     │ 7. Access granted               │
     │◀───────────────│                │
     │                │                │
```

1. User tries to access your application
2. SP generates SAML Request and redirects to IdP
3. Browser delivers SAML Request to IdP
4. User authenticates at IdP
5. IdP generates SAML Response with user info
6. Browser POSTs SAML Response to SP's ACS
7. SP validates response and grants access

### IdP-Initiated SSO

User starts at the IdP portal and clicks an app icon:

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  User    │     │   IdP    │     │    SP    │
│ Browser  │     │          │     │          │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     │ 1. Click app   │                │
     │───────────────▶│                │
     │                │                │
     │ 2. SAML Response                │
     │◀───────────────│                │
     │                                 │
     │ 3. POST to ACS                  │
     │────────────────────────────────▶│
     │                                 │
     │ 4. Access granted               │
     │◀────────────────────────────────│
```

## Key SAML Components

### SAML Request (AuthnRequest)

Sent from SP to IdP to request authentication:

```xml
<samlp:AuthnRequest
    xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
    ID="_abc123"
    Version="2.0"
    IssueInstant="2024-01-15T10:00:00Z"
    Destination="https://idp.example.com/sso"
    AssertionConsumerServiceURL="https://sp.example.com/acs">
    <saml:Issuer>https://sp.example.com</saml:Issuer>
</samlp:AuthnRequest>
```

Key elements:
- **ID**: Unique identifier for correlating with response
- **Destination**: IdP's SSO endpoint
- **AssertionConsumerServiceURL**: Where to send the response
- **Issuer**: SP's entity ID

### SAML Response

Sent from IdP to SP containing the authentication result:

```xml
<samlp:Response
    xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
    ID="_response123"
    InResponseTo="_abc123"
    Destination="https://sp.example.com/acs"
    IssueInstant="2024-01-15T10:01:00Z">
    <saml:Issuer>https://idp.example.com</saml:Issuer>
    <samlp:Status>
        <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
    </samlp:Status>
    <saml:Assertion>
        <!-- User information here -->
    </saml:Assertion>
</samlp:Response>
```

### SAML Assertion

The core of SAML—contains statements about the user:

```xml
<saml:Assertion
    xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
    ID="_assertion456"
    IssueInstant="2024-01-15T10:01:00Z"
    Version="2.0">
    
    <saml:Issuer>https://idp.example.com</saml:Issuer>
    
    <!-- Who the assertion is about -->
    <saml:Subject>
        <saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">
            user@example.com
        </saml:NameID>
    </saml:Subject>
    
    <!-- When/for whom it's valid -->
    <saml:Conditions
        NotBefore="2024-01-15T10:00:00Z"
        NotOnOrAfter="2024-01-15T10:05:00Z">
        <saml:AudienceRestriction>
            <saml:Audience>https://sp.example.com</saml:Audience>
        </saml:AudienceRestriction>
    </saml:Conditions>
    
    <!-- How the user authenticated -->
    <saml:AuthnStatement AuthnInstant="2024-01-15T10:00:30Z">
        <saml:AuthnContext>
            <saml:AuthnContextClassRef>
                urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport
            </saml:AuthnContextClassRef>
        </saml:AuthnContext>
    </saml:AuthnStatement>
    
    <!-- User attributes -->
    <saml:AttributeStatement>
        <saml:Attribute Name="email">
            <saml:AttributeValue>user@example.com</saml:AttributeValue>
        </saml:Attribute>
        <saml:Attribute Name="groups">
            <saml:AttributeValue>admin</saml:AttributeValue>
            <saml:AttributeValue>users</saml:AttributeValue>
        </saml:Attribute>
    </saml:AttributeStatement>
</saml:Assertion>
```

## SAML Bindings

How SAML messages are transported:

### HTTP-POST Binding

- Used for **Responses** (most common)
- Data in HTML form, auto-submitted via JavaScript
- Base64-encoded (no compression)

```html
<form method="post" action="https://sp.example.com/acs">
    <input type="hidden" name="SAMLResponse" value="PHNhbWxwOlJlc3BvbnNl..."/>
    <input type="submit" value="Submit"/>
</form>
<script>document.forms[0].submit();</script>
```

Use SAMLurai:
```bash
echo "PHNhbWxwOlJlc3BvbnNl..." | samlurai inspect
```

### HTTP-Redirect Binding

- Used for **Requests** (most common)
- Data in URL query parameter
- Deflate-compressed, then Base64-encoded, then URL-encoded

```
https://idp.example.com/sso?SAMLRequest=fZJNT8Mw...&RelayState=abc
```

Use SAMLurai:
```bash
# URL-decode first if needed, then:
samlurai decode --deflate "fZJNT8Mw..."
```

## NameID Formats

How the user is identified:

| Format | Example | Use Case |
|:-------|:--------|:---------|
| `emailAddress` | `user@example.com` | Most common |
| `persistent` | `_abc123xyz` | Persistent pseudonym |
| `transient` | `_temp789` | Session-specific |
| `unspecified` | Varies | Application decides |

## Common Attributes

Standard attribute names:

| Attribute | Description |
|:----------|:------------|
| `email` / `mail` | Email address |
| `givenName` | First name |
| `sn` / `surname` | Last name |
| `displayName` | Full display name |
| `groups` / `memberOf` | Group memberships |
| `uid` / `username` | Username |

{: .note }
Attribute names vary by IdP. Always check your IdP's documentation.

## Security Considerations

### Signatures

SAML messages should be signed to ensure integrity:

- **Response signature**: Signs the entire response
- **Assertion signature**: Signs just the assertion
- Usually both are signed

### Encryption

Sensitive assertions can be encrypted:

- Uses SP's public key
- Protects data in transit and at rest in browser

### Validation Checklist

SPs should validate:

1. ✅ Signature is valid
2. ✅ Issuer matches expected IdP
3. ✅ Destination matches our ACS URL
4. ✅ Audience includes our entity ID
5. ✅ Current time is within NotBefore/NotOnOrAfter
6. ✅ InResponseTo matches our request ID (for SP-initiated)
7. ✅ Assertion ID hasn't been seen before (replay protection)

## Using SAMLurai for Learning

### Inspect a Real Response

```bash
# Capture from browser and inspect
pbpaste | samlurai inspect
```

### View the Raw Structure

```bash
samlurai inspect -f response.xml -o xml
```

### Extract Specific Data

```bash
# Get all attributes
samlurai inspect -f response.xml -o json | jq '.assertion.attributes'

# Get the subject
samlurai inspect -f response.xml -o json | jq '.assertion.subject'
```

## Further Reading

- [OASIS SAML 2.0 Specification](https://docs.oasis-open.org/security/saml/v2.0/)
- [SAML Technical Overview](https://docs.oasis-open.org/security/saml/Post2.0/sstc-saml-tech-overview-2.0.html)

## See Also

- [Debugging SSO]({% link guides/debugging-sso.md %})
- [Encrypted Assertions]({% link guides/encrypted-assertions.md %})
