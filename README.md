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

![dashboard](docs/dashboard.png)

## The course

Each chapter is a directory under `examples/` — open its `main.tf`, follow
the comments, and keep the dashboard visible.

| Chapter | What the board teaches |
|---|---|
| [`chapter-1-basics`](examples/chapter-1-basics) | Provider & resource blocks, references, variables, outputs, the plan/apply/destroy loop, idempotence, why some changes are "replace". |
| [`chapter-2-count-vs-foreach`](examples/chapter-2-count-vs-foreach) | Paint a diagonal with `count`, remove the first list element, and watch the index-shift footgun destroy and recreate marks you never touched — sometimes failing mid-apply on cell collisions. Then `for_each` does it surgically. |
| [`chapter-3-drift-and-import`](examples/chapter-3-drift-and-import) | The console is your coworker who clicks things. Unmanaged infra (plan says nothing!), real drift on managed resources, and adopting strays with `terraform import` / `import` blocks. |
| [`chapter-4-modules`](examples/chapter-4-modules) | A glyph module stamps board+marks patterns from one block, then `for_each` over the module conjures a whole gallery. Includes the rename footgun and its `moved {}` cure. |
| [`chapter-5-shared-state`](examples/chapter-5-shared-state) | Two state files import the same nameplate 🪧 (the only in-place-updatable resource) and fight an **apply war** — yellow `~` ping-pong in the feed. Fixes: `ignore_changes`, single owner + data sources, and why locking backends exist. Plus `prevent_destroy` and the `create_before_destroy` collision footgun. |
| [`chapter-6-dependencies`](examples/chapter-6-dependencies) | Terraform runs a graph, not a script: references are edges, apply order is visible live in the feed, `depends_on` adds edges by hand, and `data.fakecloud_tictactoe_board.opponent.id` shows the graph reaching infrastructure you don't manage — and where it stops. |
| [`side-quest-duel`](examples/side-quest-duel) | A `mode = "duel"` board is refereed: X starts, turns alternate, out-of-turn applies fail with the reason. Two state files, one board, best strategist wins. |

## What's in the box

| Directory   | What it is |
|-------------|------------|
| `server/`   | The fakecloud API + embedded dashboard. Pure Go stdlib, in-memory state. |
| `provider/` | `terraform-provider-fakecloud`, built on terraform-plugin-framework. |
| `examples/` | The chapters above. |

The provider ships three resources and a data source:

- `fakecloud_tictactoe_board` (resource + data source) — `name`, `mode`
  (`freeplay` default, or `duel`); computed `cells`, `next_player`, `winner`,
  `nameplate_text`
- `fakecloud_tictactoe_move` (resource) — `board_id`, `player`, `position`;
  create = play, destroy = take it back, importable by id
- `fakecloud_nameplate` (resource) — a plaque on a board, one per board;
  `text` updates **in place**, which is what makes the chapter 5 apply war
  possible

## Quick start

**1. Run the server** (no dependencies):

```sh
cd server && go run .
```

Open <http://localhost:8000> — this is your cloud console. Keep it visible.

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
terraform apply
```

Watch the board appear on the dashboard the moment Terraform creates it.

## API

Everything the provider does is plain JSON over HTTP, so `curl` works too:

| Method & path | Effect |
|---|---|
| `GET/POST /tictactoe/boards`, `GET/DELETE /tictactoe/boards/{id}` | boards (reads include derived `cells`, `next_player`, `winner`, `moves`) |
| `POST /tictactoe/moves`, `GET/DELETE /tictactoe/moves/{id}` | play / inspect / take back a move |
| `POST /tictactoe/nameplates`, `GET/PUT/DELETE /tictactoe/nameplates/{id}` | nameplates (one per board; creating a second 409s and tells you to import) |

State is in-memory: restarting the server wipes the cloud — which is itself
a decent lesson in why real state management matters.
