from __future__ import annotations

import json
from typing import AsyncIterator, Iterator, Optional

import httpx

from .types import StreamEvent


def _parse_sse_line(line: str) -> Optional[StreamEvent]:
    """Parse a single SSE ``data: {...}`` line into a StreamEvent.

    Returns None for blank lines, comments, or keepalives.
    """
    stripped = line.strip()
    if not stripped or stripped.startswith(":"):
        return None
    if not stripped.startswith("data:"):
        return None
    payload = stripped[len("data:"):].strip()
    if not payload:
        return None

    data = json.loads(payload)
    return StreamEvent(
        type=data.get("type", ""),
        content=data.get("content"),
        cmd_id=data.get("cmd_id"),
        exit_code=data.get("exit_code"),
    )


def iter_sse_sync(response: httpx.Response) -> Iterator[StreamEvent]:
    """Yield StreamEvent objects from a streaming httpx response (sync)."""
    try:
        for line in response.iter_lines():
            event = _parse_sse_line(line)
            if event is not None:
                yield event
    finally:
        response.close()


async def iter_sse_async(response: httpx.Response) -> AsyncIterator[StreamEvent]:
    """Yield StreamEvent objects from a streaming httpx response (async)."""
    try:
        async for line in response.aiter_lines():
            event = _parse_sse_line(line)
            if event is not None:
                yield event
    finally:
        await response.aclose()
