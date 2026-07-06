# ================================================================
# Chapter 1 — Terraform basics
# ================================================================
#
# Before you start: run the fakecloud server (`cd server && go run .`)
# and open http://localhost:8000 next to your terminal. Every command
# you run here changes real state you can see there.
#
# The whole course uses one primitive: a tic-tac-toe board, where every
# mark on it is a resource.

# --- The terraform block -----------------------------------------
# Declares which providers this configuration needs. A provider is a
# plugin that teaches Terraform how to talk to some API — AWS, GitHub,
# or in our case, fakecloud.
terraform {
  required_providers {
    fakecloud = {
      source = "pokgak/fakecloud"
    }
  }
}

# --- The provider block ------------------------------------------
# Configures that plugin. Here: where the fakecloud server lives.
provider "fakecloud" {
  endpoint = "http://localhost:8000"
}

# --- Your first resource -----------------------------------------
# A resource block says "I want one of these to exist". The two labels
# are the TYPE (fakecloud_tictactoe_board) and a NAME you choose
# ("lesson") that other blocks use to refer to it.
#
# Run `terraform apply` and watch the board appear in the console.
resource "fakecloud_tictactoe_board" "lesson" {
  name = "chapter-1"
}

# --- References --------------------------------------------------
# board_id isn't hardcoded — it refers to the board above. Terraform
# sees the dependency and knows to create the board first. This graph
# of references is how Terraform orders everything it does.
resource "fakecloud_tictactoe_move" "center" {
  board_id = fakecloud_tictactoe_board.lesson.id
  player   = "X"
  position = 4 # cells are numbered 0-8, row by row; 4 is the center
}

# --- Variables ---------------------------------------------------
# Inputs to your configuration. Override at apply time:
#   terraform apply -var corner=2
variable "corner" {
  description = "Which corner O takes (0, 2, 6, or 8)"
  type        = number
  default     = 0
}

resource "fakecloud_tictactoe_move" "corner" {
  board_id = fakecloud_tictactoe_board.lesson.id
  player   = "O"
  position = var.corner
}

# --- Outputs ------------------------------------------------------
# Values printed after apply (and readable with `terraform output`).
# `cells` is a *computed* attribute: the server derives it from the
# moves, Terraform just reads it back.
output "board_id" {
  value = fakecloud_tictactoe_board.lesson.id
}

output "cells" {
  value = fakecloud_tictactoe_board.lesson.cells
}

# --- Exercises ----------------------------------------------------
# 1. `terraform plan`  — read it: 3 to add. Nothing happened yet;
#    plan is always a dry run.
# 2. `terraform apply` — watch the console: board + two marks appear,
#    and the state activity feed logs each one.
# 3. Change `position = 4` to another empty cell and apply. Read the
#    plan first: why does it say "must be replaced" instead of
#    "update in place"? (A mark can't slide; it can only be taken
#    back and played again. The provider declares position as
#    "requires replacement".)
# 4. Run `terraform apply` again with no changes: "No changes." —
#    Terraform is idempotent; it converges on the config, it doesn't
#    re-run scripts.
# 5. `terraform state list` — this is Terraform's memory of what it
#    manages. We'll abuse it in chapter 3.
# 6. `terraform destroy` — everything you made disappears from the
#    console, in reverse dependency order.
