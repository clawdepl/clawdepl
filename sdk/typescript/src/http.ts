import { raiseForStatus } from "./errors.js";

const DEFAULT_BASE_URL =
  "https://ghcvpnixedcjpokvbsep.supabase.co/functions/v1";

export class HttpClient {
  private readonly baseUrl: string;
  private readonly headers: Record<string, string>;

  constructor(token: string, baseUrl = DEFAULT_BASE_URL) {
    this.baseUrl = baseUrl.replace(/\/+$/, "");
    this.headers = {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    };
  }

  async post(
    path: string,
    body: Record<string, unknown> = {},
  ): Promise<Record<string, unknown>> {
    const resp = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: this.headers,
      body: JSON.stringify(body),
    });

    if (!resp.ok) {
      let parsed: Record<string, unknown>;
      try {
        parsed = (await resp.json()) as Record<string, unknown>;
      } catch {
        parsed = { error: await resp.text() };
      }
      raiseForStatus(resp.status, parsed);
    }

    return (await resp.json()) as Record<string, unknown>;
  }

  async postStream(
    path: string,
    body: Record<string, unknown> = {},
  ): Promise<Response> {
    const resp = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: this.headers,
      body: JSON.stringify(body),
    });

    if (!resp.ok) {
      let parsed: Record<string, unknown>;
      try {
        parsed = (await resp.json()) as Record<string, unknown>;
      } catch {
        parsed = { error: await resp.text() };
      }
      raiseForStatus(resp.status, parsed);
    }

    return resp;
  }
}

export { DEFAULT_BASE_URL };
