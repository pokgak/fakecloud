// GitHub OAuth + stateless signed-cookie sessions. Auth gates sandbox
// CREATION only — the sandbox API itself is authenticated by the
// unguessable sandbox id (a capability URL), so the Terraform provider
// needs no credentials and duel partners can just share an id.
//
// Locally (wrangler dev), .dev.vars sets REQUIRE_AUTH=false and everyone
// is the anonymous "local" user — same code, no GitHub round-trip.

import type { Env } from "./index";

export interface Session {
  login: string; // GitHub username
  uid: string; // GitHub numeric id, stable across renames
  exp: number; // unix ms
}

const SESSION_COOKIE = "fakecloud_session";
const STATE_COOKIE = "fakecloud_oauth_state";
const SESSION_TTL_MS = 7 * 24 * 60 * 60 * 1000;

const enc = new TextEncoder();

async function hmac(secret: string, data: string): Promise<string> {
  const key = await crypto.subtle.importKey("raw", enc.encode(secret), { name: "HMAC", hash: "SHA-256" }, false, [
    "sign",
  ]);
  const sig = await crypto.subtle.sign("HMAC", key, enc.encode(data));
  return btoa(String.fromCharCode(...new Uint8Array(sig))).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
}

async function sign(secret: string, payload: object): Promise<string> {
  const body = btoa(JSON.stringify(payload)).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
  return `${body}.${await hmac(secret, body)}`;
}

async function verify<T>(secret: string, token: string): Promise<T | null> {
  const [body, sig] = token.split(".");
  if (!body || !sig) return null;
  if ((await hmac(secret, body)) !== sig) return null;
  try {
    return JSON.parse(atob(body.replaceAll("-", "+").replaceAll("_", "/"))) as T;
  } catch {
    return null;
  }
}

function readCookie(request: Request, name: string): string | null {
  const header = request.headers.get("Cookie") ?? "";
  for (const part of header.split(";")) {
    const [k, ...v] = part.trim().split("=");
    if (k === name) return v.join("=");
  }
  return null;
}

function cookie(name: string, value: string, maxAgeSeconds: number): string {
  return `${name}=${value}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=${maxAgeSeconds}`;
}

export function authRequired(env: Env): boolean {
  return env.REQUIRE_AUTH !== "false";
}

/** Returns the session, the anonymous local user when auth is off, or null. */
export async function getSession(request: Request, env: Env): Promise<Session | null> {
  if (!authRequired(env)) {
    return { login: "local", uid: "local", exp: Date.now() + SESSION_TTL_MS };
  }
  const token = readCookie(request, SESSION_COOKIE);
  if (!token) return null;
  const session = await verify<Session>(env.SESSION_SECRET, token);
  if (!session || session.exp < Date.now()) return null;
  return session;
}

export async function handleLogin(request: Request, env: Env): Promise<Response> {
  if (!authRequired(env)) return Response.redirect(new URL("/", request.url).toString(), 302);
  const state = crypto.randomUUID();
  const redirectUri = new URL("/auth/callback", request.url).toString();
  const authorize = new URL("https://github.com/login/oauth/authorize");
  authorize.searchParams.set("client_id", env.GITHUB_CLIENT_ID);
  authorize.searchParams.set("redirect_uri", redirectUri);
  authorize.searchParams.set("state", state);
  return new Response(null, {
    status: 302,
    headers: {
      Location: authorize.toString(),
      "Set-Cookie": cookie(STATE_COOKIE, state, 600),
    },
  });
}

export async function handleCallback(request: Request, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const code = url.searchParams.get("code");
  const state = url.searchParams.get("state");
  if (!code || !state || state !== readCookie(request, STATE_COOKIE)) {
    return new Response("OAuth state mismatch — start again at /auth/login", { status: 400 });
  }

  const tokenResp = await fetch("https://github.com/login/oauth/access_token", {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify({
      client_id: env.GITHUB_CLIENT_ID,
      client_secret: env.GITHUB_CLIENT_SECRET,
      code,
      redirect_uri: new URL("/auth/callback", request.url).toString(),
    }),
  });
  const tokenBody = (await tokenResp.json()) as { access_token?: string };
  if (!tokenBody.access_token) {
    return new Response("GitHub token exchange failed", { status: 502 });
  }

  const userResp = await fetch("https://api.github.com/user", {
    headers: {
      Authorization: `Bearer ${tokenBody.access_token}`,
      "User-Agent": "fakecloud",
      Accept: "application/vnd.github+json",
    },
  });
  if (!userResp.ok) return new Response("GitHub user lookup failed", { status: 502 });
  const user = (await userResp.json()) as { id: number; login: string };

  const session: Session = { login: user.login, uid: String(user.id), exp: Date.now() + SESSION_TTL_MS };
  const token = await sign(env.SESSION_SECRET, session);
  return new Response(null, {
    status: 302,
    headers: {
      Location: new URL("/", request.url).toString(),
      "Set-Cookie": cookie(SESSION_COOKIE, token, SESSION_TTL_MS / 1000),
    },
  });
}

export function handleLogout(request: Request): Response {
  return new Response(null, {
    status: 302,
    headers: {
      Location: new URL("/", request.url).toString(),
      "Set-Cookie": cookie(SESSION_COOKIE, "", 0),
    },
  });
}
