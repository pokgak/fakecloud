# ================================================================
# Chapter 3 — Drift & import
# ================================================================
#
# Terraform only knows about resources in ITS STATE FILE. The real
# world has other actors: coworkers clicking consoles, scripts,
# autoscalers. This chapter uses the fakecloud console as that
# "coworker who clicks things".
#
# Setup: terraform apply, then keep the console open.

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

resource "fakecloud_tictactoe_board" "lesson" {
  name = "chapter-3"
}

# Two managed marks — Terraform's little kingdom.
resource "fakecloud_tictactoe_move" "managed_a" {
  board_id = fakecloud_tictactoe_board.lesson.id
  player   = "X"
  position = 4
}

resource "fakecloud_tictactoe_move" "managed_b" {
  board_id = fakecloud_tictactoe_board.lesson.id
  player   = "X"
  position = 0
}

# --- Exercise 1: unmanaged is not drift ---------------------------
# In the console, CLICK AN EMPTY CELL (place an O somewhere). Then:
#
#   terraform plan
#
# "No changes." Surprised? Terraform doesn't scan your cloud — it
# only refreshes the resources in its state. The O you clicked is
# UNMANAGED infrastructure: invisible to Terraform, and `terraform
# destroy` won't clean it up either. (This is why real teams love
# rules like "everything goes through Terraform".)

# --- Exercise 2: drift --------------------------------------------
# Now CLICK ONE OF THE X MARKS Terraform made (cell 4 or 0) to remove
# it out-of-band. Then:
#
#   terraform plan
#
# THIS Terraform sees: the refresh finds a managed resource missing,
# and the plan offers to recreate it. That's drift detection. Apply
# to heal the board.

# --- Exercise 3: import — adopting the stray ----------------------
# That unmanaged O from exercise 1 can be brought under management.
# The console shows every mark's id. Two ways to adopt it:
#
# (a) The classic CLI way — uncomment the resource below, fill in the
#     attributes to match reality (player/position of your clicked O),
#     then:
#
#       terraform import fakecloud_tictactoe_move.adopted <ID>
#       terraform plan   # must say "No changes" — if not, your block
#                        # doesn't match reality; fix it, don't apply!
#
# (b) The modern declarative way (Terraform 1.5+) — uncomment the
#     import block too, and just run `terraform apply`. Terraform can
#     even write the resource body for you:
#
#       terraform plan -generate-config-out=generated.tf
#
# resource "fakecloud_tictactoe_move" "adopted" {
#   board_id = fakecloud_tictactoe_board.lesson.id
#   player   = "O"
#   position = 8 # <- match where you actually clicked!
# }
#
# import {
#   to = fakecloud_tictactoe_move.adopted
#   id = "7" # <- the move id from the console
# }
#
# After adopting: the O is Terraform's now. `terraform destroy`
# takes it down with everything else. Adoption is forever.

output "cells" {
  value = fakecloud_tictactoe_board.lesson.cells
}
