# Dark Storage CLI - Installation Guide

Multiple installation methods available for all platforms!

---

## Quick Install (Recommended)

### macOS / Linux

```bash
curl -fsSL https://install.darkstorage.io | sh
```

### Windows (PowerShell)

```powershell
irm https://install.darkstorage.io/windows.ps1 | iex
```

---

## Platform-Specific Installation

### macOS

#### Homebrew (Recommended)

```bash
brew install afterdarksys/tap/darkstorage
```

#### Go Install

```bash
go install github.com/darkstorage/cli@latest
```

### Linux

#### Universal Script

```bash
curl -fsSL https://install.darkstorage.io | sh
```

#### Snap

```bash
sudo snap install darkstorage
```

#### APT (Debian/Ubuntu)

```bash
# Add repository
curl -fsSL https://apt.darkstorage.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/darkstorage.gpg
echo "deb [signed-by=/usr/share/keyrings/darkstorage.gpg] https://apt.darkstorage.io/debian stable main" | sudo tee /etc/apt/sources.list.d/darkstorage.list

# Install
sudo apt update
sudo apt install darkstorage
```

#### YUM/DNF (RHEL/CentOS/Fedora)

```bash
# Add repository
sudo tee /etc/yum.repos.d/darkstorage.repo <<EOF
[darkstorage]
name=Dark Storage Repository
baseurl=https://yum.darkstorage.io/el/\$releasever/\$basearch
enabled=1
gpgcheck=1
gpgkey=https://yum.darkstorage.io/gpg
EOF

# Install
sudo dnf install darkstorage
# or
sudo yum install darkstorage
```

#### Go Install

```bash
go install github.com/darkstorage/cli@latest
```

### Windows

#### Scoop (Recommended)

```powershell
scoop bucket add afterdarksys https://github.com/afterdarksys/scoop-bucket
scoop install darkstorage
```

#### Chocolatey

```powershell
choco install darkstorage
```

#### Winget

```powershell
winget install AfterDarkSys.DarkStorage
```

#### Go Install

```powershell
go install github.com/darkstorage/cli@latest
```

---

## Manual Installation

### Download Binary

1. Go to [Releases](https://github.com/afterdarksys/darkstorage-cli/releases/latest)
2. Download the binary for your platform:
   - **macOS (Intel)**: `darkstorage_<version>_darwin_amd64.tar.gz`
   - **macOS (Apple Silicon)**: `darkstorage_<version>_darwin_arm64.tar.gz`
   - **Linux (x64)**: `darkstorage_<version>_linux_amd64.tar.gz`
   - **Linux (ARM64)**: `darkstorage_<version>_linux_arm64.tar.gz`
   - **Windows (x64)**: `darkstorage_<version>_windows_amd64.zip`

3. Extract the archive:
   ```bash
   # macOS/Linux
   tar -xzf darkstorage_*.tar.gz

   # Windows
   # Use Windows Explorer or:
   Expand-Archive darkstorage_*.zip
   ```

4. Move to a directory in your PATH:
   ```bash
   # macOS/Linux
   sudo mv darkstorage /usr/local/bin/

   # Windows (PowerShell as Admin)
   Move-Item darkstorage.exe C:\Windows\System32\
   ```

5. Verify installation:
   ```bash
   darkstorage version
   ```

---

## Docker

### Pull Image

```bash
docker pull darkstorage/cli:latest
```

### Run Commands

```bash
# Run a command
docker run --rm darkstorage/cli:latest ls

# Interactive shell
docker run --rm -it darkstorage/cli:latest sh

# Mount config
docker run --rm -v ~/.darkstorage:/home/darkstorage/.darkstorage darkstorage/cli:latest whoami
```

### Docker Compose

```yaml
version: '3.8'
services:
  darkstorage:
    image: darkstorage/cli:latest
    volumes:
      - ~/.darkstorage:/home/darkstorage/.darkstorage
      - ./data:/data
    command: sync
```

---

## Building from Source

### Prerequisites

- Go 1.24 or later
- Git
- GCC (for CGO dependencies)

### Clone and Build

```bash
# Clone repository
git clone https://github.com/afterdarksys/darkstorage-cli.git
cd darkstorage-cli

# Install dependencies
go mod download

# Build
go build -o darkstorage main.go

# Install globally
sudo mv darkstorage /usr/local/bin/

# Verify
darkstorage version
```

### Development Build

```bash
# Build with debug info
go build -gcflags="all=-N -l" -o darkstorage main.go

# Run tests
go test ./...

# Run with race detector
go run -race main.go
```

---

## Post-Installation

### 1. Verify Installation

```bash
darkstorage version
darkstorage --help
```

### 2. Log In

Choose one of the following methods:

#### OAuth/SSO (Recommended)

```bash
darkstorage login
```

This will open your browser for authentication.

#### API Key

```bash
darkstorage login --key YOUR_API_KEY
```

Get your API key from [console.darkstorage.io](https://console.darkstorage.io)

#### Config File

Download your config file from the web console and save to:
- **macOS/Linux**: `~/.darkstorage/config.yaml`
- **Windows**: `%USERPROFILE%\.darkstorage\config.yaml`

### 3. Test It Out

```bash
# Check authentication
darkstorage whoami

# List buckets
darkstorage ls

# Upload a file
echo "Hello Dark Storage" > test.txt
darkstorage put test.txt my-bucket/

# Download a file
darkstorage get my-bucket/test.txt ./downloaded.txt
```

---

## Configuration

### Config File Location

- **macOS/Linux**: `~/.darkstorage/config.yaml`
- **Windows**: `%USERPROFILE%\.darkstorage\config.yaml`

### Environment Variables

You can override config with environment variables:

```bash
export DARKSTORAGE_ENDPOINT="storage.darkstorage.io"
export DARKSTORAGE_ACCESS_KEY="your-access-key"
export DARKSTORAGE_SECRET_KEY="your-secret-key"
export DARKSTORAGE_USE_SSL="true"
```

### Example Config

```yaml
version: "1.0"

auth:
  api_key: "dk_live_abc123..."
  endpoint: "https://storage.darkstorage.io"

storage:
  endpoint: "storage.darkstorage.io"
  access_key: "AKIAIOSFODNN7EXAMPLE"
  secret_key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  region: "us-east-1"
  use_ssl: true

user:
  email: "user@example.com"
  account_id: "acc_abc123"
  plan: "professional"
```

---

## Shell Completion

### Bash

```bash
darkstorage completion bash > /etc/bash_completion.d/darkstorage
```

### Zsh

```bash
darkstorage completion zsh > "${fpath[1]}/_darkstorage"
```

### Fish

```bash
darkstorage completion fish > ~/.config/fish/completions/darkstorage.fish
```

### PowerShell

```powershell
darkstorage completion powershell | Out-String | Invoke-Expression
```

---

## Upgrading

### Homebrew

```bash
brew upgrade darkstorage
```

### Scoop

```powershell
scoop update darkstorage
```

### Chocolatey

```powershell
choco upgrade darkstorage
```

### Go Install

```bash
go install github.com/darkstorage/cli@latest
```

### Universal Script

```bash
curl -fsSL https://install.darkstorage.io | sh
```

---

## Uninstallation

### Homebrew

```bash
brew uninstall darkstorage
```

### Scoop

```powershell
scoop uninstall darkstorage
```

### Chocolatey

```powershell
choco uninstall darkstorage
```

### Manual

```bash
# Remove binary
sudo rm /usr/local/bin/darkstorage

# Remove config (optional)
rm -rf ~/.darkstorage
```

---

## Troubleshooting

### Command Not Found

**Issue**: `darkstorage: command not found`

**Solution**: Add install directory to PATH:

```bash
# macOS/Linux
export PATH="/usr/local/bin:$PATH"

# Or for Go install
export PATH="$HOME/go/bin:$PATH"
```

### Permission Denied

**Issue**: Permission errors when installing

**Solution**:
```bash
# Use sudo for system-wide install
sudo mv darkstorage /usr/local/bin/

# Or install to user directory
mv darkstorage ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
```

### SSL Certificate Errors

**Issue**: SSL verification errors

**Solution**:
```bash
# Update CA certificates
# macOS
brew install ca-certificates

# Ubuntu/Debian
sudo apt-get update && sudo apt-get install ca-certificates

# RHEL/CentOS
sudo yum install ca-certificates
```

### Connection Refused

**Issue**: Cannot connect to storage backend

**Solution**:
1. Check your internet connection
2. Verify endpoint in config: `darkstorage config get storage.endpoint`
3. Test connectivity: `ping storage.darkstorage.io`
4. Check firewall settings

---

## Getting Help

- **Documentation**: https://docs.darkstorage.io
- **GitHub Issues**: https://github.com/afterdarksys/darkstorage-cli/issues
- **Community Discord**: https://discord.gg/darkstorage
- **Email Support**: support@darkstorage.io

### Quick Commands

```bash
# Get help
darkstorage --help

# Get help for a command
darkstorage ls --help

# Check version
darkstorage version --verbose

# View config
darkstorage config list

# Check connectivity
darkstorage whoami
```

---

## Next Steps

Once installed, check out:

- [Quick Start Guide](QUICK_START.md) - Get started in 5 minutes
- [User Guide](docs/user-guide.md) - Complete feature documentation
- [CLI Reference](docs/cli-reference.md) - All available commands
- [API Documentation](docs/api.md) - For developers

---

*Last Updated: 2026-02-24*
*Version: 0.2*
