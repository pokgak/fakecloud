# ================================================================
# Chapter 2 — count, for_each, and the footguns
# ================================================================
#
# One resource block can manage many resources. There are two ways,
# and the difference between them is one of Terraform's classic
# footguns. The board makes it visible.
#
# Keep the console open. You'll watch a board tear itself apart.

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
  name = "chapter-2"
}

variable "x_cells" {
  description = "Cells X paints, in order"
  type        = list(number)
  default     = [0, 4, 8] # the diagonal
}

# ================================================================
# PART 1 — count: instances are numbered
# ================================================================
# count makes N instances addressed by INDEX: x[0], x[1], x[2].
# Each instance's identity is its position in the list.
#
#   terraform apply        -> the diagonal appears on the console
#   terraform state list   -> note the [0] [1] [2] addresses

resource "fakecloud_tictactoe_move" "x" {
  count    = length(var.x_cells)
  board_id = fakecloud_tictactoe_board.lesson.id
  player   = "X"
  position = var.x_cells[count.index]
}

# --- THE FOOTGUN --------------------------------------------------
# Now remove the FIRST element:
#
#   terraform plan -var 'x_cells=[4, 8]'
#
# You wanted: "remove the mark at 0". Terraform plans:
#
#   x[0]: position 0 -> 4  (replace!)
#   x[1]: position 4 -> 8  (replace!)
#   x[2]: destroyed
#
# Every element after the removed one SHIFTED INDEX, so Terraform
# thinks they all changed. Five operations instead of one.
#
# Now apply it and WATCH THE CONSOLE, not the terminal:
#
#   terraform apply -var 'x_cells=[4, 8]'
#
# Marks get torn off cells 4 and 8 and redrawn — cells you never
# meant to touch. The end state is right (Terraform is convergent),
# but the intermediate states were real: for a moment your "cloud"
# didn't have the marks you were relying on. With real infrastructure
# this same pattern is how "remove one server from the list" becomes
# "recreate most of the fleet" — surprise downtime, new IPs, and in
# clouds where names must be unique, sometimes a mid-apply collision
# failure at 2am.
#
# Clean up before part 2:  terraform destroy

# ================================================================
# PART 2 — for_each: instances are named
# ================================================================
# for_each makes instances addressed by KEY: x["0"], x["4"], x["8"].
# Identity follows the key, not the position in a list.
#
# Comment out ALL of the "x" resource above, uncomment this block,
# then repeat the experiment:
#
#   terraform apply
#   terraform apply -var 'x_cells=[4, 8]'
#
# The plan is exactly what you meant: destroy x["0"], touch nothing
# else. One operation. No races, no shifted identities.

# resource "fakecloud_tictactoe_move" "x" {
#   for_each = toset([for c in var.x_cells : tostring(c)])
#   board_id = fakecloud_tictactoe_board.lesson.id
#   player   = "X"
#   position = tonumber(each.key)
# }

# --- More footgun-lore for the road -------------------------------
# * for_each keys must be known at plan time and must be strings —
#   hence the tostring/tonumber dance above.
# * `toset` also deduplicates and drops ordering. If your input has
#   meaningful duplicates, for_each over a MAP instead.
# * count is still right when instances are genuinely interchangeable
#   copies (count = var.replicas), or for the conditional-resource
#   trick: count = var.enabled ? 1 : 0.
# * Migrating existing count state to for_each without destroying
#   anything is what `terraform state mv` (or `moved` blocks) is for.

output "cells" {
  value = fakecloud_tictactoe_board.lesson.cells
}
