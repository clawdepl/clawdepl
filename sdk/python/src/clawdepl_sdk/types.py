from __future__ import annotations

from dataclasses import dataclass
from typing import Optional


@dataclass(frozen=True)
class TokenInfo:
    valid: bool
    user_id: str
    email: str
    name: str


@dataclass(frozen=True)
class SandboxInfo:
    id: str
    state: str


@dataclass(frozen=True)
class CreateBotResult:
    success: bool
    bot_name: str
    sandbox: SandboxInfo


@dataclass(frozen=True)
class BotInfo:
    id: str
    name: str
    sandbox_id: str
    state: str
    created_at: str


@dataclass(frozen=True)
class SandboxStatus:
    state: str
    ready: bool


@dataclass(frozen=True)
class ExecResult:
    output: str
    exit_code: int


@dataclass(frozen=True)
class StreamEvent:
    type: str
    content: Optional[str] = None
    cmd_id: Optional[str] = None
    exit_code: Optional[int] = None


@dataclass(frozen=True)
class SSHAccess:
    ssh_token: str
    ssh_command: str
    expires_in_minutes: int
    sandbox_id: str


@dataclass(frozen=True)
class SubscriptionInfo:
    subscribed: bool
    product_id: Optional[str]
    subscription_end: Optional[str]
    active_bot_count: int
