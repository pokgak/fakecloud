# ================================================================
# Side quest: the duel — player X
# ================================================================
#
# A board with mode = "duel" is refereed by the server: X starts,
# turns alternate, the board locks once someone wins. Share the
# board id (see the output) with player O — they join with the data
# source since the board isn't in *their* state (see ../player-o).
#
# Play by adding one fakecloud_tictactoe_move block per turn and
# applying. Applying out of turn fails with the referee's reason —
# a failed apply just means "wait for your opponent".
#
# Cells: 0 | 1 | 2
#        3 | 4 | 5
#        6 | 7 | 8

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

resource "fakecloud_tictactoe_board" "duel" {
  name = "x-vs-o"
  mode = "duel"
}

resource "fakecloud_tictactoe_move" "opening" {
  board_id = fakecloud_tictactoe_board.duel.id
  player   = "X"
  position = 4
}

# your second move: uncomment after O has played, pick a cell, apply
#
# resource "fakecloud_tictactoe_move" "second" {
#   board_id = fakecloud_tictactoe_board.duel.id
#   player   = "X"
#   position = 0
# }

output "board_id" {
  description = "Share this with player O"
  value       = fakecloud_tictactoe_board.duel.id
}

output "status" {
  value = (fakecloud_tictactoe_board.duel.winner != ""
    ? "game over: ${fakecloud_tictactoe_board.duel.winner}"
  : "next up: ${fakecloud_tictactoe_board.duel.next_player}")
}
