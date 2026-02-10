"""clawdepl SDK — Python client for the clawdepl API."""

from .bot import AsyncBot, Bot
from .client import AsyncClawdeplClient, ClawdeplClient
from .config import OpenClawConfigBuilder
from .errors import (
    AuthError,
    ClawdeplError,
    ConflictError,
    LimitReachedError,
    NotFoundError,
    ServerError,
    TimeoutError,
    UpstreamError,
)
from .session import AsyncSession, Session
from .types import (
    BotInfo,
    CreateBotResult,
    ExecResult,
    SandboxInfo,
    SandboxStatus,
    SSHAccess,
    StreamEvent,
    SubscriptionInfo,
    TokenInfo,
)

__all__ = [
    # Clients
    "ClawdeplClient",
    "AsyncClawdeplClient",
    # Resources
    "Bot",
    "AsyncBot",
    "Session",
    "AsyncSession",
    # Config
    "OpenClawConfigBuilder",
    # Types
    "TokenInfo",
    "SandboxInfo",
    "CreateBotResult",
    "BotInfo",
    "SandboxStatus",
    "ExecResult",
    "StreamEvent",
    "SSHAccess",
    "SubscriptionInfo",
    # Errors
    "ClawdeplError",
    "AuthError",
    "LimitReachedError",
    "ConflictError",
    "NotFoundError",
    "ServerError",
    "UpstreamError",
    "TimeoutError",
]
