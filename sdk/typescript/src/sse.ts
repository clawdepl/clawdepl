import type { StreamEvent } from "./types.js";

function parseSSELine(line: string): StreamEvent | null {
  const trimmed = line.trim();
  if (!trimmed || trimmed.startsWith(":")) return null;
  if (!trimmed.startsWith("data:")) return null;

  const payload = trimmed.slice("data:".length).trim();
  if (!payload) return null;

  const data = JSON.parse(payload) as Record<string, unknown>;
  return {
    type: (data.type as string) ?? "",
    content: data.content as string | undefined,
    cmd_id: data.cmd_id as string | undefined,
    exit_code: data.exit_code as number | undefined,
  };
}

export async function* iterSSE(response: Response): AsyncGenerator<StreamEvent> {
  const body = response.body;
  if (!body) return;

  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });

      // Split on newlines — SSE events are separated by \n\n but data lines by \n
      const lines = buffer.split("\n");
      // Keep the last partial line in the buffer
      buffer = lines.pop() ?? "";

      for (const line of lines) {
        const event = parseSSELine(line);
        if (event) yield event;
      }
    }

    // Process any remaining content in buffer
    if (buffer.trim()) {
      const event = parseSSELine(buffer);
      if (event) yield event;
    }
  } finally {
    reader.releaseLock();
  }
}
