const DEFAULT_MODEL = "anthropic/claude-opus-4-6";

const MODEL_REF_RE = /^[a-z0-9][a-z0-9.\-]*\/[A-Za-z0-9._:\/-]+$/;

const NATIVE_PROVIDERS = new Set([
  "anthropic",
  "openai",
  "openai-codex",
  "opencode",
  "google",
  "google-vertex",
  "google-antigravity",
  "google-gemini-cli",
  "zai",
  "vercel-ai-gateway",
  "openrouter",
  "amazon-bedrock",
  "xai",
  "groq",
  "cerebras",
  "mistral",
  "github-copilot",
  "moonshot",
  "kimi-coding",
  "qwen-portal",
  "qianfan",
  "synthetic",
  "minimax",
  "ollama",
]);

const PROVIDER_ALIASES: Record<string, string> = {
  "z.ai": "zai",
  "z-ai": "zai",
};

function normalizeProvider(provider: string): string {
  const key = provider.toLowerCase().trim();
  return PROVIDER_ALIASES[key] ?? key;
}

export function canonicalAnthropicAuthChoice(input: string): string {
  const key = input.toLowerCase().trim();
  if (key === "" || key === "api-key" || key === "anthropic-api-key") {
    return "anthropic-api-key";
  }
  if (
    key === "token" ||
    key === "setup-token" ||
    key === "anthropic-setup-token"
  ) {
    return "anthropic-setup-token";
  }
  throw new Error(
    `Invalid auth choice "${input}" (use one of: api-key, setup-token)`,
  );
}

export function validateModel(model: string): string {
  let m = model?.trim() || "";
  if (!m) m = DEFAULT_MODEL;

  if (!MODEL_REF_RE.test(m)) {
    throw new Error(`Invalid model "${m}" (expected format: provider/model)`);
  }

  const slashIdx = m.indexOf("/");
  const providerRaw = m.slice(0, slashIdx);
  const modelName = m.slice(slashIdx + 1);
  const provider = normalizeProvider(providerRaw);

  if (!NATIVE_PROVIDERS.has(provider)) {
    throw new Error(
      `Provider "${provider}" is not in OpenClaw native providers`,
    );
  }

  return `${provider}/${modelName}`;
}

export class OpenClawConfigBuilder {
  private _credential = "";
  private _authChoice = "";
  private _model = "";

  credential(value: string): this {
    this._credential = value;
    return this;
  }

  authChoice(value: string): this {
    this._authChoice = value;
    return this;
  }

  model(value: string): this {
    this._model = value;
    return this;
  }

  build(): string {
    const credential = this._credential.trim();
    if (!credential) {
      throw new Error("Claude credential is required");
    }

    const choice = canonicalAnthropicAuthChoice(this._authChoice);
    const modelRef = validateModel(this._model);

    const env: Record<string, string> = {
      ANTHROPIC_API_KEY: credential,
    };
    if (choice === "anthropic-setup-token") {
      env["ANTHROPIC_SETUP_TOKEN"] = credential;
    }

    const config = {
      env,
      agents: {
        defaults: {
          model: {
            primary: modelRef,
          },
        },
      },
      clawdepl: {
        auth: {
          provider: "anthropic",
          method: choice,
        },
      },
    };

    return JSON.stringify(config);
  }
}
