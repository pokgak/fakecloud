// The Account Durable Object: one per GitHub user. Tracks the user's
// sandboxes and enforces the per-user quota, so a single login can't mint
// unlimited playgrounds.

import { DurableObject } from "cloudflare:workers";
import type { Env } from "./index";

const MAX_SANDBOXES_PER_USER = 5;

export interface SandboxRecord {
  id: string;
  created_at: string;
}

export class Account extends DurableObject<Env> {
  /** Registers a sandbox for this user; throws if the quota is reached. */
  async registerSandbox(id: string): Promise<SandboxRecord[]> {
    const sandboxes = (await this.ctx.storage.get<SandboxRecord[]>("sandboxes")) ?? [];
    if (sandboxes.length >= MAX_SANDBOXES_PER_USER) {
      throw new Error(
        `playground limit reached (${MAX_SANDBOXES_PER_USER} per account) — delete one from the landing page first`,
      );
    }
    sandboxes.push({ id, created_at: new Date().toISOString() });
    await this.ctx.storage.put("sandboxes", sandboxes);
    return sandboxes;
  }

  async listSandboxes(): Promise<SandboxRecord[]> {
    return (await this.ctx.storage.get<SandboxRecord[]>("sandboxes")) ?? [];
  }

  async removeSandbox(id: string): Promise<boolean> {
    const sandboxes = (await this.ctx.storage.get<SandboxRecord[]>("sandboxes")) ?? [];
    const remaining = sandboxes.filter((s) => s.id !== id);
    if (remaining.length === sandboxes.length) return false;
    await this.ctx.storage.put("sandboxes", remaining);
    return true;
  }
}
