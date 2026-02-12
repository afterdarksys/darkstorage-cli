# Dark Storage CLI

API-first, developer-centric command-line interface for Dark Storage.

## Features

- S3-compatible cloud storage operations
- Block storage management
- User and group management
- Permissions and sharing
- Audit logging
- File scanning and trash management

## Installation

### From Source

```bash
go build -o darkstorage
```

Or use the Makefile:

```bash
make build
```

### Install to System

```bash
make install
```

This will install the binary to `/usr/local/bin/darkstorage`.

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
