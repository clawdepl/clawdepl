#!/usr/bin/env python3
"""
CLI wrapper for clawdpl.

This module provides a thin wrapper around the clawdpl Go binary,
allowing it to be invoked via Python/pip installation.
"""

import os
import platform
import shutil
import subprocess
import sys
from pathlib import Path


def get_binary_name() -> str:
    """Get the platform-specific binary name."""
    if platform.system() == "Windows":
        return "clawdpl.exe"
    return "clawdpl"


def find_binary() -> str | None:
    """
    Find the clawdpl binary.
    
    Searches in the following order:
    1. Same directory as this script
    2. Package data directory
    3. System PATH
    """
    binary_name = get_binary_name()
    
    # Check in the package directory
    package_dir = Path(__file__).parent
    local_binary = package_dir / binary_name
    if local_binary.exists():
        return str(local_binary)
    
    # Check in a bin subdirectory
    bin_binary = package_dir / "bin" / binary_name
    if bin_binary.exists():
        return str(bin_binary)
    
    # Check system PATH
    path_binary = shutil.which(binary_name)
    if path_binary:
        return path_binary
    
    return None


def main() -> int:
    """
    Main entry point for the CLI wrapper.
    
    Finds and executes the clawdpl binary with all provided arguments.
    """
    binary_path = find_binary()
    
    if binary_path is None:
        print("Error: clawdpl binary not found.", file=sys.stderr)
        print("", file=sys.stderr)
        print("Please install the binary using one of these methods:", file=sys.stderr)
        print("", file=sys.stderr)
        print("  # Install from source (requires Go):", file=sys.stderr)
        print("  go install github.com/moltyverse/clawdpl@latest", file=sys.stderr)
        print("", file=sys.stderr)
        print("  # Or build locally:", file=sys.stderr)
        print("  git clone https://github.com/moltyverse/clawdpl.git", file=sys.stderr)
        print("  cd clawdpl", file=sys.stderr)
        print("  go build -o clawdpl .", file=sys.stderr)
        print("", file=sys.stderr)
        return 1
    
    # Pass all arguments to the binary
    args = [binary_path] + sys.argv[1:]
    
    try:
        result = subprocess.run(args, check=False)
        return result.returncode
    except FileNotFoundError:
        print(f"Error: Could not execute {binary_path}", file=sys.stderr)
        return 1
    except KeyboardInterrupt:
        return 130  # Standard exit code for SIGINT


if __name__ == "__main__":
    sys.exit(main())
