from __future__ import annotations

import json
import re
from typing import Any

DEFAULT_MODEL = "anthropic/claude-opus-4-6"

_MODEL_REF_RE = re.compile(r"^[a-z0-9][a-z0-9.\-]*/[A-Za-z0-9._:/-]+$")

NATIVE_PROVIDERS: frozenset[str] = frozenset(
    {
        "anthropic",
        "openai",
        "openai-codex",
        "opencode",
        "google",
        "google-vertex",
        "google-antigravity",
        "google-gemini-cli",
        "zai",
        "vercel-ai-gateway",
        "openrouter",
        "amazon-bedrock",
        "xai",
        "groq",
        "cerebras",
        "mistral",
        "github-copilot",
        "moonshot",
        "kimi-coding",
        "qwen-portal",
        "qianfan",
        "synthetic",
        "minimax",
        "ollama",
    }
)

_PROVIDER_ALIASES: dict[str, str] = {
    "z.ai": "zai",
    "z-ai": "zai",
}


def _normalize_provider(provider: str) -> str:
    key = provider.lower().strip()
    return _PROVIDER_ALIASES.get(key, key)


def canonical_anthropic_auth_choice(input_: str) -> str:
    """Normalize an auth choice string to its canonical form.

    Returns ``"anthropic-api-key"`` or ``"anthropic-setup-token"``.
    Raises ``ValueError`` for invalid input.
    """
    key = input_.lower().strip()
    if key in ("", "api-key", "anthropic-api-key"):
        return "anthropic-api-key"
    if key in ("token", "setup-token", "anthropic-setup-token"):
        return "anthropic-setup-token"
    raise ValueError(
        f"Invalid auth choice {input_!r} (use one of: api-key, setup-token)"
    )


def validate_model(model: str) -> str:
    """Validate and normalize a model reference (``provider/model``).

    Returns the normalized reference, or raises ``ValueError``.
    """
    m = model.strip() if model else ""
    if not m:
        m = DEFAULT_MODEL

    if not _MODEL_REF_RE.match(m):
        raise ValueError(f"Invalid model {m!r} (expected format: provider/model)")

    provider_raw, model_name = m.split("/", 1)
    provider = _normalize_provider(provider_raw)
    if provider not in NATIVE_PROVIDERS:
        raise ValueError(f"Provider {provider!r} is not in OpenClaw native providers")

    return f"{provider}/{model_name}"


class OpenClawConfigBuilder:
    """Fluent builder for openclaw_config JSON."""

    def __init__(self) -> None:
        self._credential: str = ""
        self._auth_choice: str = ""
        self._model: str = ""

    def credential(self, value: str) -> OpenClawConfigBuilder:
        self._credential = value
        return self

    def auth_choice(self, value: str) -> OpenClawConfigBuilder:
        self._auth_choice = value
        return self

    def model(self, value: str) -> OpenClawConfigBuilder:
        self._model = value
        return self

    def build(self) -> str:
        """Build and return the JSON config string.

        Raises ``ValueError`` if inputs are invalid.
        """
        credential = self._credential.strip()
        if not credential:
            raise ValueError("Claude credential is required")

        choice = canonical_anthropic_auth_choice(self._auth_choice)
        model_ref = validate_model(self._model)

        env: dict[str, str] = {"ANTHROPIC_API_KEY": credential}
        if choice == "anthropic-setup-token":
            env["ANTHROPIC_SETUP_TOKEN"] = credential

        config: dict[str, Any] = {
            "env": env,
            "agents": {
                "defaults": {
                    "model": {
                        "primary": model_ref,
                    },
                },
            },
            "clawdepl": {
                "auth": {
                    "provider": "anthropic",
                    "method": choice,
                },
            },
        }

        return json.dumps(config)
