# ================================================================
# Chapter 5 — Shared state & stepping on each other — player X
# ================================================================
#
# Every workspace has its own state file, and each state file believes
# it owns the resources in it. Nothing stops TWO state files from both
# claiming the same real object — and that's when coworkers start
# ruining each other's afternoons. This chapter stages that fight on
# purpose, with a friend or with yourself in two terminals.
#
# The weapon: the nameplate 🪧 — the only fakecloud resource that
# updates IN PLACE (a yellow ~ in the plan, not a replace). Each board
# holds exactly one.
#
# Setup here in player-x/:   terraform apply
# Then go to ../player-o and follow its comments.

terraform {
  required_providers {
    fakecloud = {
      source = "pokgak/fakecloud"
    }
  }
}

provider "fakecloud" {
  endpoint = "http://localhost:8000"
}

resource "fakecloud_tictactoe_board" "shared" {
  name = "chapter-5"
}

resource "fakecloud_nameplate" "plaque" {
  board_id = fakecloud_tictactoe_board.shared.id
  text     = "X was here"

  # --- Exercise 3 (the truce): uncomment and re-apply ---------------
  # ignore_changes tells Terraform "I set this on create, but if it
  # changes later, that's fine — leave it alone". The war ends: O's
  # edits stop showing up in your plans at all.
  #
  # lifecycle {
  #   ignore_changes = [text]
  # }

  # --- Exercise 4: uncomment, then try `terraform destroy` ----------
  # prevent_destroy makes Terraform refuse to plan anything that would
  # delete this resource — a tripwire for your most precious objects.
  # (You'll have to comment it back out to actually clean up.)
  #
  # lifecycle {
  #   prevent_destroy = true
  # }
}

output "board_id" {
  description = "Share with player O"
  value       = fakecloud_tictactoe_board.shared.id
}

output "nameplate_id" {
  description = "Share with player O — they will import this"
  value       = fakecloud_nameplate.plaque.id
}

# --- Exercise 1: the war (after O has imported the plaque) ---------
# O applies "O rules". Now run `terraform plan` here:
#
#   ~ text = "O rules" -> "X was here"
#
# Terraform sees YOUR config as truth and O's change as drift. Apply,
# and now O's next plan wants it back. Alternate applies a few times
# and watch the yellow ~ lines ping-pong in the console feed. Neither
# of you is wrong; the SETUP is wrong: one object, two owners.
#
# --- Exercise 2: spot the third player ----------------------------
# Click the plaque on the dashboard and edit the text by hand. Now
# BOTH your plans show drift. Any number of actors can step on a
# shared object — Terraform can only defend the states it knows about.
#
# --- The real fixes (in escalating order of correctness) ----------
# a) ignore_changes (exercise 3): fine when the value genuinely
#    doesn't matter after creation.
# b) ONE owner: the resource lives in exactly one config; everyone
#    else reads it via the data source. This is the rule real teams
#    live by. (Player O's file shows the reformed version at the end.)
# c) Shared remote state with locking (S3+DynamoDB, Terraform Cloud,
#    etc.): when multiple people must manage the SAME config, they
#    share ONE state file, and locking stops concurrent applies from
#    corrupting it. fakecloud can't demo backends, but the pain you
#    just felt is exactly what they prevent.

# ================================================================
# Bonus: the create_before_destroy footgun
# ================================================================
# CBD flips replacement order: build the new thing first, then delete
# the old one. Great for servers (no downtime) — FATAL for anything
# with a uniqueness constraint.
#
# Uncomment the mark below and apply. Then change player "X" -> "O"
# (same cell!) and apply again:
#
#   * without CBD: destroy old mark, create new one. Works.
#   * with CBD:    create the new mark FIRST -> the cell is still
#                  occupied -> "position 8 is already taken". Apply
#                  fails.
#
# Real-world versions of this: IAM roles, S3 buckets, DNS records —
# anything where the new copy collides with the old one's name.
#
# resource "fakecloud_tictactoe_move" "cbd_demo" {
#   board_id = fakecloud_tictactoe_board.shared.id
#   player   = "X"
#   position = 8
#
#   lifecycle {
#     create_before_destroy = true
#   }
# }
