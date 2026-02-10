export interface TokenInfo {
  valid: boolean;
  userId: string;
  email: string;
  name: string;
}

export interface SandboxInfo {
  id: string;
  state: string;
}

export interface CreateBotResult {
  success: boolean;
  bot_name: string;
  sandbox: SandboxInfo;
}

export interface BotInfo {
  id: string;
  name: string;
  sandbox_id: string;
  state: string;
  created_at: string;
}

export interface SandboxStatus {
  state: string;
  ready: boolean;
}

export interface ExecResult {
  output: string;
  exit_code: number;
}

export interface StreamEvent {
  type: string;
  content?: string;
  cmd_id?: string;
  exit_code?: number;
}

export interface SSHAccess {
  ssh_token: string;
  ssh_command: string;
  expires_in_minutes: number;
  sandbox_id: string;
}

export interface SubscriptionInfo {
  subscribed: boolean;
  product_id: string | null;
  subscription_end: string | null;
  active_bot_count: number;
}

export interface CreateBotOptions {
  name?: string;
  openclawConfig?: string;
  moltyPrompt?: string;
}

export interface WaitUntilReadyOptions {
  pollInterval?: number;
  timeout?: number;
  onProgress?: (state: string, ready: boolean) => void;
}
