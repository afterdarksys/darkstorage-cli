# Dark Storage CLI

Secure, distributed cloud storage with S3 compatibility, Web3 integration, and disaster recovery built-in.

[![Release](https://img.shields.io/github/v/release/afterdarksys/darkstorage-cli)](https://github.com/afterdarksys/darkstorage-cli/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/afterdarksys/darkstorage-cli)](https://go.dev/)
[![License](https://img.shields.io/github/license/afterdarksys/darkstorage-cli)](LICENSE)
[![CI](https://github.com/afterdarksys/darkstorage-cli/workflows/CI/badge.svg)](https://github.com/afterdarksys/darkstorage-cli/actions)

---

## ✨ Features

- 🗄️ **S3-Compatible Storage** - Works with existing S3 tools and SDKs
- 📁 **Recursive Operations** - Upload/download entire directories
- 🔐 **Client-Side Encryption** - 3+1 key system with automatic rotation
- 🌐 **Web3 Integration** - Optional Storj and IPFS backends
- 🚨 **Instant DR** - Automatic disaster recovery for websites
- 📧 **Disaster Mail** - Email continuity during outages
- ⚡ **Progress Tracking** - Real-time upload/download progress
- 🎯 **Storage Classes** - AWS-compatible tiers (Standard, Glacier, etc.)
- 🔗 **Pre-signed URLs** - Secure temporary file sharing
- 🖥️ **Desktop GUI** - Native sync application
- 🔄 **Background Sync** - Automatic folder synchronization

---

## 🚀 Quick Install

### From Source (Simplest - Works Everywhere!)

```bash
# Clone the repo
git clone https://github.com/afterdarksys/darkstorage-cli.git
cd darkstorage-cli

# Install it!
./install.sh

# Or with Python
python3 install.py
```

**Options:**
```bash
./install.sh --fresh     # Clean build
./install.sh --update    # Update and rebuild
./install.sh --dev       # Debug build
python3 install.py --help
```

### Package Managers

**macOS/Linux:**
```bash
# Homebrew
brew install afterdarksys/tap/darkstorage

# Universal installer
curl -fsSL https://install.darkstorage.io | sh

# Go
go install github.com/darkstorage/cli@latest
```

**Windows:**
```powershell
# Scoop
scoop bucket add afterdarksys https://github.com/afterdarksys/scoop-bucket
scoop install darkstorage
```

**Docker:**
```bash
docker pull darkstorage/cli:latest
```

**[📖 Full Installation Guide](INSTALLATION.md)** - All platforms, package managers, and options

## Quick Start

1. Login to your Dark Storage account:
```bash
darkstorage login
```

2. List your buckets:
```bash
darkstorage ls
```

3. Upload a file:
```bash
darkstorage put ./file.txt my-bucket/
```

4. Download a file:
```bash
darkstorage get my-bucket/file.txt ./
```

## Available Commands

- `login` - Authenticate with Dark Storage
- `logout` - Remove stored credentials
- `whoami` - Display current user information
- `ls` - List buckets and files
- `put` - Upload files
- `get` - Download files
- `rm` - Delete files
- `cp` - Copy files
- `mv` - Move files
- `groups` - Manage user groups
- `perms` - Manage permissions
- `shares` - Manage file shares
- `scan` - Scan files for viruses/malware
- `trash` - Manage trash/deleted files
- `audit` - View audit logs
- `config` - Manage CLI configuration
- `version` - Display version information

## Configuration

The CLI stores configuration in `~/.darkstorage/config.yaml`.

### Environment Variables

- `DARKSTORAGE_API_KEY` - API key for authentication
- `DARKSTORAGE_ENDPOINT` - API endpoint (default: https://api.darkstorage.io)

### Command-line Flags

Global flags available for all commands:

- `--config` - Path to config file
- `--api-key` - API key (overrides config)
- `--endpoint` - API endpoint
- `-v, --verbose` - Verbose output
- `--json` - Output in JSON format

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Cleaning

```bash
make clean
```

## License

See LICENSE file for details.
