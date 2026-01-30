---
layout: default
title: Home
nav_order: 1
description: "SAMLurai - A powerful CLI tool for decoding, decrypting, and debugging SAML assertions"
permalink: /
---

# SAMLurai 🥷
{: .fs-9 }

A powerful CLI tool for decoding, decrypting, and debugging SAML assertions.
{: .fs-6 .fw-300 }

[Get Started](#quick-start){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/gliwka/SAMLurai){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## Features

- **HAR file support** - Inspect or extract SAML from browser HAR captures
- **Inspect** SAML AuthnRequests, Responses, and Assertions
- **Decode** base64-encoded SAML responses and requests
- **Deflate** support for HTTP-Redirect binding
- **Decrypt** encrypted SAML assertions with private key
- **Smart auto-detection** - automatically decodes/decrypts as needed
- Multiple output formats: pretty, JSON, XML

## Quick Start

### Installation

```bash
# Homebrew (Recommended)
brew install gliwka/tap/samlurai

# Or from source
go install github.com/gliwka/SAMLurai@latest
```

### Inspect a HAR File

Captured a SAML flow in your browser? SAMLurai can inspect HAR files directly, showing all SAML messages in chronological order:

```bash
$ samlurai inspect -f sso-capture.har
```

**Output:**
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
  Issue Instant:  2024-01-15T10:30:00Z
  Destination:    https://idp.example.com/sso

▸ Request Details
  ACS URL:           https://sp.example.com/acs
  Protocol Binding:  HTTP-POST
  Force Authn:       false
  Is Passive:        false


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 [2/2] Response from request-body
       Parameter: SAMLResponse
       URL: https://sp.example.com/acs
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

═══════════════════════════════════════════════════════════════
 SAML Response
═══════════════════════════════════════════════════════════════

▸ Basic Information
  ID:              _response456
  Issuer:          https://idp.example.com
  Issue Instant:   2024-01-15T10:30:05Z
  Destination:     https://sp.example.com/acs
  In Response To:  _abc123

▸ Status
  Status Code:  Success

───────────────────────────────────────────────────────────────
 Embedded Assertion
───────────────────────────────────────────────────────────────

▸ Subject
  NameID:  user@example.com
  Format:  emailAddress

▸ Attributes
  Email (email):   user@example.com
  Groups (groups): admins, users
```

### Inspect a Single SAML File

Have a SAML Response or Assertion XML file? Inspect it directly:

```bash
$ samlurai inspect -f response.xml
```

**Output:**
```
═══════════════════════════════════════════════════════════════
 SAML Response
═══════════════════════════════════════════════════════════════

▸ Basic Information
  ID:              _response123
  Issuer:          https://idp.example.com
  Issue Instant:   2024-01-15T10:30:00Z
  Destination:     https://sp.example.com/acs

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

### Decrypt Encrypted Assertions

If your HAR file or SAML file contains encrypted assertions, provide a private key:

```bash
$ samlurai inspect -f capture.har -k private.pem
```

Without the key, SAMLurai still shows the Response metadata:

```
⚠️  Encrypted assertion detected - provide -k flag to decrypt

═══════════════════════════════════════════════════════════════
 SAML Response
═══════════════════════════════════════════════════════════════

▸ Basic Information
  ID:              _response123
  Issuer:          https://idp.example.com
  ...
```

### JSON Output for Scripting

Need to extract specific values? Use JSON output with `-o json`:

```bash
$ samlurai inspect -f response.xml -o json | jq '.assertion.subject.name_id'
"user@example.com"
```

### Extract SAML from HAR Files

Need to save each SAML message to a separate file?

```bash
# Preview what will be extracted
$ samlurai extract --list -f capture.har

# Extract to files
$ samlurai extract -f capture.har -d ./output
```

### Decode Base64 from Browser

When you copy a SAMLResponse from browser dev tools:

```bash
# From clipboard on macOS
$ pbpaste | samlurai inspect

# Or pipe directly
$ echo "PD94bWwgdmVyc2lvbj0i..." | samlurai inspect
```

## Command Overview

| Command | Description | HAR Support | Auto-decode | Auto-decrypt |
|:--------|:------------|:-----------:|:-----------:|:------------:|
| [`inspect`]({% link commands/inspect.md %}) | Parse and display SAML details | ✅ | ✅ | ✅ (with `-k`) |
| [`extract`]({% link commands/extract.md %}) | Extract SAML from HAR to files | ✅ | ✅ | ❌ |
| [`decode`]({% link commands/decode.md %}) | Decode base64-encoded SAML | ❌ | ❌ | ❌ |
| [`decrypt`]({% link commands/decrypt.md %}) | Decrypt encrypted assertions | ❌ | ✅ | ✅ |

## Output Formats

All commands support the `-o` or `--output` flag:

| Format | Description |
|:-------|:------------|
| `pretty` | Human-readable colored output (default) |
| `json` | JSON format for scripting |
| `xml` | Formatted XML output |

```bash
samlurai decode -o json "PHNhbWw..."
samlurai inspect -o xml -f assertion.xml
```

## CLI Reference

```
$ samlurai --help

SAMLurai is a command-line tool for decoding, decrypting, and inspecting 
SAML assertions. It helps developers and security professionals debug 
SAML-based authentication flows.

Examples:
  # Inspect a HAR file from browser DevTools
  samlurai inspect -f sso-capture.har

  # Inspect with decryption
  samlurai inspect -f capture.har -k private.pem

  # Extract SAML from HAR to files
  samlurai extract -f capture.har -d ./output

  # Decode a base64-encoded SAML response
  echo "PHNhbWw..." | samlurai decode

  # Inspect SAML assertion details
  samlurai inspect -f assertion.xml

Usage:
  samlurai [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  decode      Decode a base64-encoded SAML assertion
  decrypt     Decrypt an encrypted SAML assertion
  extract     Extract SAML assertions from HAR files
  help        Help about any command
  inspect     Inspect and display SAML assertion details

Flags:
  -h, --help            help for samlurai
  -o, --output string   Output format: pretty, json, xml (default "pretty")
  -v, --version         version for samlurai

Use "samlurai [command] --help" for more information about a command.
```

---

## About the Project

SAMLurai is a developer tool designed to simplify SAML debugging. Built with Go for cross-platform compatibility and fast performance.

### License

Distributed under the MIT License. See [LICENSE](https://github.com/gliwka/SAMLurai/blob/main/LICENSE) for more information.

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
