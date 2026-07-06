// fakecloud on Cloudflare Workers — and, via `wrangler dev`, on localhost:
// the exact same code serves both.
//
// URL layout:
//   /                        landing page (static asset)
//   /auth/login|callback|logout   GitHub OAuth (skipped when REQUIRE_AUTH=false)
//   /api/me                  session + your playgrounds
//   /sandboxes               POST: mint a playground (auth + quota)
//   /sandboxes/:id           DELETE: retire a playground you own
//   /s/:id                   dashboard for one playground
//   /s/:id/tictactoe/...     the fakecloud API, forwarded to that playground's DO

import { Sandbox } from "./sandbox";
import { Account } from "./account";
import { getSession, authRequired, handleLogin, handleCallback, handleLogout } from "./auth";

export { Sandbox, Account };

export interface Env {
  ASSETS: Fetcher;
  SANDBOX: DurableObjectNamespace<Sandbox>;
  ACCOUNT: DurableObjectNamespace<Account>;
  REQUIRE_AUTH: string;
  GITHUB_CLIENT_ID: string;
  GITHUB_CLIENT_SECRET: string;
  SESSION_SECRET: string;
}

const SANDBOX_ID = /^[a-z0-9][a-z0-9-]{2,30}[a-z0-9]$/;

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });

function newSandboxId(): string {
  // Unguessable and readable: the id is the capability that protects the
  // sandbox, so it needs real entropy (~50 bits here).
  const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789";
  const bytes = crypto.getRandomValues(new Uint8Array(10));
  return Array.from(bytes, (b) => alphabet[b % 36]).join("");
}

function account(env: Env, uid: string) {
  return env.ACCOUNT.get(env.ACCOUNT.idFromName(`github:${uid}`));
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const path = url.pathname;

    // --- auth ---
    if (path === "/auth/login") return handleLogin(request, env);
    if (path === "/auth/callback") return handleCallback(request, env);
    if (path === "/auth/logout") return handleLogout(request);

    // --- session info for the landing page ---
    if (path === "/api/me" && request.method === "GET") {
      const session = await getSession(request, env);
      if (!session) return json(200, { authenticated: false, auth_required: authRequired(env) });
      const sandboxes = await account(env, session.uid).listSandboxes();
      return json(200, {
        authenticated: true,
        auth_required: authRequired(env),
        login: session.login,
        sandboxes,
      });
    }

    // --- mint a playground ---
    if (path === "/sandboxes" && request.method === "POST") {
      const session = await getSession(request, env);
      if (!session) {
        return json(401, { error: "sign in with GitHub to create a playground", login_url: "/auth/login" });
      }
      const id = newSandboxId();
      try {
        await account(env, session.uid).registerSandbox(id);
      } catch (err) {
        return json(429, { error: err instanceof Error ? err.message : "quota exceeded" });
      }
      return json(201, {
        sandbox: id,
        endpoint: url.origin,
        dashboard_url: `${url.origin}/s/${id}`,
      });
    }

    // --- retire a playground ---
    const retire = path.match(/^\/sandboxes\/([a-z0-9-]+)$/);
    if (retire && request.method === "DELETE") {
      const session = await getSession(request, env);
      if (!session) return json(401, { error: "not signed in" });
      const owned = await account(env, session.uid).removeSandbox(retire[1]);
      if (!owned) return json(404, { error: "not your playground" });
      const stub = env.SANDBOX.get(env.SANDBOX.idFromName(retire[1]));
      await stub.wipe();
      return json(200, { status: "deleted" });
    }

    // --- per-sandbox dashboard + API ---
    const sandboxMatch = path.match(/^\/s\/([^/]+)(\/.*)?$/);
    if (sandboxMatch) {
      const [, id, rest] = sandboxMatch;
      if (!SANDBOX_ID.test(id)) return json(404, { error: "invalid sandbox id" });

      // The dashboard page itself. Fetch the asset by its extension-less
      // path: asking for /dashboard.html would make the assets layer
      // respond with a redirect to /dashboard, which the browser would
      // follow — losing the /s/:id URL the dashboard reads its sandbox
      // id from.
      if (!rest || rest === "/") {
        return env.ASSETS.fetch(new Request(new URL("/dashboard", url.origin), request));
      }

      // Everything else goes to the sandbox's Durable Object with the
      // /s/:id prefix stripped, so the DO speaks the classic fakecloud API.
      const stub = env.SANDBOX.get(env.SANDBOX.idFromName(id));
      const forwarded = new URL(rest, url.origin);
      return stub.fetch(new Request(forwarded.toString(), request));
    }

    // A provider pointed at the root without a sandbox: explain the fix
    // instead of 404ing mysteriously.
    if (path.startsWith("/tictactoe/")) {
      return json(400, {
        error:
          `no sandbox in the URL — this is the shared fakecloud. Visit ${url.origin} to create your ` +
          `playground, then set the provider's "sandbox" attribute (or FAKECLOUD_SANDBOX).`,
      });
    }

    // Everything else (/, static files) is an asset.
    return env.ASSETS.fetch(request);
  },
} satisfies ExportedHandler<Env>;
