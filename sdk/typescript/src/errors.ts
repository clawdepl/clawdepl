export class ClawdeplError extends Error {
  public readonly status: number;

  constructor(message: string, status = 0) {
    super(message);
    this.name = "ClawdeplError";
    this.status = status;
  }
}

export class AuthError extends ClawdeplError {
  constructor(message: string) {
    super(message, 401);
    this.name = "AuthError";
  }
}

export class LimitReachedError extends ClawdeplError {
  public readonly tier?: string;
  public readonly limit?: number;
  public readonly active?: number;

  constructor(
    message: string,
    opts?: { tier?: string; limit?: number; active?: number },
  ) {
    super(message, 403);
    this.name = "LimitReachedError";
    this.tier = opts?.tier;
    this.limit = opts?.limit;
    this.active = opts?.active;
  }
}

export class ConflictError extends ClawdeplError {
  constructor(message: string) {
    super(message, 409);
    this.name = "ConflictError";
  }
}

export class NotFoundError extends ClawdeplError {
  constructor(message: string) {
    super(message, 404);
    this.name = "NotFoundError";
  }
}

export class ServerError extends ClawdeplError {
  constructor(message: string, status = 500) {
    super(message, status);
    this.name = "ServerError";
  }
}

export class UpstreamError extends ClawdeplError {
  constructor(message: string) {
    super(message, 502);
    this.name = "UpstreamError";
  }
}

export class TimeoutError extends ClawdeplError {
  constructor(message: string) {
    super(message, 408);
    this.name = "TimeoutError";
  }
}

export function raiseForStatus(
  status: number,
  body: Record<string, unknown>,
): never {
  const message = (body.error as string) ?? `HTTP ${status}`;

  if (status === 401) throw new AuthError(message);
  if (status === 403) {
    throw new LimitReachedError(message, {
      tier: body.tier as string | undefined,
      limit: body.limit as number | undefined,
      active: body.active as number | undefined,
    });
  }
  if (status === 404) throw new NotFoundError(message);
  if (status === 409) throw new ConflictError(message);
  if (status === 502) throw new UpstreamError(message);
  if (status >= 500) throw new ServerError(message, status);

  throw new ClawdeplError(message, status);
}
