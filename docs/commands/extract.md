---
layout: default
title: extract
parent: Commands
nav_order: 2
---

# extract
{: .no_toc }

Extract SAML messages from HAR files to individual files.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Synopsis

```
samlurai extract [flags]
```

The `extract` command parses HAR (HTTP Archive) files and extracts all SAML messages into separate XML files.

## Description

When debugging SAML flows, you often need to examine multiple SAML messages captured during an SSO session. The extract command:

1. **Parses HAR files** from browser DevTools
2. **Finds all SAML messages** (AuthnRequests, Responses, LogoutRequests)
3. **Decodes** base64-encoded content automatically
4. **Saves** each message as a separate, formatted XML file

This is useful when you want to:
- Archive SAML messages for later analysis
- Share specific SAML messages with others
- Process SAML files with other tools
- Keep a record of SSO flows for debugging

## Flags

| Flag | Short | Description | Default |
|:-----|:------|:------------|:--------|
| `--file` | `-f` | Path to HAR file (required) | |
| `--dir` | `-d` | Output directory for extracted files | current directory |
| `--list` | | List SAML messages without extracting | `false` |
| `--help` | `-h` | Help for extract | |

## Examples

### List SAML messages in HAR file

Preview what will be extracted without creating files:

```bash
samlurai extract --list -f capture.har
```

**Output:**
```
Found 2 SAML message(s):

  1. AuthnRequest
     Source: request-body (SAMLRequest)
     URL: https://idp.example.com/sso

  2. Response
     Source: request-body (SAMLResponse)
     URL: https://sp.example.com/acs
```

### Extract to current directory

```bash
samlurai extract -f capture.har
```

Creates files like:
- `saml_001_authnrequest.xml`
- `saml_002_response.xml`

### Extract to specific directory

```bash
samlurai extract -f capture.har -d ./extracted
```

### Extract multiple HAR files

```bash
for har in *.har; do
  samlurai extract -f "$har" -d "./extracted/$(basename "$har" .har)"
done
```

## Output File Naming

Files are named sequentially with the SAML message type:

| Pattern | Example | Description |
|:--------|:--------|:------------|
| `saml_NNN_authnrequest.xml` | `saml_001_authnrequest.xml` | SAML AuthnRequest |
| `saml_NNN_response.xml` | `saml_002_response.xml` | SAML Response |
| `saml_NNN_assertion.xml` | `saml_003_assertion.xml` | Standalone Assertion |
| `saml_NNN_logoutrequest.xml` | `saml_004_logoutrequest.xml` | Logout Request |

The numbering preserves the order in which messages appeared in the HAR file.

## Capturing HAR Files

### Chrome / Edge

1. Open DevTools (F12 or Cmd+Option+I)
2. Go to **Network** tab
3. Check **"Preserve log"** to keep requests across page navigation
4. Perform the SSO login flow
5. Click the **Export HAR** button (⬇️) in the toolbar, or right-click any request → **"Save all as HAR"**

{: .note }
Chrome offers "Save all as HAR (sanitized)" which removes sensitive headers. For SAML debugging, you may need the full data version.

### Firefox

1. Open DevTools (F12)
2. Go to **Network** tab
3. Check **"Persist Logs"** in the gear menu (⚙️)
4. Perform the SSO login flow
5. Right-click any request → **"Save All As HAR"**

### Safari

1. Enable Developer menu (Settings → Advanced → Show features for web developers)
2. Open Web Inspector (Cmd+Option+I)
3. Go to **Network** tab
4. Perform the SSO login flow
5. Click **Export** in the upper right

## Where SAML is Found

The extract command looks for SAML in:

| Location | Parameter Names |
|:---------|:---------------|
| POST body | `SAMLRequest`, `SAMLResponse` |
| Query string | `SAMLRequest`, `SAMLResponse` |
| Response body | XML containing SAML elements |

## Common Workflows

### Debug an SSO flow

```bash
# 1. Capture HAR from browser
# 2. Preview contents
samlurai extract --list -f sso-flow.har

# 3. Extract to analyze
samlurai extract -f sso-flow.har -d ./debug

# 4. Inspect each file
samlurai inspect -f ./debug/saml_001_authnrequest.xml
samlurai inspect -f ./debug/saml_002_response.xml -k private.pem
```

### Compare multiple SSO flows

```bash
# Extract from different captures
samlurai extract -f working.har -d ./working
samlurai extract -f broken.har -d ./broken

# Diff the requests
diff ./working/saml_001_authnrequest.xml ./broken/saml_001_authnrequest.xml
```

### Archive SSO captures

```bash
# Extract and archive
samlurai extract -f capture.har -d ./archive/$(date +%Y%m%d)
```

## Error Handling

| Error | Cause | Solution |
|:------|:------|:---------|
| `no SAML data found` | HAR doesn't contain SAML | Check you captured the right flow |
| `failed to parse HAR file` | Invalid JSON | Ensure file is a valid HAR export |
| `failed to create output directory` | Permission denied | Check write permissions |

## See Also

- [`inspect`]({% link commands/inspect.md %}) - Inspect SAML details (also supports HAR files directly)
- [`decode`]({% link commands/decode.md %}) - Decode base64-encoded SAML
