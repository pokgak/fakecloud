# ================================================================
# Side quest: the duel — player O
# ================================================================
#
# You didn't create the board, so it isn't in your state; look it up
# with the DATA SOURCE using the id player X shared with you. Then
# add one fakecloud_tictactoe_move block per turn and apply.
#
# Bonus lesson: `terraform plan` re-reads the data source, so it
# doubles as a "has X moved yet?" check — watch the board and status
# outputs change.

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

variable "board_id" {
  description = "The board id player X shared with you"
  type        = number
  default     = 1
}

data "fakecloud_tictactoe_board" "duel" {
  id = var.board_id
}

resource "fakecloud_tictactoe_move" "first" {
  board_id = data.fakecloud_tictactoe_board.duel.id
  player   = "O"
  position = 0
}

output "cells" {
  value = data.fakecloud_tictactoe_board.duel.cells
}

output "status" {
  value = (data.fakecloud_tictactoe_board.duel.winner != ""
    ? "game over: ${data.fakecloud_tictactoe_board.duel.winner}"
  : "next up: ${data.fakecloud_tictactoe_board.duel.next_player}")
}
