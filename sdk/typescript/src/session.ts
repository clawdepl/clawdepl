import type { HttpClient } from "./http.js";
import { iterSSE } from "./sse.js";
import type { ExecResult, StreamEvent } from "./types.js";

export class Session {
  private readonly http: HttpClient;
  public readonly sandboxId: string;
  public readonly sessionId: string;

  constructor(http: HttpClient, sandboxId: string, sessionId: string) {
    this.http = http;
    this.sandboxId = sandboxId;
    this.sessionId = sessionId;
  }

  async exec(command: string): Promise<ExecResult> {
    const data = await this.http.post("/sandbox-exec", {
      sandbox_id: this.sandboxId,
      action: "exec",
      session_id: this.sessionId,
      command,
    });
    return {
      output: (data.output as string) ?? "",
      exit_code: (data.exit_code as number) ?? -1,
    };
  }

  async *streamExec(
    command: string,
    timeout?: number,
  ): AsyncGenerator<StreamEvent> {
    const body: Record<string, unknown> = {
      sandbox_id: this.sandboxId,
      action: "stream-exec",
      session_id: this.sessionId,
      command,
    };
    if (timeout !== undefined) {
      body.timeout = timeout;
    }
    const resp = await this.http.postStream("/sandbox-exec", body);
    yield* iterSSE(resp);
  }

  async delete(): Promise<void> {
    await this.http.post("/sandbox-exec", {
      sandbox_id: this.sandboxId,
      action: "delete-session",
      session_id: this.sessionId,
    });
  }

  async [Symbol.asyncDispose](): Promise<void> {
    await this.delete();
  }
}
