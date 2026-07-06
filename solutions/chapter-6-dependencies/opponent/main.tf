# ================================================================
# Chapter 6 — Dependencies & ordering — the opponent (run me FIRST)
# ================================================================
#
# This tiny config plays the role of "someone else's infrastructure":
# a board created by a different config, in a different state file.
# The main lesson is in ../you — apply this first, note the board id,
# then go there.

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

resource "fakecloud_tictactoe_board" "theirs" {
  name = "opponents-board"
}

output "board_id" {
  value = fakecloud_tictactoe_board.theirs.id
}
