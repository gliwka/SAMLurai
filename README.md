# SAMLurai 🥷

A powerful CLI tool for decoding, decrypting, and debugging SAML assertions.


## Features

- **HAR file support** - extract and analyze SAML messages from browser network captures
- **Decode** base64-encoded SAML responses and requests
- **Deflate** support for HTTP-Redirect binding
- **Decrypt** encrypted SAML assertions with private key
- **Inspect** SAML assertions with human-readable output
- **Smart auto-detection** - automatically decodes/decrypts as needed
- Multiple output formats: pretty, JSON, XML

## Installation

### Homebrew (Recommended)

```bash
brew install gliwka/tap/samlurai
```

### From Source

```bash
# Using go install
go install github.com/gliwka/SAMLurai@latest

# Or clone and build
git clone https://github.com/gliwka/SAMLurai.git
cd samlurai
make build
```

## Quick Start

### From HAR Files (Recommended)

The easiest way to debug SAML SSO issues is to capture a HAR file from your browser, then analyze it:

```bash
# Inspect all SAML messages in a HAR file
samlurai inspect -f sso-session.har

# With encrypted assertions, provide a private key
samlurai inspect -f sso-session.har -k private.pem

# Extract SAML messages to files for further analysis
samlurai extract -f session.har -o ./saml-messages/
```

**Capture a HAR file:** Open DevTools (F12) → Network tab → reproduce the SSO flow → click "Export HAR" (Chrome/Edge) or right-click → "Save All As HAR" (Firefox).

### From Individual Files

```bash
# Inspect from XML file
samlurai inspect -f response.xml

# Inspect base64-encoded SAML (auto-decodes)
echo "PHNhbWxwOlJlc3BvbnNl..." | samlurai inspect

# Inspect encrypted SAML (auto-decodes + auto-decrypts)
samlurai inspect -f encrypted_response.xml -k private.pem
```

Without the `-k` flag, SAMLurai still shows Response metadata but indicates encrypted assertions need a key to decrypt.

## Usage

### Decode a SAML Response

The `decode` command only performs base64 decoding:

```bash
# From argument
samlurai decode "PHNhbWxwOlJlc3BvbnNl..."

# From file
samlurai decode -f response.txt

# From stdin (pipe)
echo "PHNhbWxwOlJlc3BvbnNl..." | samlurai decode

# With deflate decompression (HTTP-Redirect binding)
samlurai decode --deflate -f request.txt
```

### Decrypt an Encrypted Assertion

The `decrypt` command auto-decodes if needed, then decrypts:

```bash
# Decrypt with private key (auto-decodes base64 if needed)
samlurai decrypt -k private.pem -f encrypted_assertion.xml

# Decrypt base64-encoded encrypted SAML
echo "PHNhbWw6RW5jcnlwdGVkQXNzZXJ0aW9u..." | samlurai decrypt -k private.pem
```

### Inspect SAML Details

The `inspect` command does it all - auto-decodes and auto-decrypts as needed:

```bash
# Inspect from file (raw XML)
samlurai inspect -f assertion.xml

# Inspect base64-encoded SAML (auto-decodes)
echo "PHNhbWw..." | samlurai inspect

# Inspect encrypted SAML (auto-decrypts with key)
samlurai inspect -f encrypted.xml -k private.pem

# Full pipeline: base64 → decode → decrypt → inspect
echo "BASE64_ENCRYPTED_SAML" | samlurai inspect -k private.pem

# Output as JSON
samlurai inspect -f assertion.xml -o json
```

### Extract from HAR Files

The `extract` command saves SAML messages from HAR files:

```bash
# List SAML messages in a HAR file
samlurai extract -f session.har --list

# Extract all SAML messages to files
samlurai extract -f session.har -o ./saml-messages/

# Extract specific message by index
samlurai extract -f session.har -o ./saml-messages/ --index 0
```

### Output Formats

All commands support the `-o` or `--output` flag:

- `pretty` (default) - Human-readable colored output
- `json` - JSON format
- `xml` - Formatted XML

```bash
samlurai decode -o json "PHNhbWw..."
samlurai inspect -o xml -f assertion.xml
```

## Command Behavior

| Command   | HAR Support | Auto-decode | Auto-decrypt | Notes |
|-----------|-------------|-------------|--------------|-------|
| `inspect` | ✅ | ✅ | ✅ (with `-k`) | Full pipeline, displays parsed info |
| `extract` | ✅ | ❌ | ❌ | Extract SAML messages from HAR files |
| `decode`  | ❌ | ❌ | ❌ | Raw base64 decode only |
| `decrypt` | ❌ | ✅ | ✅ | Decodes if needed, then decrypts |

## Development

### Prerequisites

- Go 1.21+

### Build

```bash
make build
```

### Test

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Update golden files
make update-golden
```

### Project Structure

```
SAMLurai/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go
│   ├── decode.go
│   ├── decrypt.go
│   └── inspect.go
├── internal/
│   ├── saml/              # SAML processing logic
│   │   ├── decoder.go     # Base64/deflate decoding
│   │   ├── decryptor.go   # Encryption handling
│   │   ├── parser.go      # XML parsing
│   │   └── types.go       # Data structures
│   ├── output/            # Output formatting
│   │   └── formatter.go
│   └── testutil/          # Test utilities
│       └── golden.go      # Golden file testing
├── testdata/
│   ├── fixtures/          # Test input data
│   └── golden/            # Expected test outputs
├── main.go
├── go.mod
├── Makefile
└── README.md
```

## Testing Strategy

### Unit Tests
Each internal package has comprehensive unit tests using [testify](https://github.com/stretchr/testify).

### Golden File Tests
CLI output is tested against golden files for consistency. Update golden files with:

```bash
make update-golden
```

### Integration Tests
End-to-end tests exercise full CLI workflows including file I/O and piping.

## License

MIT
