"""
clawdpl: Create and manage OpenClaw AI Agent orchestrator instances.

This is a thin Python wrapper around the clawdpl Go binary.
For pipx users: `pipx run clawdpl --help`
"""

__version__ = "0.1.0"
__author__ = "Moltyverse"

from clawdpl.cli import main

__all__ = ["main", "__version__", "__author__"]
