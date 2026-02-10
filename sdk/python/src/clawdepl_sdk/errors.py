from __future__ import annotations

from typing import Any, Optional


class ClawdeplError(Exception):
    """Base error for all clawdepl SDK errors."""

    def __init__(self, message: str, status: int = 0) -> None:
        super().__init__(message)
        self.status = status


class AuthError(ClawdeplError):
    """401 — invalid or missing token."""


class LimitReachedError(ClawdeplError):
    """403 — bot limit reached for current tier."""

    def __init__(
        self,
        message: str,
        *,
        tier: Optional[str] = None,
        limit: Optional[int] = None,
        active: Optional[int] = None,
    ) -> None:
        super().__init__(message, status=403)
        self.tier = tier
        self.limit = limit
        self.active = active


class ConflictError(ClawdeplError):
    """409 — duplicate resource (e.g. bot name already taken)."""


class NotFoundError(ClawdeplError):
    """404 — resource not found."""


class ServerError(ClawdeplError):
    """500 — internal server error."""


class UpstreamError(ClawdeplError):
    """502 — upstream (Daytona) failure."""


class TimeoutError(ClawdeplError):
    """Request or operation timed out."""


def raise_for_status(status: int, body: dict[str, Any]) -> None:
    """Map an HTTP error response to the appropriate exception."""
    message = body.get("error", f"HTTP {status}")

    if status == 401:
        raise AuthError(message, status=401)

    if status == 403:
        raise LimitReachedError(
            message,
            tier=body.get("tier"),
            limit=body.get("limit"),
            active=body.get("active"),
        )

    if status == 404:
        raise NotFoundError(message, status=404)

    if status == 409:
        raise ConflictError(message, status=409)

    if status == 502:
        raise UpstreamError(message, status=502)

    if status >= 500:
        raise ServerError(message, status=status)

    raise ClawdeplError(message, status=status)
