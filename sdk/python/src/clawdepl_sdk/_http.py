from __future__ import annotations

from typing import Any, Optional

import httpx

from .errors import raise_for_status

DEFAULT_BASE_URL = "https://ghcvpnixedcjpokvbsep.supabase.co/functions/v1"
_TIMEOUT = 30.0


class SyncHTTP:
    """Synchronous HTTP transport backed by httpx."""

    def __init__(self, token: str, base_url: str = DEFAULT_BASE_URL) -> None:
        self._base_url = base_url.rstrip("/")
        self._client = httpx.Client(
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
            },
            timeout=_TIMEOUT,
        )

    def post(self, path: str, body: Optional[dict[str, Any]] = None) -> dict[str, Any]:
        url = f"{self._base_url}{path}"
        resp = self._client.post(url, json=body or {})
        return self._handle(resp)

    def post_stream(self, path: str, body: Optional[dict[str, Any]] = None) -> httpx.Response:
        """POST that returns the raw response for SSE streaming."""
        url = f"{self._base_url}{path}"
        return self._client.send(
            self._client.build_request("POST", url, json=body or {}),
            stream=True,
        )

    def close(self) -> None:
        self._client.close()

    @staticmethod
    def _handle(resp: httpx.Response) -> dict[str, Any]:
        if resp.status_code >= 400:
            try:
                body = resp.json()
            except Exception:
                body = {"error": resp.text}
            raise_for_status(resp.status_code, body)
        result: dict[str, Any] = resp.json()
        return result


class AsyncHTTP:
    """Async HTTP transport backed by httpx."""

    def __init__(self, token: str, base_url: str = DEFAULT_BASE_URL) -> None:
        self._base_url = base_url.rstrip("/")
        self._client = httpx.AsyncClient(
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
            },
            timeout=_TIMEOUT,
        )

    async def post(self, path: str, body: Optional[dict[str, Any]] = None) -> dict[str, Any]:
        url = f"{self._base_url}{path}"
        resp = await self._client.post(url, json=body or {})
        return self._handle(resp)

    async def post_stream(self, path: str, body: Optional[dict[str, Any]] = None) -> httpx.Response:
        """POST that returns the raw response for SSE streaming."""
        url = f"{self._base_url}{path}"
        return await self._client.send(
            self._client.build_request("POST", url, json=body or {}),
            stream=True,
        )

    async def close(self) -> None:
        await self._client.aclose()

    @staticmethod
    def _handle(resp: httpx.Response) -> dict[str, Any]:
        if resp.status_code >= 400:
            try:
                body = resp.json()
            except Exception:
                body = {"error": resp.text}
            raise_for_status(resp.status_code, body)
        result: dict[str, Any] = resp.json()
        return result
