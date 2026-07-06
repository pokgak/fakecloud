# Chapter 5 — Shared state & stepping on each other

Every workspace has its own state file, and each state file believes it owns
the resources in it. Nothing stops TWO state files from both claiming the
same real object — and that's when coworkers start ruining each other's
afternoons. You're going to stage that fight on purpose: grab a friend (both
players use the SAME sandbox id — that's the point), or be both players
yourself in two terminals.

The weapon is the **nameplate** 🪧 — `fakecloud_nameplate`, a plaque that
hangs on a board (`board_id` + `text`, one per board). It's the only
fakecloud resource that updates **in place**: changing `text` is a yellow
`~` in the plan, not a replace.

## Missions

**1. Player X stakes a claim** (in `player-x/`).
A board plus a nameplate saying something suitably smug. Output the board
and nameplate ids for player O.

**2. Player O muscles in** (in `player-o/`).
Look up X's board with the **data source** (`data.fakecloud_tictactoe_board`
— reading is always safe), then try to *create* a nameplate on it with your
own text. Read the error carefully; fakecloud is unusually helpful — real
clouds either error opaquely or hand you a duplicate.

**3. Do what the error says.**
Import X's nameplate into O's state (chapter 3 skill), then apply. Your
apply flips the text — your first in-place `~` update. Now one plaque lives
in two state files.

**4. The war.**
Player X: run `terraform plan`. Terraform sees YOUR config as truth and O's
text as drift to correct. Apply. Now O's plan wants it back. Alternate a few
rounds and watch the yellow `~` lines ping-pong in the dashboard feed.
Neither player is wrong — the *setup* is wrong: one object, two owners.
(Also try editing the plaque by clicking it on the dashboard: a third actor
both of you will see as drift.)

**5. The truce, and the real fixes.**
In escalating order of correctness:

- a `lifecycle` block with `ignore_changes` on the contested attribute —
  X stops seeing O's edits at all. Fine when the value genuinely doesn't
  matter after creation.
- **one owner**: the resource lives in exactly ONE config; everyone else
  reads through the data source (it exposes `nameplate_text`). Player O:
  delete your resource block and evict it from your state without touching
  the real thing — `terraform state rm` is the tool.
- shared remote state with **locking** (S3+DynamoDB, Terraform Cloud): when
  several people must manage the *same* config they share one state file,
  and locking serializes their applies. fakecloud can't demo backends, but
  the pain you just felt is exactly what they prevent.

## Bonus missions — lifecycle grab-bag

**6. Tripwire.** Give the nameplate `prevent_destroy` and try
`terraform destroy`. Terraform refuses to even plan it.

**7. The create_before_destroy footgun.** Add a mark with a `lifecycle`
block setting `create_before_destroy = true`, apply, then change its
`player` (same cell!) and apply again. CBD builds the replacement *first* —
and the cell is still occupied. Compare with the default order (remove the
lifecycle block): works fine. CBD is great for servers, fatal for anything
with a uniqueness constraint — IAM roles, S3 buckets, DNS names.

## Check yourself

- Mission 2: the create fails with a 409 that names the existing plaque id.
- Mission 4: each player's plan shows `~ text = "<theirs>" -> "<yours>"`.
- Mission 5 truce: X's plan says **No changes** while the plaque still shows
  O's text.
- Mission 7 with CBD: apply fails mid-flight with "position already taken".
