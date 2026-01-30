---
layout: default
title: Commands
nav_order: 3
has_children: true
---

# Commands
{: .no_toc }

SAMLurai provides four commands for working with SAML data.
{: .fs-6 .fw-300 }

---

## Command Overview

| Command | Description | HAR Support | Auto-decode | Auto-decrypt |
|:--------|:------------|:-----------:|:-----------:|:------------:|
| [`inspect`]({% link commands/inspect.md %}) | Parse and display SAML details | ✅ | ✅ | ✅ (with `-k`) |
| [`extract`]({% link commands/extract.md %}) | Extract SAML from HAR to files | ✅ | ✅ | ❌ |
| [`decode`]({% link commands/decode.md %}) | Decode base64-encoded SAML | ❌ | ❌ | ❌ |
| [`decrypt`]({% link commands/decrypt.md %}) | Decrypt encrypted assertions | ❌ | ✅ | ✅ |

## Choosing the Right Command

```
┌─────────────────────────────────────────────────────────────┐
│                  Which command should I use?                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ Have a HAR file │───Yes──▶ inspect -f file.har
                    │ from browser?   │         (or extract to files)
                    └─────────────────┘
                              │ No
                              ▼
                    ┌─────────────────┐
                    │ Need to see raw │───Yes──▶ decode
                    │ XML only?       │
                    └─────────────────┘
                              │ No
                              ▼
                    ┌─────────────────┐
                    │ Is the SAML     │───Yes──▶ inspect -k key.pem
                    │ encrypted?      │         (or decrypt -k key.pem)
                    └─────────────────┘
                              │ No / Don't know
                              ▼
                    ┌─────────────────┐
                    │ inspect         │
                    │ (handles all)   │
                    └─────────────────┘
```

{: .tip }
When in doubt, use `inspect`. It automatically handles decoding and can decrypt when you provide a key.

## Global Flags

These flags are available for all commands:

| Flag | Short | Description | Default |
|:-----|:------|:------------|:--------|
| `--output` | `-o` | Output format: `pretty`, `json`, `xml` | `pretty` |
| `--help` | `-h` | Display help for the command | |
| `--version` | `-v` | Display version information | |

## Input Methods

All commands support multiple input methods:

### From File

Use the `-f` or `--file` flag:

```bash
samlurai inspect -f response.xml
```

### From Stdin (Pipe)

Pipe data directly:

```bash
cat response.xml | samlurai inspect
echo "PHNhbWw..." | samlurai decode
```

### From Argument (decode only)

```bash
samlurai decode "PHNhbWxwOlJlc3BvbnNl..."
```

## Output Formats

### Pretty (default)

Human-readable, colored output with formatting:

```bash
samlurai inspect -f response.xml
# or
samlurai inspect -f response.xml -o pretty
```

### JSON

Structured JSON for scripting and automation:

```bash
samlurai inspect -f response.xml -o json

# Combine with jq
samlurai inspect -f response.xml -o json | jq '.assertion.attributes'
```

### XML

Formatted, indented XML:

```bash
samlurai decode -o xml "PHNhbWw..."
```
