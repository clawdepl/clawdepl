import { Bot } from "./bot.js";
import { resolveToken } from "./credentials.js";
import { DEFAULT_BASE_URL, HttpClient } from "./http.js";
import type {
  BotInfo,
  CreateBotOptions,
  CreateBotResult,
  SubscriptionInfo,
  TokenInfo,
} from "./types.js";

export class ClawdeplClient {
  private readonly http: HttpClient;

  constructor(token?: string, baseUrl = DEFAULT_BASE_URL) {
    const resolved = resolveToken(token);
    this.http = new HttpClient(resolved, baseUrl);
  }

  async verifyToken(): Promise<TokenInfo> {
    const data = await this.http.post("/verify-token");
    return {
      valid: (data.valid as boolean) ?? false,
      userId: (data.userId as string) ?? "",
      email: (data.email as string) ?? "",
      name: (data.name as string) ?? "",
    };
  }

  async listBots(): Promise<BotInfo[]> {
    const data = await this.http.post("/list-bots");
    const bots = (data.bots as Record<string, unknown>[]) ?? [];
    return bots.map((b) => ({
      id: b.id as string,
      name: b.name as string,
      sandbox_id: b.sandbox_id as string,
      state: b.state as string,
      created_at: b.created_at as string,
    }));
  }

  async createBot(opts: CreateBotOptions = {}): Promise<CreateBotResult> {
    const body: Record<string, string> = {};
    if (opts.name) body.name = opts.name;
    if (opts.openclawConfig) body.openclaw_config = opts.openclawConfig;
    if (opts.moltyPrompt) body.molty_prompt = opts.moltyPrompt;

    const data = await this.http.post("/create-sandbox", body);
    const sandbox = (data.sandbox as Record<string, unknown>) ?? {};
    return {
      success: (data.success as boolean) ?? false,
      bot_name: (data.bot_name as string) ?? "",
      sandbox: {
        id: (sandbox.id as string) ?? "",
        state: (sandbox.state as string) ?? "",
      },
    };
  }

  bot(sandboxId: string, name = ""): Bot {
    return new Bot(this.http, sandboxId, name);
  }

  async checkSubscription(): Promise<SubscriptionInfo> {
    const data = await this.http.post("/check-subscription");
    return {
      subscribed: (data.subscribed as boolean) ?? false,
      product_id: (data.product_id as string | null) ?? null,
      subscription_end: (data.subscription_end as string | null) ?? null,
      active_bot_count: (data.active_bot_count as number) ?? 0,
    };
  }
}
