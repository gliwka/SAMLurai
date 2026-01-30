---
layout: default
title: decode
parent: Commands
nav_order: 1
---

# decode
{: .no_toc }

Decode a base64-encoded SAML assertion or response.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Synopsis

```
samlurai decode [base64-encoded-saml] [flags]
```

The `decode` command performs raw base64 decoding on SAML data. It does **not** automatically handle decryption—use [`inspect`]({% link commands/inspect.md %}) for that.

## Description

The decode command takes base64-encoded SAML data and outputs the decoded XML. This is useful when you need to see the raw XML content of a SAML response or request.

Input can be provided via:
- Command line argument
- File (`-f` flag)  
- Standard input (pipe)

For SAML requests using HTTP-Redirect binding, use the `--deflate` flag to decompress the deflated content after base64 decoding.

## Flags

| Flag | Short | Description | Default |
|:-----|:------|:------------|:--------|
| `--file` | `-f` | Read base64-encoded SAML from file | |
| `--deflate` | | Apply deflate decompression (for HTTP-Redirect binding) | `false` |
| `--output` | `-o` | Output format: `pretty`, `json`, `xml` | `pretty` |
| `--help` | `-h` | Help for decode | |

## Examples

### Decode from argument

```bash
samlurai decode "PHNhbWxwOlJlc3BvbnNlIHhtbG5zOnNhbWxwPSJ1cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoyLjA6cHJvdG9jb2wiPjwvc2FtbHA6UmVzcG9uc2U+"
```

Output:
```xml
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"></samlp:Response>
```

### Decode from file

```bash
samlurai decode -f response.txt
```

### Decode from stdin (pipe)

```bash
echo "PHNhbWxwOlJlc3BvbnNl..." | samlurai decode
```

Or using a file:

```bash
cat response.txt | samlurai decode
```

### Decode with deflate decompression

For HTTP-Redirect binding, SAML requests are typically deflate-compressed before base64 encoding:

```bash
samlurai decode --deflate -f request.txt
```

Or from stdin:

```bash
echo "nVLLTsMwELwj8Q+R79..." | samlurai decode --deflate
```

### Output as JSON

```bash
samlurai decode -o json "PHNhbWxwOlJlc3BvbnNl..."
```

### Output as formatted XML

```bash
samlurai decode -o xml -f response.txt
```

## Understanding Deflate

SAML uses two primary bindings for browser-based SSO:

| Binding | Encoding | Use Case |
|:--------|:---------|:---------|
| HTTP-POST | Base64 | Responses (typically) |
| HTTP-Redirect | Deflate + Base64 + URL-encode | Requests (typically) |

The `--deflate` flag handles the HTTP-Redirect binding format by applying DEFLATE decompression after base64 decoding.

{: .tip }
If you get garbled output or an error when decoding, try adding the `--deflate` flag. SAML requests in URL query parameters almost always use deflate compression.

## Error Handling

Common errors and solutions:

| Error | Cause | Solution |
|:------|:------|:---------|
| `illegal base64 data` | Invalid base64 encoding | Check for URL encoding, try URL-decoding first |
| `unexpected EOF` | Truncated input | Ensure complete base64 string |
| `flate: corrupt input` | Wrong deflate setting | Toggle `--deflate` flag |

## See Also

- [`decrypt`]({% link commands/decrypt.md %}) - Decrypt encrypted SAML assertions
- [`inspect`]({% link commands/inspect.md %}) - Inspect SAML details (auto-decodes)
