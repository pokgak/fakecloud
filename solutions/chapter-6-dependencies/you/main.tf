# ================================================================
# Chapter 6 — Dependencies & ordering
# ================================================================
#
# Terraform never runs your file top to bottom. It builds a GRAPH:
# every reference from one block to another is an edge, and apply
# walks the graph — dependencies first, independent things in
# parallel. This chapter makes the graph visible twice over: in
# `terraform graph`, and live on the dashboard as things appear in
# dependency order.
#
# Setup: apply ../opponent first and note its board_id output.

terraform {
  required_providers {
    fakecloud = {
      source  = "pokgak/fakecloud"
      version = "~> 0.3"
    }
  }
}

provider "fakecloud" {
  sandbox = "your-sandbox-id" # from your playground's dashboard

  # Running fakecloud locally (cd server && npx wrangler dev)? Add:
  # endpoint = "http://localhost:8787"
}

# ================================================================
# PART 1 — implicit dependencies: references ARE the ordering
# ================================================================
# Nothing below says "create the board first". It doesn't have to:
# the marks and the plaque REFERENCE the board's id, and that
# reference is the dependency. Apply and watch the console feed —
# the board always appears before anything on it, while the two
# marks (independent of each other) land in whatever order the
# parallel walk gets to them.

resource "fakecloud_tictactoe_board" "mine" {
  name = "chapter-6"
}

resource "fakecloud_nameplate" "plaque" {
  board_id = fakecloud_tictactoe_board.mine.id
  text     = "graphs, not scripts"
}

resource "fakecloud_tictactoe_move" "first" {
  board_id = fakecloud_tictactoe_board.mine.id
  player   = "X"
  position = 4
}

# ================================================================
# PART 2 — explicit dependencies: depends_on
# ================================================================
# Sometimes there's a real ordering requirement but NO natural
# reference to carry it (in real clouds: "the app must not boot
# before the IAM policy attaches"). depends_on adds the edge by hand.
#
# This mark never mentions the plaque, but depends_on forces it to
# wait. Apply from scratch (destroy first) a couple of times and
# check the feed: this mark is ALWAYS after the plaque. Then comment
# depends_on out, destroy, re-apply — now it's a coin flip.

resource "fakecloud_tictactoe_move" "ceremonial" {
  board_id = fakecloud_tictactoe_board.mine.id
  player   = "O"
  position = 0

  depends_on = [fakecloud_nameplate.plaque]
}

# ================================================================
# PART 3 — dependencies on things you DON'T manage
# ================================================================
# Your move below needs the opponent's board id, but that board
# belongs to another config. The data source is the bridge, and it's
# part of the same graph: Terraform must READ the opponent's board
# before it can CREATE your move, because the move references
# data.fakecloud_tictactoe_board.opponent.id.
#
#   terraform apply -var opponent_board_id=<id from ../opponent>

variable "opponent_board_id" {
  description = "The board id from ../opponent's output"
  type        = number
}

data "fakecloud_tictactoe_board" "opponent" {
  id = var.opponent_board_id
}

resource "fakecloud_tictactoe_move" "away_game" {
  board_id = data.fakecloud_tictactoe_board.opponent.id
  player   = "O"
  position = 4
}

output "their_cells" {
  value = data.fakecloud_tictactoe_board.opponent.cells
}

# --- Exercises ----------------------------------------------------
# 1. See the actual graph:
#      terraform graph
#    (paste the output into any Graphviz viewer, or just read the
#    edge list — every `->` is a "must happen after").
# 2. Destroy and re-apply while watching the feed. Destroy runs the
#    SAME graph in reverse: marks and plaque first, board last.
# 3. The limit of the graph: destroy ../opponent's board, then plan
#    here. The data source fails — Terraform orders operations WITHIN
#    one config; it cannot order ACROSS configs. Making "apply
#    opponent before you" happen is your job (scripts, CI pipelines,
#    or tools like Terragrunt / Terraform stacks).
# 4. Circular check: point the opponent's nameplate at YOUR board id
#    while your config references theirs, and imagine both being one
#    config — Terraform would refuse: "Cycle". Graphs must be acyclic;
#    if you ever hit that error, one of the references has to give.
