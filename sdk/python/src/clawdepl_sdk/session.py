from __future__ import annotations

from typing import AsyncIterator, Iterator, Optional

from ._http import AsyncHTTP, SyncHTTP
from ._sse import iter_sse_async, iter_sse_sync
from .types import ExecResult, StreamEvent


class Session:
    """A terminal session inside a sandbox (sync)."""

    def __init__(self, http: SyncHTTP, sandbox_id: str, session_id: str) -> None:
        self._http = http
        self._sandbox_id = sandbox_id
        self._session_id = session_id

    @property
    def sandbox_id(self) -> str:
        return self._sandbox_id

    @property
    def session_id(self) -> str:
        return self._session_id

    def exec(self, command: str) -> ExecResult:
        """Execute a command synchronously and return the result."""
        data = self._http.post(
            "/sandbox-exec",
            {
                "sandbox_id": self._sandbox_id,
                "action": "exec",
                "session_id": self._session_id,
                "command": command,
            },
        )
        return ExecResult(output=data.get("output", ""), exit_code=data.get("exit_code", -1))

    def stream_exec(self, command: str, timeout: Optional[int] = None) -> Iterator[StreamEvent]:
        """Execute a command with SSE streaming output."""
        body: dict[str, object] = {
            "sandbox_id": self._sandbox_id,
            "action": "stream-exec",
            "session_id": self._session_id,
            "command": command,
        }
        if timeout is not None:
            body["timeout"] = timeout
        resp = self._http.post_stream("/sandbox-exec", body)
        return iter_sse_sync(resp)

    def delete(self) -> None:
        """Delete this session."""
        self._http.post(
            "/sandbox-exec",
            {
                "sandbox_id": self._sandbox_id,
                "action": "delete-session",
                "session_id": self._session_id,
            },
        )

    def __enter__(self) -> Session:
        return self

    def __exit__(self, *_: object) -> None:
        self.delete()


class AsyncSession:
    """A terminal session inside a sandbox (async)."""

    def __init__(self, http: AsyncHTTP, sandbox_id: str, session_id: str) -> None:
        self._http = http
        self._sandbox_id = sandbox_id
        self._session_id = session_id

    @property
    def sandbox_id(self) -> str:
        return self._sandbox_id

    @property
    def session_id(self) -> str:
        return self._session_id

    async def exec(self, command: str) -> ExecResult:
        """Execute a command synchronously and return the result."""
        data = await self._http.post(
            "/sandbox-exec",
            {
                "sandbox_id": self._sandbox_id,
                "action": "exec",
                "session_id": self._session_id,
                "command": command,
            },
        )
        return ExecResult(output=data.get("output", ""), exit_code=data.get("exit_code", -1))

    async def stream_exec(
        self, command: str, timeout: Optional[int] = None
    ) -> AsyncIterator[StreamEvent]:
        """Execute a command with SSE streaming output."""
        body: dict[str, object] = {
            "sandbox_id": self._sandbox_id,
            "action": "stream-exec",
            "session_id": self._session_id,
            "command": command,
        }
        if timeout is not None:
            body["timeout"] = timeout
        resp = await self._http.post_stream("/sandbox-exec", body)
        return iter_sse_async(resp)

    async def delete(self) -> None:
        """Delete this session."""
        await self._http.post(
            "/sandbox-exec",
            {
                "sandbox_id": self._sandbox_id,
                "action": "delete-session",
                "session_id": self._session_id,
            },
        )

    async def __aenter__(self) -> AsyncSession:
        return self

    async def __aexit__(self, *_: object) -> None:
        await self.delete()
