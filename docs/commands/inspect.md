---
layout: default
title: inspect
parent: Commands
nav_order: 1
---

# inspect
{: .no_toc }

Parse and display SAML assertion details in a human-readable format.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Synopsis

```
samlurai inspect [flags]
```

The `inspect` command is the most powerful command in SAMLurai. It automatically decodes, decrypts, and parses SAML data to display all relevant information. It supports both single SAML files and HAR files containing complete SSO flows.

## Description

The inspect command provides a complete pipeline for SAML analysis:

1. **HAR Support**: Detects HAR files and extracts all SAML messages in order
2. **Auto-decode**: Detects and decodes base64 (with optional deflate)
3. **Auto-decrypt**: Decrypts encrypted assertions when a key is provided
4. **Parse**: Extracts all SAML information
5. **Display**: Shows human-readable output

This command displays:
- AuthnRequest details (for SSO initiation)
- Response/Assertion type and ID
- Issuer information
- Subject (NameID)
- Conditions (validity period, audience)
- Attributes (with values)
- Authentication statements
- Signature information

## Flags

| Flag | Short | Description | Default |
|:-----|:------|:------------|:--------|
| `--file` | `-f` | Read SAML from file (supports HAR and XML) | |
| `--key` | `-k` | Path to private key for decryption (PEM format) | |
| `--output` | `-o` | Output format: `pretty`, `json`, `xml` | `pretty` |
| `--help` | `-h` | Help for inspect | |

## Examples

### Inspect HAR file (recommended)

The easiest way to debug a complete SSO flow is to capture it as a HAR file:

```bash
samlurai inspect -f sso-capture.har
```

This shows all SAML messages (AuthnRequests and Responses) in chronological order with context about where each was found.

### Inspect from single file

```bash
samlurai inspect -f assertion.xml
```

### Inspect base64-encoded SAML

The command auto-decodes:

```bash
echo "PHNhbWw+Li4uPC9zYW1sPg==" | samlurai inspect
```

### Inspect encrypted assertion

Provide a private key with `-k`:

```bash
samlurai inspect -f encrypted.xml -k private.pem
```

### Inspect HAR with encrypted assertions

```bash
samlurai inspect -f capture.har -k private.pem
```

### Full pipeline from browser

Base64-encoded SAML from browser dev tools:

```bash
# Copy SAMLResponse from browser, then:
pbpaste | samlurai inspect -k private.pem
```

### Output as JSON

Perfect for scripting and automation:

```bash
samlurai inspect -f assertion.xml -o json
```

## HAR File Support

When you pass a HAR file, SAMLurai automatically:

1. Parses the HAR JSON structure
2. Finds all SAML messages in requests and responses
3. Extracts from POST body parameters, query strings, and response bodies
4. Displays each message with context (URL, parameter name, source)
5. Shows messages in the order they appear in the HAR

### Capturing a HAR File

**Chrome / Edge:**
1. Open DevTools (F12 or Cmd+Option+I)
2. Go to **Network** tab, check **"Preserve log"**
3. Perform the SSO login flow
4. Click **Export HAR** button (⬇️) or right-click → **"Save all as HAR"**

**Firefox:**
1. Open DevTools (F12), go to **Network** tab
2. Check **"Persist Logs"** in the gear menu
3. Perform the SSO login flow  
4. Right-click → **"Save All As HAR"**

### HAR Output Example

```
Found 2 SAML message(s) in HAR file:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 [1/2] AuthnRequest from request-body
       Parameter: SAMLRequest
       URL: https://idp.example.com/sso
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

═══════════════════════════════════════════════════════════════
 SAML AuthnRequest
═══════════════════════════════════════════════════════════════

▸ Basic Information
  ID:             _abc123
  Issuer:         https://sp.example.com
  ...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 [2/2] Response from request-body
       Parameter: SAMLResponse
       URL: https://sp.example.com/acs
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

═══════════════════════════════════════════════════════════════
 SAML Response
═══════════════════════════════════════════════════════════════
...
```

### Encrypted Assertions in HAR

If the HAR contains encrypted assertions and you don't provide a key, SAMLurai shows a helpful message and displays what it can (Response metadata):

```
⚠️  Encrypted assertion detected - provide -k flag to decrypt

═══════════════════════════════════════════════════════════════
 SAML Response
═══════════════════════════════════════════════════════════════

▸ Basic Information
  ID:              _response123
  Issuer:          https://idp.example.com
  Issue Instant:   2024-01-15T10:30:00Z
  ...

▸ Status
  Status Code:  Success

▸ Signature
  Signed:  Yes
  ...
```

## JSON Output Format

Example JSON output:

```json
{
  "type": "Response",
  "id": "_abc123",
  "issue_instant": "2024-01-15T10:30:00Z",
  "issuer": "https://idp.example.com",
  "status": {
    "status_code": "urn:oasis:names:tc:SAML:2.0:status:Success"
  },
  "assertion": {
    "type": "Assertion",
    "issuer": "https://idp.example.com",
    "subject": {
      "name_id": "user@example.com",
      "name_id_format": "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
    },
    "conditions": {
      "not_before": "2024-01-15T10:30:00Z",
      "not_on_or_after": "2024-01-15T10:35:00Z",
      "audience_restriction": ["https://sp.example.com"]
    },
    "attributes": [
      {
        "name": "email",
        "values": ["user@example.com"]
      },
      {
        "name": "groups",
        "values": ["admin", "users"]
      }
    ]
  }
}
```

## XML Output Format

Get formatted, indented XML:

```bash
samlurai inspect -f assertion.xml -o xml
```

## Pretty Output Format

The default `pretty` format provides a stylized, hierarchical output:

```
        ⚔️  SAMLurai  ⚔️
                  _
              o  / )
             /|)(  |
             /^\  /
            /   \/

═══════════════════════════════════════════════════════════════
 SAML Response
═══════════════════════════════════════════════════════════════

▸ Basic Information
  ID:              _response123
  Issuer:          https://idp.example.com
  Issue Instant:   2024-01-15T10:30:00Z
  Destination:     https://sp.example.com/acs
  In Response To:  _request456

▸ Status
  Status Code:  Success

───────────────────────────────────────────────────────────────
 Embedded Assertion
───────────────────────────────────────────────────────────────

▸ Subject
  NameID:             user@example.com
  Format:             emailAddress
  SP Name Qualifier:  https://sp.example.com

▸ Conditions
  Not Before:       2024-01-15T10:25:00Z
  Not On Or After:  2024-01-15T10:35:00Z
  Audiences:        https://sp.example.com

▸ Authentication
  Auth Instant:   2024-01-15T10:29:00Z
  Session Index:  _session123
  Auth Context:   PasswordProtectedTransport

▸ Attributes
  Email (email):           user@example.com
  First Name (firstName):  John
  Last Name (lastName):    Doe
  Groups (groups):         admins, users
```

## What Gets Displayed

### Response Information

| Field | Description |
|:------|:------------|
| ID | Unique identifier for the response |
| Issue Instant | When the response was created |
| Destination | Intended recipient URL (ACS) |
| In Response To | ID of the original request |
| Issuer | Identity Provider URL |
| Status | Success or error code |

### Subject Information

| Field | Description |
|:------|:------------|
| NameID | User identifier |
| Format | NameID format (email, persistent, transient, etc.) |
| SP Name Qualifier | Service Provider entity ID |

### Conditions

| Field | Description |
|:------|:------------|
| Not Before | Assertion valid from this time |
| Not On Or After | Assertion valid until this time |
| Audience Restriction | Intended Service Providers |

### Authentication Statement

| Field | Description |
|:------|:------------|
| Authn Instant | When the user authenticated |
| Session Index | Session identifier |
| Session Not On Or After | Session expiry time |
| Authn Context | How the user authenticated |

### Attributes

User attributes provided by the IdP, such as:
- Email
- Name (first, last, full)
- Groups/Roles
- Custom attributes

### Signature Information

| Field | Description |
|:------|:------------|
| Signed | Whether the assertion is signed |
| Signature Method | Algorithm used (e.g., RSA-SHA256) |
| Digest Method | Hash algorithm (e.g., SHA-256) |
| Certificate Info | Signing certificate details |

## Common Workflows

### Debugging SSO Issues

```bash
# 1. Get SAML Response from browser (Network tab → find POST to ACS)
# 2. Copy the SAMLResponse value
# 3. Inspect it
pbpaste | samlurai inspect
```

### Validating Assertions

```bash
# Check if assertion is still valid
samlurai inspect -f assertion.xml -o json | jq '.assertion.conditions'
```

### Extracting Attributes

```bash
# Get all attribute names and values
samlurai inspect -f response.xml -o json | jq '.assertion.attributes'
```

### Checking Encrypted Assertions

```bash
# First, see if it's encrypted
samlurai inspect -f response.xml
# If you see "encrypted assertion detected", add your key:
samlurai inspect -f response.xml -k private.pem
```

## Error Handling

| Error | Cause | Solution |
|:------|:------|:---------|
| `encrypted SAML detected but no private key provided` | Missing key | Add `-k private.pem` |
| `failed to parse SAML` | Invalid XML | Check the raw XML with `decode` |
| `not a valid SAML document` | Wrong document type | Ensure it's a SAML Response or Assertion |

## See Also

- [`decode`]({% link commands/decode.md %}) - Decode base64-encoded SAML
- [`decrypt`]({% link commands/decrypt.md %}) - Decrypt encrypted assertions
