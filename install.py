#!/usr/bin/env python3
"""
Dark Storage CLI - Installation Script (Python)
Usage: python3 install.py [OPTIONS]
"""

import os
import sys
import subprocess
import shutil
import argparse
from pathlib import Path
from datetime import datetime

# Colors
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

def info(msg):
    print(f"{Colors.BLUE}ℹ{Colors.NC} {msg}")

def success(msg):
    print(f"{Colors.GREEN}✓{Colors.NC} {msg}")

def error(msg):
    print(f"{Colors.RED}✗{Colors.NC} {msg}")

def warn(msg):
    print(f"{Colors.YELLOW}⚠{Colors.NC} {msg}")

def run_command(cmd, cwd=None, capture=False):
    """Run a shell command"""
    try:
        if capture:
            result = subprocess.run(cmd, shell=True, cwd=cwd,
                                   capture_output=True, text=True, check=True)
            return result.stdout.strip()
        else:
            subprocess.run(cmd, shell=True, cwd=cwd, check=True)
            return True
    except subprocess.CalledProcessError as e:
        return None

def check_requirements():
    """Check if required tools are installed"""
    info("Checking requirements...")

    # Check Go
    if not shutil.which('go'):
        error("Go is not installed")
        print("Install Go from: https://go.dev/dl/")
        sys.exit(1)

    go_version = run_command("go version", capture=True)
    success(f"Go installed: {go_version}")

    # Check Git
    if not shutil.which('git'):
        error("Git is not installed")
        sys.exit(1)

    success("Git installed")

def update_repo(args):
    """Update repository to latest version"""
    if args.update:
        info("Updating repository...")

        # Check for uncommitted changes
        status = run_command("git status --porcelain", capture=True)
        if status:
            warn("You have uncommitted changes. Stashing...")
            run_command("git stash")

        # Pull latest
        if run_command("git pull origin main"):
            success("Repository updated to latest version")
        else:
            error("Failed to update repository")
            sys.exit(1)

def clean_build(args):
    """Clean previous builds"""
    if args.fresh:
        info("Cleaning previous builds...")

        binary_name = "darkstorage"
        if sys.platform == "win32":
            binary_name = "darkstorage.exe"

        # Remove binary
        if os.path.exists(binary_name):
            os.remove(binary_name)
            success("Removed old binary")

        # Clean go cache
        run_command("go clean -cache -modcache -testcache")
        success("Cleaned Go caches")

def download_deps():
    """Download Go dependencies"""
    info("Downloading dependencies...")
    if run_command("go mod download"):
        success("Dependencies downloaded")
    else:
        error("Failed to download dependencies")
        sys.exit(1)

def build_binary(args):
    """Build the binary"""
    info("Building Dark Storage CLI...")

    binary_name = "darkstorage"
    if sys.platform == "win32":
        binary_name = "darkstorage.exe"

    # Build flags
    if args.dev:
        warn("Building with debug symbols (larger binary, slower execution)")
        build_flags = "-gcflags=all=-N -l"
    else:
        build_flags = "-ldflags=-s -w"

    # Get version info
    version = run_command("git describe --tags --always --dirty", capture=True) or "dev"
    commit = run_command("git rev-parse --short HEAD", capture=True) or "unknown"
    build_date = datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")

    # Build command
    ldflags = (
        f"-X github.com/darkstorage/cli/cmd.Version={version} "
        f"-X github.com/darkstorage/cli/cmd.Commit={commit} "
        f"-X github.com/darkstorage/cli/cmd.Date={build_date} "
        f"-X github.com/darkstorage/cli/cmd.BuiltBy=local"
    )

    if args.dev:
        cmd = f"go build {build_flags} -o {binary_name} main.go"
    else:
        cmd = f'go build -ldflags="{ldflags}" -o {binary_name} main.go'

    if run_command(cmd):
        # Get file size
        size = os.path.getsize(binary_name)
        size_mb = size / (1024 * 1024)
        success(f"Build complete ({size_mb:.1f} MB)")
    else:
        error("Build failed")
        sys.exit(1)

def test_binary():
    """Test the built binary"""
    info("Testing binary...")

    binary_name = "./darkstorage"
    if sys.platform == "win32":
        binary_name = "darkstorage.exe"

    # Make executable on Unix
    if sys.platform != "win32":
        os.chmod("darkstorage", 0o755)

    # Test version command
    result = run_command(f"{binary_name} version", capture=True)
    if result:
        success("Binary works correctly")
        print()
        # Show verbose version
        version_output = run_command(f"{binary_name} version --verbose", capture=True)
        print(version_output)
        print()
    else:
        error("Binary test failed")
        sys.exit(1)

def install_binary(args):
    """Install binary to system"""
    install_dir = args.install_dir or os.environ.get('INSTALL_DIR', '/usr/local/bin')

    info(f"Installing to {install_dir}...")

    binary_name = "darkstorage"
    if sys.platform == "win32":
        binary_name = "darkstorage.exe"

    # Create install directory if needed
    Path(install_dir).mkdir(parents=True, exist_ok=True)

    # Check if we can write to install_dir
    install_path = os.path.join(install_dir, binary_name)

    try:
        shutil.copy2(binary_name, install_path)
        success(f"Installed to {install_path}")
    except PermissionError:
        # Try with sudo
        if shutil.which('sudo'):
            info("Need sudo permissions...")
            if run_command(f"sudo cp {binary_name} {install_path}"):
                success("Installed with sudo")
            else:
                error("Failed to install with sudo")
                sys.exit(1)
        else:
            error(f"{install_dir} is not writable and sudo is not available")
            warn(f"Try: INSTALL_DIR=~/.local/bin python3 {sys.argv[0]}")
            sys.exit(1)

def verify_installation():
    """Verify the installation"""
    info("Verifying installation...")

    binary_name = "darkstorage"

    if shutil.which(binary_name):
        success("Installation verified!")
        version = run_command(f"{binary_name} version", capture=True)
        print(f"  {version}")
    else:
        warn(f"{binary_name} is installed but not in PATH")
        print()
        warn("Add the install directory to your PATH")

def print_next_steps(args):
    """Print next steps"""
    print()
    print(f"{Colors.GREEN}{'━' * 60}{Colors.NC}")
    print(f"{Colors.GREEN}  Dark Storage CLI installed successfully! 🚀{Colors.NC}")
    print(f"{Colors.GREEN}{'━' * 60}{Colors.NC}")
    print()
    print("Next steps:")
    print()
    print("  1. Log in to your Dark Storage account:")
    print(f"     {Colors.BLUE}darkstorage login{Colors.NC}")
    print()
    print("  2. Or use an API key:")
    print(f"     {Colors.BLUE}darkstorage login --key YOUR_API_KEY{Colors.NC}")
    print()
    print("  3. Test it out:")
    print(f"     {Colors.BLUE}darkstorage whoami{Colors.NC}")
    print(f"     {Colors.BLUE}darkstorage ls{Colors.NC}")
    print()
    print("  4. Get help:")
    print(f"     {Colors.BLUE}darkstorage --help{Colors.NC}")
    print()

    if args.dev:
        warn("This is a DEBUG build (not optimized for production)")
        print("  For production use: python3 install.py (without --dev)")
        print()

def main():
    """Main installation function"""
    parser = argparse.ArgumentParser(
        description='Dark Storage CLI - Installation Script',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python3 install.py                    # Normal install
  python3 install.py --fresh            # Clean install
  python3 install.py --update           # Update and reinstall
  python3 install.py --dev              # Debug build
  INSTALL_DIR=~/.local/bin python3 install.py   # Install to user directory
        """
    )

    parser.add_argument('--fresh', action='store_true',
                       help='Clean build from scratch (removes binaries and caches)')
    parser.add_argument('--update', action='store_true',
                       help='Update to latest git version and rebuild')
    parser.add_argument('--dev', action='store_true',
                       help='Build with debug symbols (no optimization)')
    parser.add_argument('--install-dir', type=str,
                       help='Installation directory (default: /usr/local/bin)')

    args = parser.parse_args()

    # Print header
    print()
    print(f"{Colors.BLUE}{'━' * 60}{Colors.NC}")
    print(f"{Colors.BLUE}  Dark Storage CLI - Local Installation{Colors.NC}")
    print(f"{Colors.BLUE}{'━' * 60}{Colors.NC}")
    print()

    # Run installation steps
    check_requirements()
    update_repo(args)
    clean_build(args)
    download_deps()
    build_binary(args)
    test_binary()
    install_binary(args)
    verify_installation()
    print_next_steps(args)

if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print()
        error("Installation cancelled by user")
        sys.exit(1)
    except Exception as e:
        print()
        error(f"Installation failed: {e}")
        sys.exit(1)
