import { readFileSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";

function credentialsPath(): string {
  return join(homedir(), ".clawdepl", "credentials.json");
}

function loadTokenFromFile(): string | null {
  try {
    const raw = readFileSync(credentialsPath(), "utf-8");
    const data = JSON.parse(raw) as Record<string, unknown>;

    // Prefer api_key (never expires), fall back to access_token
    if (typeof data.api_key === "string" && data.api_key) {
      return data.api_key;
    }
    if (typeof data.access_token === "string" && data.access_token) {
      return data.access_token;
    }
  } catch {
    // File not found or invalid JSON — not an error
  }
  return null;
}

export function resolveToken(explicit?: string): string {
  if (explicit) return explicit;

  const envToken = process.env["CLAWDEPL_TOKEN"];
  if (envToken) return envToken;

  const fileToken = loadTokenFromFile();
  if (fileToken) return fileToken;

  throw new Error(
    "No clawdepl token found. Provide a token explicitly, set " +
      "CLAWDEPL_TOKEN, or run `clawdepl login`.",
  );
}
