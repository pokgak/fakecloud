// The Sandbox Durable Object: one per playground. Single-threaded with
// durable storage, it plays the same referee role the old Go store's mutex
// did — but isolated per tenant and persistent across deploys.

import { DurableObject } from "cloudflare:workers";
import * as store from "./store";
import type { Env } from "./index";

// A sandbox that sees no traffic for this long is deleted by its alarm.
const IDLE_TTL_MS = 7 * 24 * 60 * 60 * 1000; // 7 days

// Entity caps: a learning playground never needs more, and it keeps a
// leaked sandbox id from becoming a free database.
const MAX_BOARDS = 12;

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });

export class Sandbox extends DurableObject<Env> {
  private state!: store.State;

  constructor(ctx: DurableObjectState, env: Env) {
    super(ctx, env);
    ctx.blockConcurrencyWhile(async () => {
      this.state = (await ctx.storage.get<store.State>("state")) ?? store.emptyState();
    });
  }

  private async persist(): Promise<void> {
    await this.ctx.storage.put("state", this.state);
    await this.ctx.storage.setAlarm(Date.now() + IDLE_TTL_MS);
  }

  async alarm(): Promise<void> {
    await this.ctx.storage.deleteAll();
  }

  /** Wipes the sandbox immediately (owner retired it from the landing page). */
  async wipe(): Promise<void> {
    this.state = store.emptyState();
    await this.ctx.storage.deleteAll();
  }

  async fetch(request: Request): Promise<Response> {
    try {
      return await this.route(request);
    } catch (err) {
      if (err instanceof store.ApiError) {
        return json(err.status, { error: err.message });
      }
      return json(500, { error: err instanceof Error ? err.message : "internal error" });
    }
  }

  // Routes the sandbox-relative path, e.g. /tictactoe/boards/3.
  private async route(request: Request): Promise<Response> {
    const url = new URL(request.url);
    const parts = url.pathname.split("/").filter(Boolean); // ["tictactoe", "boards", "3"]
    const method = request.method;

    if (parts[0] !== "tictactoe") return json(404, { error: "not found" });
    const [, collection, rawId] = parts;
    const id = rawId !== undefined ? Number(rawId) : undefined;
    if (rawId !== undefined && (!Number.isInteger(id) || parts.length > 3)) {
      return json(404, { error: "not found" });
    }

    const body = async () => {
      try {
        return (await request.json()) as Record<string, unknown>;
      } catch {
        throw new store.ApiError(400, "invalid JSON body");
      }
    };

    if (collection === "boards") {
      if (method === "GET" && id === undefined) return json(200, store.listBoards(this.state));
      if (method === "GET") return json(200, store.getBoard(this.state, id!));
      if (method === "POST" && id === undefined) {
        if (Object.keys(this.state.boards).length >= MAX_BOARDS) {
          throw new store.ApiError(409, `sandbox is full (max ${MAX_BOARDS} boards) — delete some first`);
        }
        const b = await body();
        const board = store.createBoard(this.state, b.name, b.mode);
        await this.persist();
        return json(201, board);
      }
      if (method === "DELETE" && id !== undefined) {
        store.deleteBoard(this.state, id);
        await this.persist();
        return json(200, { status: "deleted" });
      }
    }

    if (collection === "moves") {
      if (method === "GET" && id !== undefined) return json(200, store.getMove(this.state, id));
      if (method === "POST" && id === undefined) {
        const b = await body();
        const move = store.createMove(this.state, b.board_id, b.player, b.position);
        await this.persist();
        return json(201, move);
      }
      if (method === "DELETE" && id !== undefined) {
        store.deleteMove(this.state, id);
        await this.persist();
        return json(200, { status: "deleted" });
      }
    }

    if (collection === "nameplates") {
      if (method === "GET" && id !== undefined) return json(200, store.getNameplate(this.state, id));
      if (method === "POST" && id === undefined) {
        const b = await body();
        const plate = store.createNameplate(this.state, b.board_id, b.text);
        await this.persist();
        return json(201, plate);
      }
      if (method === "PUT" && id !== undefined) {
        const b = await body();
        const plate = store.updateNameplate(this.state, id, b.text);
        await this.persist();
        return json(200, plate);
      }
      if (method === "DELETE" && id !== undefined) {
        store.deleteNameplate(this.state, id);
        await this.persist();
        return json(200, { status: "deleted" });
      }
    }

    return json(404, { error: "not found" });
  }
}
