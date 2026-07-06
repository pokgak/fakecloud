# Chapter 6 — Dependencies & ordering

Terraform never runs your file top to bottom. It builds a **graph**: every
reference from one block to another is an edge, and apply walks the graph —
dependencies first, independent things in parallel. You'll see the graph
twice: in `terraform graph`, and live on the dashboard as resources appear
in dependency order.

Two configs here: `opponent/` plays the role of someone else's
infrastructure; `you/` is where the lessons happen.

## Missions

**1. Someone else's board** (in `opponent/`).
Just a board with a name, plus an output of its id. Apply it and note the
id — this config's state now owns that board, and yours never will.

**2. Implicit edges** (in `you/`).
Your own board, a nameplate on it, and two marks — with every `board_id`
written as a *reference*, never a literal number. Apply and watch the
dashboard feed: the board always lands before the things on it, while the
independent siblings race in parallel. Nothing in your file said "board
first"; the references did.

**3. An edge with no reference.**
Add one more mark that must be created strictly *after* the nameplate —
but a mark has no reason to mention a nameplate, so there's no reference to
carry the ordering. That's what `depends_on` is for. Destroy and re-apply a
couple of times checking the feed order; then remove `depends_on` and watch
the order become a coin flip. (Real-world version: "the app must not boot
before the IAM policy attaches".)

**4. Depending on things you don't manage.**
Take the opponent's board id as a variable, look the board up with the data
source, and place a move on it — `board_id` referencing
`data.fakecloud_tictactoe_board.opponent.id`. The data source is part of the
same graph: Terraform must *read* their board before it can *create* your
move.

**5. See the actual graph.**
`terraform graph` — paste into a Graphviz viewer or just read the edge
list; every `->` is a "must happen after". Find your `depends_on` edge and
the data-source edge.

**6. Find the limit.**
Destroy `opponent/`, then plan in `you/`. The data source fails — Terraform
orders operations *within* one config, but it cannot order applies *across*
configs. Making "opponent before you" happen is your job: scripts, CI
pipelines, or tools like Terragrunt and Terraform stacks.

## Check yourself

- Mission 2: in the feed, the board's `+` is always above its marks'.
- Mission 3 with `depends_on`: your ceremonial mark is ALWAYS after the
  plaque; without it, sometimes not.
- Mission 5: `terraform graph` shows an edge from your ceremonial mark to
  the nameplate, and from your away-game move to the data source.
- Mission 6: the plan fails with a 404 from the data source.

## Think about it

If both configs were merged into one and the opponent's nameplate referenced
YOUR board while your move referenced THEIRS, could Terraform apply it?
(Graphs must be acyclic — the error is literally called `Cycle`. When you
hit one, one of the references has to give.)
