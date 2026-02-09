"""
clawdepl: Create and manage OpenClaw AI Agent orchestrator instances.

This is a thin Python wrapper around the clawdepl Go binary.
For pipx users: `pipx run clawdepl --help`
"""

__version__ = "0.1.0"
__author__ = "clawdepl"

from clawdepl.cli import main

__all__ = ["main", "__version__", "__author__"]
