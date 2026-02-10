from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Optional


def _credentials_path() -> Path:
    return Path.home() / ".clawdepl" / "credentials.json"


def load_token_from_file() -> Optional[str]:
    """Read the access_token or api_key from ~/.clawdepl/credentials.json.

    Returns None if the file doesn't exist or contains no usable token.
    """
    path = _credentials_path()
    if not path.is_file():
        return None

    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return None

    # Prefer api_key (never expires), fall back to access_token
    api_key = data.get("api_key")
    if api_key:
        return str(api_key)

    access_token = data.get("access_token")
    if access_token:
        return str(access_token)

    return None


def resolve_token(explicit: Optional[str] = None) -> str:
    """Resolve a token from explicit arg > CLAWDEPL_TOKEN env > credentials file.

    Raises ValueError if no token can be found.
    """
    if explicit:
        return explicit

    env_token = os.environ.get("CLAWDEPL_TOKEN")
    if env_token:
        return env_token

    file_token = load_token_from_file()
    if file_token:
        return file_token

    raise ValueError(
        "No clawdepl token found. Provide a token explicitly, set "
        "CLAWDEPL_TOKEN, or run `clawdepl login`."
    )
