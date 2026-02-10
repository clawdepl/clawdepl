from __future__ import annotations

import asyncio
import time
from typing import Callable, Optional

from ._http import AsyncHTTP, SyncHTTP
from .errors import ClawdeplError
from .session import AsyncSession, Session
from .types import SSHAccess, SandboxStatus


class Bot:
    """Handle for a single bot/sandbox (sync)."""

    def __init__(self, http: SyncHTTP, sandbox_id: str, name: str = "") -> None:
        self._http = http
        self._sandbox_id = sandbox_id
        self._name = name

    @property
    def sandbox_id(self) -> str:
        return self._sandbox_id

    @property
    def name(self) -> str:
        return self._name

    def status(self) -> SandboxStatus:
        data = self._http.post(
            "/sandbox-exec",
            {"sandbox_id": self._sandbox_id, "action": "check-sandbox-status"},
        )
        return SandboxStatus(state=data.get("state", ""), ready=data.get("ready", False))

    def start(self) -> None:
        self._http.post(
            "/sandbox-exec",
            {"sandbox_id": self._sandbox_id, "action": "start-sandbox"},
        )

    def stop(self) -> None:
        self._http.post(
            "/sandbox-exec",
            {"sandbox_id": self._sandbox_id, "action": "stop-sandbox"},
        )

    def delete(self) -> None:
        self._http.post(
            "/sandbox-exec",
            {"sandbox_id": self._sandbox_id, "action": "delete-sandbox"},
        )

    def create_session(self, session_id: str) -> Session:
        self._http.post(
            "/sandbox-exec",
            {
                "sandbox_id": self._sandbox_id,
                "action": "create-session",
                "session_id": session_id,
            },
        )
        return Session(self._http, self._sandbox_id, session_id)

    def create_ssh(self, expires_in_minutes: int = 60) -> SSHAccess:
        data = self._http.post(
            "/sandbox-terminal-url",
            {
                "sandbox_id": self._sandbox_id,
                "action": "create-ssh",
                "expires_in_minutes": expires_in_minutes,
            },
        )
        return SSHAccess(
            ssh_token=data["ssh_token"],
            ssh_command=data["ssh_command"],
            expires_in_minutes=data["expires_in_minutes"],
            sandbox_id=data["sandbox_id"],
        )

    def revoke_ssh(self, ssh_token: str) -> None:
        self._http.post(
            "/sandbox-terminal-url",
            {
                "sandbox_id": self._sandbox_id,
                "action": "revoke-ssh",
                "ssh_token": ssh_token,
            },
        )

    def wait_until_ready(
        self,
        *,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
        on_progress: Optional[Callable[[str, bool], None]] = None,
    ) -> SandboxStatus:
        """Poll until the sandbox is ready, or raise on failure/timeout."""
        deadline = time.monotonic() + timeout if timeout else None
        while True:
            st = self.status()
            if on_progress is not None:
                on_progress(st.state, st.ready)
            if st.ready:
                return st
            if st.state in ("failed", "error"):
                raise ClawdeplError(f"Sandbox entered {st.state} state")
            if deadline is not None and time.monotonic() >= deadline:
                raise ClawdeplError("Timed out waiting for sandbox to become ready")
            time.sleep(poll_interval)


class AsyncBot:
    """Handle for a single bot/sandbox (async)."""

    def __init__(self, http: AsyncHTTP, sandbox_id: str, name: str = "") -> None:
        self._http = http
        self._sandbox_id = sandbox_id
        self._name = name

    @property
    def sandbox_id(self) -> str:
        return self._sandbox_id

    @property
    def name(self) -> str:
        return self._name

    async def status(self) -> SandboxStatus:
        data = await self._http.post(
            "/sandbox-exec",
            {"sandbox_id": self._sandbox_id, "action": "check-sandbox-status"},
        )
        return SandboxStatus(state=data.get("state", ""), ready=data.get("ready", False))

    async def start(self) -> None:
        await self._http.post(
            "/sandbox-exec",
            {"sandbox_id": self._sandbox_id, "action": "start-sandbox"},
        )

    async def stop(self) -> None:
        await self._http.post(
            "/sandbox-exec",
            {"sandbox_id": self._sandbox_id, "action": "stop-sandbox"},
        )

    async def delete(self) -> None:
        await self._http.post(
            "/sandbox-exec",
            {"sandbox_id": self._sandbox_id, "action": "delete-sandbox"},
        )

    async def create_session(self, session_id: str) -> AsyncSession:
        await self._http.post(
            "/sandbox-exec",
            {
                "sandbox_id": self._sandbox_id,
                "action": "create-session",
                "session_id": session_id,
            },
        )
        return AsyncSession(self._http, self._sandbox_id, session_id)

    async def create_ssh(self, expires_in_minutes: int = 60) -> SSHAccess:
        data = await self._http.post(
            "/sandbox-terminal-url",
            {
                "sandbox_id": self._sandbox_id,
                "action": "create-ssh",
                "expires_in_minutes": expires_in_minutes,
            },
        )
        return SSHAccess(
            ssh_token=data["ssh_token"],
            ssh_command=data["ssh_command"],
            expires_in_minutes=data["expires_in_minutes"],
            sandbox_id=data["sandbox_id"],
        )

    async def revoke_ssh(self, ssh_token: str) -> None:
        await self._http.post(
            "/sandbox-terminal-url",
            {
                "sandbox_id": self._sandbox_id,
                "action": "revoke-ssh",
                "ssh_token": ssh_token,
            },
        )

    async def wait_until_ready(
        self,
        *,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
        on_progress: Optional[Callable[[str, bool], None]] = None,
    ) -> SandboxStatus:
        """Poll until the sandbox is ready, or raise on failure/timeout."""
        deadline = asyncio.get_event_loop().time() + timeout if timeout else None
        while True:
            st = await self.status()
            if on_progress is not None:
                on_progress(st.state, st.ready)
            if st.ready:
                return st
            if st.state in ("failed", "error"):
                raise ClawdeplError(f"Sandbox entered {st.state} state")
            if deadline is not None and asyncio.get_event_loop().time() >= deadline:
                raise ClawdeplError("Timed out waiting for sandbox to become ready")
            await asyncio.sleep(poll_interval)
