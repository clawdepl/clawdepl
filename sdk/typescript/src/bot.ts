import { ClawdeplError } from "./errors.js";
import type { HttpClient } from "./http.js";
import { Session } from "./session.js";
import type { SSHAccess, SandboxStatus, WaitUntilReadyOptions } from "./types.js";

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export class Bot {
  private readonly http: HttpClient;
  public readonly sandboxId: string;
  public readonly name: string;

  constructor(http: HttpClient, sandboxId: string, name = "") {
    this.http = http;
    this.sandboxId = sandboxId;
    this.name = name;
  }

  async status(): Promise<SandboxStatus> {
    const data = await this.http.post("/sandbox-exec", {
      sandbox_id: this.sandboxId,
      action: "check-sandbox-status",
    });
    return {
      state: (data.state as string) ?? "",
      ready: (data.ready as boolean) ?? false,
    };
  }

  async start(): Promise<void> {
    await this.http.post("/sandbox-exec", {
      sandbox_id: this.sandboxId,
      action: "start-sandbox",
    });
  }

  async stop(): Promise<void> {
    await this.http.post("/sandbox-exec", {
      sandbox_id: this.sandboxId,
      action: "stop-sandbox",
    });
  }

  async delete(): Promise<void> {
    await this.http.post("/sandbox-exec", {
      sandbox_id: this.sandboxId,
      action: "delete-sandbox",
    });
  }

  async createSession(sessionId: string): Promise<Session> {
    await this.http.post("/sandbox-exec", {
      sandbox_id: this.sandboxId,
      action: "create-session",
      session_id: sessionId,
    });
    return new Session(this.http, this.sandboxId, sessionId);
  }

  async createSSH(expiresInMinutes = 60): Promise<SSHAccess> {
    const data = await this.http.post("/sandbox-terminal-url", {
      sandbox_id: this.sandboxId,
      action: "create-ssh",
      expires_in_minutes: expiresInMinutes,
    });
    return {
      ssh_token: data.ssh_token as string,
      ssh_command: data.ssh_command as string,
      expires_in_minutes: data.expires_in_minutes as number,
      sandbox_id: data.sandbox_id as string,
    };
  }

  async revokeSSH(sshToken: string): Promise<void> {
    await this.http.post("/sandbox-terminal-url", {
      sandbox_id: this.sandboxId,
      action: "revoke-ssh",
      ssh_token: sshToken,
    });
  }

  async waitUntilReady(opts: WaitUntilReadyOptions = {}): Promise<SandboxStatus> {
    const pollInterval = opts.pollInterval ?? 2000;
    const deadline = opts.timeout ? Date.now() + opts.timeout : null;

    while (true) {
      const st = await this.status();
      opts.onProgress?.(st.state, st.ready);

      if (st.ready) return st;

      if (st.state === "failed" || st.state === "error") {
        throw new ClawdeplError(`Sandbox entered ${st.state} state`);
      }

      if (deadline !== null && Date.now() >= deadline) {
        throw new ClawdeplError(
          "Timed out waiting for sandbox to become ready",
        );
      }

      await sleep(pollInterval);
    }
  }
}
