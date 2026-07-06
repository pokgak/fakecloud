# fakecloud

**Learn Terraform on a tic-tac-toe board.**

fakecloud is a pretend cloud with exactly one primitive: a tic-tac-toe board
where every mark is a resource. You manage it with a real Terraform provider,
and a live dashboard shows every `apply`, `destroy`, and drift event as it
happens — green `+`, yellow `~`, red `−`, just like a plan.

The same board teaches the whole course: resource basics on it, `count`
footguns tear it apart, drift and imports happen when you click it. And when
you're done learning, flip it to `duel` mode and play a friend, one
`terraform apply` at a time.

<picture>
  <source media="(prefers-color-scheme: light)" srcset="docs/dashboard-light.png">
  <img alt="the fakecloud dashboard" src="docs/dashboard.png">
</picture>

The dashboard follows your OS theme — light and dark both included.

## The course

Each chapter is a directory under `examples/` with a `README.md` of missions
and hints and a starter `main.tf` of TODOs. **You write every line of
Terraform yourself** — the hints name the primitives and the expected plan
output, never the code. Keep the dashboard visible while you work. If you're
truly stuck, an answer key lives on the [`solutions`](../../tree/solutions)
branch.

| Chapter | What the board teaches |
|---|---|
| [`chapter-1-basics`](examples/chapter-1-basics) | Provider & resource blocks, references, variables, outputs, the plan/apply/destroy loop, idempotence, why some changes are "replace". |
| [`chapter-2-count-vs-foreach`](examples/chapter-2-count-vs-foreach) | Paint a diagonal with `count`, remove the first list element, and watch the index-shift footgun destroy and recreate marks you never touched. Then `for_each` does it surgically. |
| [`chapter-3-drift-and-import`](examples/chapter-3-drift-and-import) | The console is your coworker who clicks things. Unmanaged infra (plan says nothing!), real drift on managed resources, and adopting strays with `terraform import` / `import` blocks. |
| [`chapter-4-modules`](examples/chapter-4-modules) | Write a glyph module, stamp board+marks patterns from one block, then `for_each` a whole gallery into existence. Includes the rename footgun and its `moved {}` cure. |
| [`chapter-5-shared-state`](examples/chapter-5-shared-state) | Two state files import the same nameplate 🪧 (the only in-place-updatable resource) and fight an **apply war**. Fixes: `ignore_changes`, single owner + data sources, and why locking backends exist. Plus `prevent_destroy` and the `create_before_destroy` collision footgun. |
| [`chapter-6-dependencies`](examples/chapter-6-dependencies) | Terraform runs a graph, not a script: references are edges, apply order is visible live in the feed, `depends_on` adds edges by hand, and `data.fakecloud_tictactoe_board.opponent.id` reaches infrastructure you don't manage — and shows where the graph stops. |
| [`side-quest-duel`](examples/side-quest-duel) | A `mode = "duel"` board is refereed: X starts, turns alternate, out-of-turn applies fail with the reason. Two state files, one board, best strategist wins. |

## Quick start

**1. Get a playground.** Open the hosted fakecloud, sign in with GitHub
(sign-in only gates playground creation — nothing touches your repos), click
**+ new playground**, and keep the dashboard open. Its "Connect Terraform"
panel shows the provider block with your sandbox id.

Prefer local? The hosted version and the local version are the same code:

```sh
cd server && npm install && npx wrangler dev
```

…then open <http://localhost:8787> — no sign-in needed locally.

**2. Build the provider** and tell Terraform to use your local build via a
[dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides):

```sh
cd provider && go build -o ~/go/bin/terraform-provider-fakecloud .
```

```hcl
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "pokgak/fakecloud" = "/home/YOU/go/bin"   # dir containing the binary
  }
  direct {}
}
```

**3. Start chapter 1** (skip `terraform init` — dev overrides don't need it):

```sh
cd examples/chapter-1-basics
```

…and open its `README.md`.

## What's in the box

| Directory   | What it is |
|-------------|------------|
| `server/`   | The fakecloud platform: a Cloudflare Worker (TypeScript) serving the API, dashboard, and landing page. One Durable Object per playground — a single-threaded referee with durable state. `wrangler dev` runs the identical code locally. |
| `provider/` | `terraform-provider-fakecloud`, built on terraform-plugin-framework (Go). |
| `examples/` | The course. |

The provider ships three resources and a data source:

- `fakecloud_tictactoe_board` (resource + data source) — `name`, `mode`
  (`freeplay` default, or `duel`); computed `cells`, `next_player`, `winner`,
  `nameplate_text`
- `fakecloud_tictactoe_move` (resource) — `board_id`, `player`, `position`;
  create = play, destroy = take it back, importable by id
- `fakecloud_nameplate` (resource) — a plaque on a board, one per board;
  `text` updates **in place**, which is what makes the chapter 5 apply war
  possible

Provider configuration: `sandbox` (your playground id, or
`FAKECLOUD_SANDBOX`) and `endpoint` (defaults to the hosted fakecloud, or
`FAKECLOUD_ENDPOINT`; use `http://localhost:8787` with `wrangler dev`).

## The API

Everything the provider does is plain JSON over HTTP under your playground's
prefix, so `curl` works too:

| Method & path (under `/s/<sandbox>`) | Effect |
|---|---|
| `GET/POST /tictactoe/boards`, `GET/DELETE /tictactoe/boards/{id}` | boards (reads include derived `cells`, `next_player`, `winner`, `moves`, `nameplate`) |
| `POST /tictactoe/moves`, `GET/DELETE /tictactoe/moves/{id}` | play / inspect / take back a move |
| `POST /tictactoe/nameplates`, `GET/PUT/DELETE /tictactoe/nameplates/{id}` | nameplates (one per board; creating a second 409s and tells you to import) |

Playground management (top level): `POST /sandboxes` (requires GitHub
sign-in on the hosted version), `DELETE /sandboxes/{id}`, `GET /api/me`.
Playgrounds idle for 30 days are deleted; a sandbox holds at most 25 boards.
The sandbox id is a capability — anyone who has it can use the playground,
which is exactly how you invite a duel opponent.

## Deploying your own

```sh
cd server
npx wrangler deploy
```

Then, to enable GitHub sign-in on your deployment:

1. Create a [GitHub OAuth app](https://github.com/settings/developers) with
   callback URL `https://<your-worker-host>/auth/callback`.
2. Put the client id in `wrangler.jsonc` (`GITHUB_CLIENT_ID`), and set the
   secrets:

   ```sh
   npx wrangler secret put GITHUB_CLIENT_SECRET
   npx wrangler secret put SESSION_SECRET     # any long random string
   ```

3. Update `DefaultEndpoint` in `provider/internal/provider/provider.go` to
   your worker's URL so learners get your deployment by default.

Auth gates playground *creation* only; the per-playground API is
authenticated by the unguessable sandbox id, so the Terraform provider never
needs credentials.
