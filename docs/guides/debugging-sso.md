---
layout: default
title: Debugging SSO
parent: Guides
nav_order: 1
---

# Debugging SSO Issues
{: .no_toc }

A step-by-step guide to debugging SAML-based Single Sign-On.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Capturing SAML Traffic

### The Easy Way: HAR Files (Recommended)

The easiest way to capture SAML traffic is to save a HAR (HTTP Archive) file from the browser. This captures the entire SSO flow, including both the AuthnRequest and Response.

**Why HAR files are great for debugging:**
- Captures the complete SSO flow in one file
- Users can easily send you HAR files for support
- No need to manually copy/paste individual SAML messages
- Preserves request order and timing
- Works across all browsers

#### Capturing a HAR File

**Chrome / Edge:**
1. Open Developer Tools (F12 or Cmd+Option+I on Mac)
2. Go to the **Network** tab
3. Check **"Preserve log"** to keep requests across redirects
4. Perform the SSO login flow
5. Click the **Export HAR** (⬇️) button in the toolbar, or right-click any request and select **"Save all as HAR"**

{: .note }
Chrome offers two export options: "Save all as HAR (sanitized)" removes sensitive headers like cookies, while "Save all as HAR (with sensitive data)" keeps everything. For SAML debugging, you typically need the full data.

**Firefox:**
1. Open Developer Tools (F12)
2. Go to the **Network** tab  
3. Check **"Persist Logs"** in the gear menu
4. Perform the SSO login flow
5. Right-click any request → **"Save All As HAR"**

**Safari:**
1. Enable Developer menu (Settings → Advanced → Show features for web developers)
2. Open Web Inspector (Cmd+Option+I)
3. Go to the **Network** tab
4. Perform the SSO login flow
5. Click **Export** in the upper right

#### Inspecting HAR Files

Once you have the HAR file, inspect it directly:

```bash
samlurai inspect -f sso-flow.har
```

This shows all SAML messages in chronological order:

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
  Destination:    https://idp.example.com/sso
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

#### Extract Individual Files from HAR

Need to analyze each SAML message separately?

```bash
# Preview what's in the HAR
samlurai extract --list -f sso-flow.har

# Extract to separate files
samlurai extract -f sso-flow.har -d ./debug
```

### Support Workflows: Getting HARs from Users

{: .tip }
**For support teams:** Ask users to send you a HAR file instead of screenshots or copied text. It's easier for them and gives you complete information.

**Instructions to give users:**

> **How to capture a HAR file for troubleshooting:**
> 1. Open your browser's Developer Tools (press F12 or right-click → Inspect)
> 2. Click the **Network** tab
> 3. Check "Preserve log" (keeps data when page changes)
> 4. Try to log in again
> 5. Click the **Export HAR** button (download icon ⬇️) in the toolbar
> 6. Send us the downloaded .har file

Once you receive the HAR file:

```bash
# Quick overview
samlurai inspect -f user-capture.har

# If encrypted, add your key
samlurai inspect -f user-capture.har -k private.pem

# Extract for detailed analysis
samlurai extract -f user-capture.har -d ./support-case-123
```

### Alternative: Copy Individual SAML Messages

If you can't use HAR files, you can still copy individual SAML messages:

1. Open Developer Tools (F12 or Cmd+Option+I)
2. Go to the **Network** tab
3. Find the POST request to your ACS endpoint
4. Look at the **Payload** or **Form Data** tab
5. Copy the `SAMLResponse` value

```bash
# From clipboard
pbpaste | samlurai inspect

# Or save to file first
pbpaste > saml_response.txt
samlurai inspect -f saml_response.txt
```

{: .note }
The SAMLResponse is base64-encoded. SAMLurai handles decoding automatically.

### Using a Browser Extension

SAML debugging browser extensions can help capture messages:
- SAML-tracer (Firefox)
- SAML Chrome Panel (Chrome)

However, HAR files are often more convenient for sharing and archiving.

## Analyzing SAML Messages

### Quick Inspection

For HAR files (recommended):
```bash
samlurai inspect -f capture.har
```

For individual SAML files:
```bash
samlurai inspect -f response.xml
```

### What to Check

#### 1. Status Code

First, check if the authentication was successful:

```bash
samlurai inspect -f response.xml -o json | jq '.status'
```

Expected successful response:
```json
{
  "status_code": "urn:oasis:names:tc:SAML:2.0:status:Success"
}
```

Common error codes:

| Status Code | Meaning |
|:------------|:--------|
| `Success` | Authentication successful |
| `Requester` | Problem with the request |
| `Responder` | Problem at the IdP |
| `AuthnFailed` | Authentication failed |
| `NoPassive` | Cannot authenticate passively |

#### 2. Issuer

Verify the response comes from the expected IdP:

```bash
samlurai inspect -f response.xml -o json | jq '.issuer'
```

#### 3. Destination

Check the response was intended for your service:

```bash
samlurai inspect -f response.xml -o json | jq '.destination'
```

This should match your ACS URL exactly.

#### 4. Conditions

Check the assertion validity window:

```bash
samlurai inspect -f response.xml -o json | jq '.assertion.conditions'
```

```json
{
  "not_before": "2024-01-15T10:30:00Z",
  "not_on_or_after": "2024-01-15T10:35:00Z",
  "audience_restriction": ["https://your-sp.example.com"]
}
```

{: .warning }
SAML assertions typically have a short validity window (often 5 minutes). Ensure your server time is synchronized.

#### 5. Subject

Verify the user identifier:

```bash
samlurai inspect -f response.xml -o json | jq '.assertion.subject'
```

```json
{
  "name_id": "user@example.com",
  "name_id_format": "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
}
```

#### 6. Attributes

Check what user attributes were provided:

```bash
samlurai inspect -f response.xml -o json | jq '.assertion.attributes'
```

## Common Issues

### Clock Skew

**Symptom**: Assertion appears invalid or expired immediately.

**Diagnosis**:
```bash
# Check assertion times
samlurai inspect -f response.xml -o json | jq '.assertion.conditions'

# Compare with current time
date -u
```

**Solution**: Synchronize server time using NTP.

### Wrong Audience

**Symptom**: "Invalid audience" or "Audience mismatch" errors.

**Diagnosis**:
```bash
samlurai inspect -f response.xml -o json | jq '.assertion.conditions.audience_restriction'
```

**Solution**: Configure the IdP with the correct SP entity ID.

### Missing Attributes

**Symptom**: User attributes not appearing in your application.

**Diagnosis**:
```bash
samlurai inspect -f response.xml -o json | jq '.assertion.attributes | map(.name)'
```

**Solution**: Configure attribute release at the IdP.

### Signature Issues

**Symptom**: "Signature validation failed" errors.

**Diagnosis**:
```bash
samlurai inspect -f response.xml -o json | jq '.signature'
```

Check:
- Certificate has not expired
- Signature algorithm is supported
- Response or assertion (or both) is signed

### Encrypted Assertion

**Symptom**: Cannot see assertion contents.

**Diagnosis**:
```bash
# This will tell you if it's encrypted
samlurai inspect -f response.xml
```

**Solution**: Provide your private key:
```bash
samlurai inspect -f response.xml -k private.pem
```

## Debugging Workflow

### Quick Workflow with HAR Files

```bash
# 1. Get the HAR file from user or capture it yourself
# 2. Quick inspection - see the full SSO flow
samlurai inspect -f capture.har

# 3. If assertions are encrypted, add your key
samlurai inspect -f capture.har -k private.pem

# 4. Extract for detailed analysis
samlurai extract -f capture.har -d ./debug

# 5. Analyze specific messages
samlurai inspect -f ./debug/saml_001_authnrequest.xml
samlurai inspect -f ./debug/saml_002_response.xml -o json | jq '.status'
```

### Detailed Workflow with Individual Files

```bash
# 1. Save the SAML Response
pbpaste > saml_response.txt

# 2. Quick inspection (see overview)
samlurai inspect -f saml_response.txt

# 3. Check status
samlurai inspect -f saml_response.txt -o json | jq '.status'

# 4. Check times
samlurai inspect -f saml_response.txt -o json | \
  jq '{issue_instant, conditions: .assertion.conditions}'

# 5. Check attributes
samlurai inspect -f saml_response.txt -o json | \
  jq '.assertion.attributes'

# 6. Get raw XML for deeper inspection
samlurai inspect -f saml_response.txt -o xml > decoded.xml
```

### Comparison Script

Create a script to compare expected vs actual values:

```bash
#!/bin/bash
RESPONSE=$1
EXPECTED_ISSUER="https://idp.example.com"
EXPECTED_AUDIENCE="https://sp.example.com"

echo "=== SAML Response Analysis ==="

# Check issuer
ACTUAL_ISSUER=$(samlurai inspect -f "$RESPONSE" -o json | jq -r '.issuer')
if [ "$ACTUAL_ISSUER" = "$EXPECTED_ISSUER" ]; then
    echo "✓ Issuer matches"
else
    echo "✗ Issuer mismatch: expected $EXPECTED_ISSUER, got $ACTUAL_ISSUER"
fi

# Check audience
ACTUAL_AUDIENCE=$(samlurai inspect -f "$RESPONSE" -o json | jq -r '.assertion.conditions.audience_restriction[0]')
if [ "$ACTUAL_AUDIENCE" = "$EXPECTED_AUDIENCE" ]; then
    echo "✓ Audience matches"
else
    echo "✗ Audience mismatch: expected $EXPECTED_AUDIENCE, got $ACTUAL_AUDIENCE"
fi

# Check validity
NOT_AFTER=$(samlurai inspect -f "$RESPONSE" -o json | jq -r '.assertion.conditions.not_on_or_after')
echo "Assertion valid until: $NOT_AFTER"
```

## See Also

- [Working with Encrypted Assertions]({% link guides/encrypted-assertions.md %})
- [SAML Basics]({% link guides/saml-basics.md %})
- [`inspect` command reference]({% link commands/inspect.md %})
