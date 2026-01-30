---
layout: default
title: decrypt
parent: Commands
nav_order: 2
---

# decrypt
{: .no_toc }

Decrypt an encrypted SAML assertion using a private key.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Synopsis

```
samlurai decrypt [flags]
```

The `decrypt` command handles encrypted SAML assertions. It automatically decodes base64 input before decryption.

## Description

The decrypt command decrypts encrypted SAML assertions using a PEM-formatted private key. It supports:

- **Auto-decode**: Automatically detects and decodes base64-encoded input
- **Smart detection**: Handles both raw XML and encoded formats
- Multiple encryption algorithms (AES-128, AES-256, etc.)

Input can be provided via:
- File (`-f` flag)
- Standard input (pipe)

## Flags

| Flag | Short | Description | Required |
|:-----|:------|:------------|:---------|
| `--key` | `-k` | Path to private key (PEM format) | ✅ Yes |
| `--file` | `-f` | Read encrypted SAML from file | |
| `--output` | `-o` | Output format: `pretty`, `json`, `xml` | |
| `--help` | `-h` | Help for decrypt | |

## Examples

### Decrypt from file

```bash
samlurai decrypt -k private.pem -f encrypted_assertion.xml
```

### Decrypt from stdin

```bash
cat encrypted.xml | samlurai decrypt -k private.pem
```

### Decrypt base64-encoded input

The command auto-detects and decodes base64:

```bash
echo "PHNhbWw6RW5jcnlwdGVkQXNzZXJ0aW9uPi4uLg==" | samlurai decrypt -k private.pem
```

### Output as JSON

```bash
samlurai decrypt -k private.pem -f encrypted.xml -o json
```

### Full pipeline from browser

When copying a SAML response from browser dev tools:

```bash
# The response is typically base64-encoded
pbpaste | samlurai decrypt -k private.pem
```

## Private Key Format

The private key must be in PEM format:

```
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA...
...
-----END RSA PRIVATE KEY-----
```

Or PKCS#8 format:

```
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBg...
...
-----END PRIVATE KEY-----
```

{: .warning }
Keep your private keys secure! Never commit them to version control or share them in logs.

## Supported Encryption

SAMLurai supports the following XML encryption algorithms:

### Key Transport

| Algorithm | OID |
|:----------|:----|
| RSA-OAEP | `http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p` |
| RSA-OAEP (SHA-256) | `http://www.w3.org/2009/xmlenc11#rsa-oaep` |
| RSA-1_5 | `http://www.w3.org/2001/04/xmlenc#rsa-1_5` |

### Data Encryption

| Algorithm | Key Size |
|:----------|:---------|
| AES128-CBC | 128-bit |
| AES192-CBC | 192-bit |
| AES256-CBC | 256-bit |
| AES128-GCM | 128-bit |
| AES256-GCM | 256-bit |
| TripleDES-CBC | 168-bit |

## Understanding Encrypted SAML

An encrypted SAML assertion looks like this:

```xml
<saml:EncryptedAssertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
  <xenc:EncryptedData xmlns:xenc="http://www.w3.org/2001/04/xmlenc#">
    <xenc:EncryptionMethod Algorithm="http://www.w3.org/2001/04/xmlenc#aes256-cbc"/>
    <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
      <xenc:EncryptedKey>
        <xenc:EncryptionMethod Algorithm="http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p"/>
        <xenc:CipherData>
          <xenc:CipherValue>...</xenc:CipherValue>
        </xenc:CipherData>
      </xenc:EncryptedKey>
    </ds:KeyInfo>
    <xenc:CipherData>
      <xenc:CipherValue>...</xenc:CipherValue>
    </xenc:CipherData>
  </xenc:EncryptedData>
</saml:EncryptedAssertion>
```

The decryption process:
1. Extract the encrypted session key from `EncryptedKey`
2. Decrypt the session key using your RSA private key
3. Use the session key to decrypt the assertion data
4. Return the decrypted XML

## Error Handling

Common errors and solutions:

| Error | Cause | Solution |
|:------|:------|:---------|
| `failed to load private key` | Invalid key file | Check PEM format and file path |
| `crypto/rsa: decryption error` | Wrong private key | Use the matching private key |
| `no encrypted assertion found` | Not encrypted | Use `decode` or `inspect` instead |
| `unsupported encryption algorithm` | Uncommon algorithm | Check IdP configuration |

## See Also

- [`decode`]({% link commands/decode.md %}) - Decode base64-encoded SAML
- [`inspect`]({% link commands/inspect.md %}) - Inspect SAML details (auto-decrypts)
