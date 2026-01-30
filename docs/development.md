---
layout: default
title: Development
nav_order: 5
---

# Development
{: .no_toc }

Guide for contributing to SAMLurai.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Prerequisites

- **Go 1.21+**
- **Make** (optional, but recommended)
- **Git**

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/gliwka/SAMLurai.git
cd samlurai
```

### Build

```bash
# Using Make
make build

# Or directly with Go
go build -o samlurai .
```

### Run Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Verbose output
go test -v ./...
```

## Project Structure

```
SAMLurai/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── root_test.go
│   ├── decode.go          # Decode command
│   ├── decode_test.go
│   ├── decrypt.go         # Decrypt command
│   ├── decrypt_test.go
│   ├── inspect.go         # Inspect command
│   └── inspect_test.go
├── internal/
│   ├── saml/              # SAML processing logic
│   │   ├── decoder.go     # Base64/deflate decoding
│   │   ├── decoder_test.go
│   │   ├── decryptor.go   # Encryption handling
│   │   ├── decryptor_test.go
│   │   ├── parser.go      # XML parsing
│   │   ├── parser_test.go
│   │   └── types.go       # Data structures
│   ├── output/            # Output formatting
│   │   ├── formatter.go
│   │   └── formatter_test.go
│   └── testutil/          # Test utilities
│       ├── golden.go      # Golden file testing
│       └── golden_test.go
├── testdata/
│   ├── fixtures/          # Test input data
│   │   └── assertions/
│   └── golden/            # Expected test outputs
│       ├── decode/
│       └── inspect/
├── docs/                  # Documentation (GitHub Pages)
├── main.go               # Entry point
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Architecture

### Command Layer (`cmd/`)

Built with [Cobra](https://github.com/spf13/cobra):

- `root.go`: Base command, global flags
- `decode.go`: Base64 decoding
- `decrypt.go`: Decryption with auto-decode
- `inspect.go`: Full pipeline (decode → decrypt → parse)

### SAML Processing (`internal/saml/`)

Core SAML logic:

- **Decoder**: Base64 and deflate handling
  - `Decode()`: Standard base64
  - `DecodeDeflate()`: Deflate + base64
  - `SmartDecode()`: Auto-detect encoding

- **Decryptor**: XML encryption handling
  - Uses [crewjam/saml](https://github.com/crewjam/saml) library
  - RSA key transport
  - AES data encryption

- **Parser**: XML parsing to structured data
  - Uses [beevik/etree](https://github.com/beevik/etree)
  - Extracts all SAML fields
  - Returns `SAMLInfo` struct

### Output Formatting (`internal/output/`)

Multiple output formats:

- `pretty`: Colored, human-readable
- `json`: Structured JSON
- `xml`: Formatted XML

## Testing Strategy

### Unit Tests

Each package has comprehensive unit tests:

```bash
go test ./internal/saml/...
go test ./internal/output/...
go test ./cmd/...
```

### Golden File Tests

Expected outputs stored as golden files:

```bash
# Run golden tests
go test ./... -run Golden

# Update golden files
make update-golden
# Or
go test ./... -update
```

Golden files location: `testdata/golden/`

### Test Fixtures

Test input data in `testdata/fixtures/`:

```
testdata/
├── fixtures/
│   └── assertions/
│       ├── assertion.xml
│       ├── request.xml
│       └── response.xml
└── golden/
    ├── decode/
    │   └── response_decoded.golden
    └── inspect/
        ├── assertion_inspect.golden
        └── response_inspect.golden
```

### Running Specific Tests

```bash
# Run tests matching a pattern
go test ./... -run TestDecode

# Run tests in a specific package
go test ./internal/saml/...

# Run with race detection
go test -race ./...
```

## Code Style

### Formatting

```bash
# Format code
go fmt ./...

# Or use goimports
goimports -w .
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Guidelines

1. **Error handling**: Always wrap errors with context
   ```go
   if err != nil {
       return fmt.Errorf("failed to parse SAML: %w", err)
   }
   ```

2. **Documentation**: Public functions need comments
   ```go
   // Decode decodes a base64-encoded SAML string.
   // It handles both standard and URL-safe base64.
   func (d *Decoder) Decode(input string) ([]byte, error) {
   ```

3. **Testing**: Write tests for new functionality

## Adding a New Command

1. Create `cmd/newcommand.go`:

```go
package cmd

import "github.com/spf13/cobra"

var newCmd = &cobra.Command{
    Use:   "newcommand",
    Short: "Brief description",
    Long:  `Detailed description...`,
    RunE:  runNewCommand,
}

func init() {
    rootCmd.AddCommand(newCmd)
    // Add flags
}

func runNewCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

2. Add tests in `cmd/newcommand_test.go`

3. Update documentation

## Dependencies

Key dependencies:

| Package | Purpose |
|:--------|:--------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [beevik/etree](https://github.com/beevik/etree) | XML processing |
| [crewjam/saml](https://github.com/crewjam/saml) | SAML decryption |
| [fatih/color](https://github.com/fatih/color) | Terminal colors |
| [stretchr/testify](https://github.com/stretchr/testify) | Testing utilities |

### Updating Dependencies

```bash
go get -u ./...
go mod tidy
```

## Making a Release

1. Update version in code
2. Create git tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. GitHub Actions builds and publishes binaries

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write/update tests
5. Run `make test` and `make lint`
6. Submit a pull request

### Commit Messages

Follow conventional commits:

```
feat: add new feature
fix: fix bug in decoder
docs: update README
test: add tests for parser
refactor: simplify decrypt logic
```
