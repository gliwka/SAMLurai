---
layout: default
title: Installation
nav_order: 2
---

# Installation
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Requirements

- **Go 1.21+** (for building from source)
- No external runtime dependencies

## Homebrew (Recommended)

The easiest way to install SAMLurai on macOS or Linux is via Homebrew:

```bash
brew install gliwka/tap/samlurai
```

This will automatically install the latest version and keep it updated when you run `brew upgrade`.

## Install from Source

### Using go install

The simplest way to install SAMLurai is using `go install`:

```bash
go install github.com/gliwka/SAMLurai@latest
```

This will download, compile, and install the binary to your `$GOPATH/bin` directory.

{: .note }
Make sure `$GOPATH/bin` is in your `PATH` environment variable.

### Clone and Build

For development or to get the latest changes:

```bash
# Clone the repository
git clone https://github.com/gliwka/SAMLurai.git
cd samlurai

# Build the binary
make build

# The binary is created as ./samlurai
./samlurai --version
```

### Build Options

```bash
# Build for current platform
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Install to $GOPATH/bin
make install

# Clean build artifacts
make clean
```

## Pre-built Binaries

Pre-built binaries for various platforms are available on the [GitHub Releases](https://github.com/gliwka/SAMLurai/releases) page.

### macOS

```bash
# Download for macOS (Apple Silicon)
curl -LO https://github.com/gliwka/SAMLurai/releases/latest/download/samlurai-darwin-arm64.tar.gz
tar xzf samlurai-darwin-arm64.tar.gz
sudo mv samlurai /usr/local/bin/

# Download for macOS (Intel)
curl -LO https://github.com/gliwka/SAMLurai/releases/latest/download/samlurai-darwin-amd64.tar.gz
tar xzf samlurai-darwin-amd64.tar.gz
sudo mv samlurai /usr/local/bin/
```

### Linux

```bash
# Download for Linux (x64)
curl -LO https://github.com/gliwka/SAMLurai/releases/latest/download/samlurai-linux-amd64.tar.gz
tar xzf samlurai-linux-amd64.tar.gz
sudo mv samlurai /usr/local/bin/

# Download for Linux (ARM64)
curl -LO https://github.com/gliwka/SAMLurai/releases/latest/download/samlurai-linux-arm64.tar.gz
tar xzf samlurai-linux-arm64.tar.gz
sudo mv samlurai /usr/local/bin/
```

### Windows

Download the appropriate ZIP file from the [releases page](https://github.com/gliwka/SAMLurai/releases) and add the executable to your PATH.

## Verify Installation

After installation, verify SAMLurai is working:

```bash
samlurai --version
```

You should see output like:

```
samlurai version v1.0.0
```

## Shell Completion

SAMLurai supports shell completion for bash, zsh, fish, and PowerShell.

### Bash

```bash
# Add to ~/.bashrc
source <(samlurai completion bash)
```

### Zsh

```bash
# Add to ~/.zshrc
source <(samlurai completion zsh)
```

### Fish

```bash
samlurai completion fish | source
```

### PowerShell

```powershell
samlurai completion powershell | Out-String | Invoke-Expression
```

## Upgrading

To upgrade to the latest version:

```bash
# Using go install
go install github.com/gliwka/SAMLurai@latest

# Or if you cloned the repository
cd samlurai
git pull
make build
```

## Uninstalling

```bash
# If installed with go install
rm $(which samlurai)

# If installed manually
sudo rm /usr/local/bin/samlurai
```
