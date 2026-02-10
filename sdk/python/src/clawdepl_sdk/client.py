from __future__ import annotations

from typing import Optional

from ._credentials import resolve_token
from ._http import DEFAULT_BASE_URL, AsyncHTTP, SyncHTTP
from .bot import AsyncBot, Bot
from .types import BotInfo, CreateBotResult, SandboxInfo, SubscriptionInfo, TokenInfo


class ClawdeplClient:
    """Synchronous clawdepl API client."""

    def __init__(
        self,
        token: Optional[str] = None,
        *,
        base_url: str = DEFAULT_BASE_URL,
    ) -> None:
        resolved = resolve_token(token)
        self._http = SyncHTTP(resolved, base_url)

    def verify_token(self) -> TokenInfo:
        data = self._http.post("/verify-token")
        return TokenInfo(
            valid=data.get("valid", False),
            user_id=data.get("userId", ""),
            email=data.get("email", ""),
            name=data.get("name", ""),
        )

    def list_bots(self) -> list[BotInfo]:
        data = self._http.post("/list-bots")
        return [
            BotInfo(
                id=b["id"],
                name=b["name"],
                sandbox_id=b["sandbox_id"],
                state=b["state"],
                created_at=b["created_at"],
            )
            for b in data.get("bots", [])
        ]

    def create_bot(
        self,
        name: str = "",
        *,
        openclaw_config: str = "",
        molty_prompt: str = "",
    ) -> CreateBotResult:
        body: dict[str, str] = {}
        if name:
            body["name"] = name
        if openclaw_config:
            body["openclaw_config"] = openclaw_config
        if molty_prompt:
            body["molty_prompt"] = molty_prompt

        data = self._http.post("/create-sandbox", body)
        sandbox_data = data.get("sandbox", {})
        return CreateBotResult(
            success=data.get("success", False),
            bot_name=data.get("bot_name", ""),
            sandbox=SandboxInfo(
                id=sandbox_data.get("id", ""),
                state=sandbox_data.get("state", ""),
            ),
        )

    def bot(self, sandbox_id: str, name: str = "") -> Bot:
        """Create a Bot handle for an existing sandbox."""
        return Bot(self._http, sandbox_id, name)

    def check_subscription(self) -> SubscriptionInfo:
        data = self._http.post("/check-subscription")
        return SubscriptionInfo(
            subscribed=data.get("subscribed", False),
            product_id=data.get("product_id"),
            subscription_end=data.get("subscription_end"),
            active_bot_count=data.get("active_bot_count", 0),
        )

    def close(self) -> None:
        self._http.close()

    def __enter__(self) -> ClawdeplClient:
        return self

    def __exit__(self, *_: object) -> None:
        self.close()


class AsyncClawdeplClient:
    """Async clawdepl API client."""

    def __init__(
        self,
        token: Optional[str] = None,
        *,
        base_url: str = DEFAULT_BASE_URL,
    ) -> None:
        resolved = resolve_token(token)
        self._http = AsyncHTTP(resolved, base_url)

    async def verify_token(self) -> TokenInfo:
        data = await self._http.post("/verify-token")
        return TokenInfo(
            valid=data.get("valid", False),
            user_id=data.get("userId", ""),
            email=data.get("email", ""),
            name=data.get("name", ""),
        )

    async def list_bots(self) -> list[BotInfo]:
        data = await self._http.post("/list-bots")
        return [
            BotInfo(
                id=b["id"],
                name=b["name"],
                sandbox_id=b["sandbox_id"],
                state=b["state"],
                created_at=b["created_at"],
            )
            for b in data.get("bots", [])
        ]

    async def create_bot(
        self,
        name: str = "",
        *,
        openclaw_config: str = "",
        molty_prompt: str = "",
    ) -> CreateBotResult:
        body: dict[str, str] = {}
        if name:
            body["name"] = name
        if openclaw_config:
            body["openclaw_config"] = openclaw_config
        if molty_prompt:
            body["molty_prompt"] = molty_prompt

        data = await self._http.post("/create-sandbox", body)
        sandbox_data = data.get("sandbox", {})
        return CreateBotResult(
            success=data.get("success", False),
            bot_name=data.get("bot_name", ""),
            sandbox=SandboxInfo(
                id=sandbox_data.get("id", ""),
                state=sandbox_data.get("state", ""),
            ),
        )

    def bot(self, sandbox_id: str, name: str = "") -> AsyncBot:
        """Create an AsyncBot handle for an existing sandbox."""
        return AsyncBot(self._http, sandbox_id, name)

    async def check_subscription(self) -> SubscriptionInfo:
        data = await self._http.post("/check-subscription")
        return SubscriptionInfo(
            subscribed=data.get("subscribed", False),
            product_id=data.get("product_id"),
            subscription_end=data.get("subscription_end"),
            active_bot_count=data.get("active_bot_count", 0),
        )

    async def close(self) -> None:
        await self._http.close()

    async def __aenter__(self) -> AsyncClawdeplClient:
        return self

    async def __aexit__(self, *_: object) -> None:
        await self.close()
