---
layout: default
title: Encrypted Assertions
parent: Guides
nav_order: 2
---

# Working with Encrypted Assertions
{: .no_toc }

How to handle encrypted SAML assertions with SAMLurai.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Why Encrypt SAML Assertions?

SAML assertions can contain sensitive user information:
- Email addresses
- Names
- Group memberships
- Custom attributes

Encryption protects this data from:
- Network eavesdropping
- Browser history exposure
- Log file leakage

{: .note }
Encryption is separate from signing. An assertion can be signed (for integrity), encrypted (for confidentiality), or both.

## How SAML Encryption Works

### Hybrid Encryption

SAML uses hybrid encryption (asymmetric + symmetric):

1. **Generate session key**: A random symmetric key (e.g., AES-256) is generated
2. **Encrypt data**: The assertion is encrypted with the session key
3. **Encrypt key**: The session key is encrypted with the SP's public key
4. **Package**: Both are bundled in an `EncryptedAssertion` element

### Decryption Process

To decrypt:
1. Use your private key to decrypt the session key
2. Use the session key to decrypt the assertion data

SAMLurai handles all of this automatically!

## Identifying Encrypted Assertions

### Using SAMLurai with HAR Files

When you inspect a HAR file containing encrypted assertions, SAMLurai shows a helpful message:

```bash
samlurai inspect -f capture.har
```

Output:
```
Found 2 SAML message(s) in HAR file:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 [1/2] AuthnRequest from request-body
       ...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 [2/2] Response from request-body
       ...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️  Encrypted assertion detected - provide -k flag to decrypt

═══════════════════════════════════════════════════════════════
 SAML Response
═══════════════════════════════════════════════════════════════

▸ Basic Information
  ID:              _response123
  Issuer:          https://idp.example.com
  ...

▸ Status
  Status Code:  Success

▸ Signature
  Signed:  Yes
  ...
```

{: .tip }
Even without the decryption key, SAMLurai shows Response metadata (status, issuer, signature). This is useful when receiving HAR files from users.

### Visual Inspection (Raw XML)

An encrypted SAML response looks like this:

```xml
<samlp:Response>
  <!-- Metadata... -->
  <saml:EncryptedAssertion>
    <xenc:EncryptedData>
      <xenc:CipherData>
        <xenc:CipherValue>Base64EncodedEncryptedData...</xenc:CipherValue>
      </xenc:CipherData>
    </xenc:EncryptedData>
  </saml:EncryptedAssertion>
</samlp:Response>
```

## Decrypting Assertions

### From HAR Files (Recommended)

The easiest way to work with encrypted assertions from users:

```bash
# Inspect HAR with decryption
samlurai inspect -f capture.har -k private.pem
```

This decrypts and displays all SAML messages in the HAR file.

### Extract and Decrypt from HAR

If you need to save the decrypted assertions:

```bash
# First, extract the raw SAML
samlurai extract -f capture.har -d ./extracted

# Then decrypt individual files
samlurai inspect -f ./extracted/saml_002_response.xml -k private.pem -o xml > decrypted.xml
```

### From Individual Files

```bash
samlurai inspect -f encrypted_response.xml -k private.pem
```

### From Base64-encoded Input

SAMLurai auto-decodes before decrypting:

```bash
echo "PHNhbWxwOlJlc3BvbnNlPi..." | samlurai inspect -k private.pem
```

### From Clipboard

```bash
pbpaste | samlurai inspect -k private.pem
```

### Just Decrypt (No Parse)

If you only want the decrypted XML:

```bash
samlurai decrypt -k private.pem -f encrypted.xml -o xml
```

## Support Workflow: Encrypted Assertions from Users

When users send you HAR files with encrypted assertions:

```bash
# 1. First, see what's in the HAR (works without key)
samlurai inspect -f user-capture.har

# 2. You'll see Response metadata and "Encrypted assertion detected" message

# 3. Decrypt with your SP's private key
samlurai inspect -f user-capture.har -k /path/to/sp-private.pem

# 4. Now you see the full assertion details including user attributes
```

{: .warning }
Keep your private key secure! Never ask users to send private keys. Users send you the HAR file, you decrypt with your own key.

## Managing Private Keys

### Key Format

SAMLurai accepts PEM-formatted private keys:

**RSA Private Key (PKCS#1)**:
```
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA...
-----END RSA PRIVATE KEY-----
```

**Private Key (PKCS#8)**:
```
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBg...
-----END PRIVATE KEY-----
```

### Converting Key Formats

**From PKCS#12 (PFX) to PEM**:
```bash
openssl pkcs12 -in cert.pfx -nocerts -out private.pem -nodes
```

**From DER to PEM**:
```bash
openssl rsa -in private.der -inform DER -out private.pem
```

**From Java KeyStore (JKS)**:
```bash
# First convert to PKCS#12
keytool -importkeystore -srckeystore keystore.jks \
  -destkeystore keystore.p12 -deststoretype PKCS12

# Then extract the key
openssl pkcs12 -in keystore.p12 -nocerts -out private.pem -nodes
```

### Key Security Best Practices

{: .warning }
Private keys are extremely sensitive. Follow these practices:

1. **Never commit keys to version control**
   ```bash
   # Add to .gitignore
   echo "*.pem" >> .gitignore
   echo "*.key" >> .gitignore
   ```

2. **Use restrictive permissions**
   ```bash
   chmod 600 private.pem
   ```

3. **Store in secure locations**
   - macOS: Keychain
   - Linux: `/etc/ssl/private/` or encrypted home directory
   - Cloud: Key management services (AWS KMS, Azure Key Vault)

4. **Rotate keys periodically**

5. **Use separate keys for dev/staging/production**

## Troubleshooting

### Wrong Private Key

```
Error: crypto/rsa: decryption error
```

**Cause**: The private key doesn't match the public key used for encryption.

**Solution**: 
1. Verify you're using the correct key
2. Check if the IdP has the right public key/certificate

### Unsupported Algorithm

```
Error: unsupported encryption algorithm
```

**Cause**: The IdP used an encryption algorithm SAMLurai doesn't support.

**Solution**: Check the algorithm in the raw XML:
```bash
samlurai decode -f response.txt -o xml | grep Algorithm
```

### Key Format Issues

```
Error: failed to load private key
```

**Cause**: Invalid PEM format or wrong key type.

**Solution**:
1. Verify PEM format:
   ```bash
   openssl rsa -in private.pem -check
   ```
2. Convert if necessary (see Key Format section)

### Encrypted Key (Password Protected)

```
Error: failed to load private key: encrypted key not supported
```

**Cause**: The private key is password-protected.

**Solution**: Remove the password:
```bash
openssl rsa -in encrypted.pem -out private.pem
```

## Testing Encryption Setup

### Generate Test Keys

```bash
# Generate a private key
openssl genrsa -out test-private.pem 2048

# Extract the public key
openssl rsa -in test-private.pem -pubout -out test-public.pem

# Generate a self-signed certificate
openssl req -new -x509 -key test-private.pem -out test-cert.pem -days 365 \
  -subj "/CN=Test SP"
```

### Verify Key Pair Matches

```bash
# These should output the same modulus
openssl rsa -in test-private.pem -modulus -noout
openssl x509 -in test-cert.pem -modulus -noout
```

## See Also

- [Debugging SSO]({% link guides/debugging-sso.md %})
- [`decrypt` command reference]({% link commands/decrypt.md %})
- [`inspect` command reference]({% link commands/inspect.md %})
