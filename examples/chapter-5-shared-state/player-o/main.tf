# ================================================================
# Chapter 5 — Shared state & stepping on each other — player O
# ================================================================
#
# Player X already created the board and its nameplate. You're going
# to claim that SAME nameplate in your own state file — the classic
# mistake this chapter is about.
#
# Step 1: try to create it anyway:
#
#   terraform apply -var board_id=<X's board_id>
#
# The server refuses: "board already has a nameplate (id=N) — import
# it instead of creating another". Real clouds aren't usually this
# helpful; they'd either error opaquely or happily give you a
# duplicate.
#
# Step 2: do what the error says (chapter 3 skills):
#
#   terraform import -var board_id=<ID> fakecloud_nameplate.plaque <nameplate_id>
#   terraform apply  -var board_id=<ID>
#
# The apply flips the text to "O rules" — an in-place ~ update, your
# first! Now two state files own one plaque. Go back to player-x/ and
# run the war (exercise 1 there).

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

variable "board_id" {
  description = "Player X's board id"
  type        = number
}

# Reading someone else's board: data source, not resource. This part
# is CORRECT sharing — you can read all day without owning anything.
data "fakecloud_tictactoe_board" "shared" {
  id = var.board_id
}

resource "fakecloud_nameplate" "plaque" {
  board_id = data.fakecloud_tictactoe_board.shared.id
  text     = "O rules"
}

output "plaque_says" {
  value = fakecloud_nameplate.plaque.text
}

# --- The reformed version ------------------------------------------
# After the war, make peace properly: delete the resource block above,
# remove it from YOUR state without touching the real object:
#
#   terraform state rm fakecloud_nameplate.plaque
#
# ...and read the plaque through the board data source instead:
#
# output "plaque" {
#   value = data.fakecloud_tictactoe_board.shared.nameplate_text
# }
#
# One owner (X), any number of readers. That's the rule.
