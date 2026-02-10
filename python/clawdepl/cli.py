#!/usr/bin/env python3
"""
CLI wrapper for clawdepl.

This module provides a thin wrapper around the clawdepl Go binary,
allowing it to be invoked via Python/pip/pipx installation.
"""

from __future__ import annotations

import gzip
import io
import os
import platform
import shutil
import stat
import subprocess
import sys
import tarfile
from pathlib import Path
from typing import Optional

# Version must match the Go binary version
__version__ = "0.1.0"

REPO = "clawdepl/clawdepl"


def get_platform_info() -> tuple[str, str, str]:
    """Get platform-specific binary info."""
    system = platform.system().lower()
    machine = platform.machine().lower()

    platform_map = {
        "darwin": "darwin",
        "linux": "linux",
        "windows": "windows",
    }

    arch_map = {
        "x86_64": "amd64",
        "amd64": "amd64",
        "arm64": "arm64",
        "aarch64": "arm64",
    }

    os_name = platform_map.get(system)
    arch = arch_map.get(machine)

    if not os_name or not arch:
        raise RuntimeError(f"Unsupported platform: {system}-{machine}")

    ext = ".exe" if system == "windows" else ""
    binary_name = f"clawdepl{ext}"

    return os_name, arch, binary_name


def get_binary_name() -> str:
    """Get the platform-specific binary name."""
    if platform.system() == "Windows":
        return "clawdepl.exe"
    return "clawdepl"


def get_local_binary_candidates() -> list[str]:
    """
    Return local packaged binary names to probe.

    Includes a legacy typo fallback (`clawdpl`) for backwards compatibility.
    """
    binary_name = get_binary_name()
    candidates = [binary_name]
    if binary_name == "clawdepl":
        candidates.append("clawdpl")
    return candidates


def get_cache_dir() -> Path:
    """Get the cache directory for storing the binary."""
    # Use standard cache locations
    if platform.system() == "Windows":
        base = Path(os.environ.get("LOCALAPPDATA", Path.home() / "AppData" / "Local"))
    elif platform.system() == "Darwin":
        base = Path.home() / "Library" / "Caches"
    else:
        base = Path(os.environ.get("XDG_CACHE_HOME", Path.home() / ".cache"))

    cache_dir = base / "clawdepl"
    cache_dir.mkdir(parents=True, exist_ok=True)
    return cache_dir


def download_binary(dest_path: Path) -> bool:
    """Download the binary from GitHub releases."""
    try:
        import httpx
    except ImportError:
        print("Warning: httpx not installed, cannot download binary", file=sys.stderr)
        return False

    try:
        os_name, arch, binary_name = get_platform_info()
    except RuntimeError as e:
        print(f"Warning: {e}", file=sys.stderr)
        return False

    archive_name = f"clawdepl_{__version__}_{os_name}_{arch}.tar.gz"
    url = f"https://github.com/{REPO}/releases/download/v{__version__}/{archive_name}"

    print(f"Downloading clawdepl v{__version__}...", file=sys.stderr)
    print(f"Platform: {os_name}-{arch}", file=sys.stderr)

    try:
        with httpx.Client(follow_redirects=True, timeout=60.0) as client:
            response = client.get(url)
            response.raise_for_status()

            # Extract binary from tar.gz
            with gzip.GzipFile(fileobj=io.BytesIO(response.content)) as gz:
                with tarfile.open(fileobj=gz, mode="r:") as tar:
                    for member in tar.getmembers():
                        if member.name == binary_name or member.name.endswith(f"/{binary_name}"):
                            extracted_file = tar.extractfile(member)
                            if extracted_file is None:
                                print(f"Archive entry {member.name} is not a regular file", file=sys.stderr)
                                return False

                            dest_path.parent.mkdir(parents=True, exist_ok=True)
                            with open(dest_path, "wb") as out:
                                shutil.copyfileobj(extracted_file, out)

                            # Make executable for user/group/other.
                            current_mode = dest_path.stat().st_mode
                            dest_path.chmod(current_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
                            print(f"Installed clawdepl to {dest_path}", file=sys.stderr)
                            return True

        print(f"Binary {binary_name} not found in archive", file=sys.stderr)
        return False

    except httpx.HTTPStatusError as e:
        if e.response.status_code == 404:
            print(f"Release v{__version__} not found (this is expected for unreleased versions)", file=sys.stderr)
        else:
            print(f"Download failed: HTTP {e.response.status_code}", file=sys.stderr)
        return False
    except Exception as e:
        print(f"Download failed: {e}", file=sys.stderr)
        return False


def find_binary() -> Optional[str]:
    """
    Find the clawdepl binary.

    Searches in the following order:
    1. CLAWDEPL_BINARY_PATH environment variable
    2. Same directory as this script (for development)
    3. Cache directory (downloaded binary)
    4. System PATH
    """
    binary_name = get_binary_name()
    local_candidates = get_local_binary_candidates()

    # 1. Check environment variable
    env_path = os.environ.get("CLAWDEPL_BINARY_PATH")
    if env_path and Path(env_path).exists():
        return env_path

    # 2. Check in the package directory (for development)
    package_dir = Path(__file__).parent
    for candidate in local_candidates:
        local_binary = package_dir / candidate
        if local_binary.exists():
            return str(local_binary)

    # Also check project root (for development)
    project_root = package_dir.parent.parent
    for candidate in local_candidates:
        dev_binary = project_root / candidate
        if dev_binary.exists():
            return str(dev_binary)

    # 3. Check cache directory
    cache_binary = get_cache_dir() / binary_name
    if cache_binary.exists():
        return str(cache_binary)

    # 4. Check system PATH
    path_binary = shutil.which(binary_name)
    if path_binary:
        return path_binary

    # 5. Try to download to cache
    if download_binary(cache_binary):
        return str(cache_binary)

    return None


def main() -> int:
    """
    Main entry point for the CLI wrapper.

    Finds and executes the clawdepl binary with all provided arguments.
    """
    binary_path = find_binary()

    if binary_path is None:
        print("Error: clawdepl binary not found.", file=sys.stderr)
        print("", file=sys.stderr)
        print("The binary could not be downloaded and is not in your PATH.", file=sys.stderr)
        print("", file=sys.stderr)
        print("Install the binary using one of these methods:", file=sys.stderr)
        print("", file=sys.stderr)
        print("  # Install from source (requires Go 1.21+):", file=sys.stderr)
        print("  go install github.com/clawdepl/clawdepl@latest", file=sys.stderr)
        print("", file=sys.stderr)
        print("  # Or set CLAWDEPL_BINARY_PATH:", file=sys.stderr)
        print("  export CLAWDEPL_BINARY_PATH=/path/to/clawdepl", file=sys.stderr)
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
