# Side quest: tic-tac-toe over Terraform — player X.
#
# You create the game and open with the center square. After applying,
# share the game id (see the output, or the dashboard) with player O.
# Each move is a resource: add a fakecloud_tictactoe_move block per turn
# and apply. The server rejects out-of-turn moves and taken cells, so
# a failed apply means "wait for your opponent".
#
# Cells are numbered 0-8, row by row:
#   0 | 1 | 2
#   3 | 4 | 5
#   6 | 7 | 8

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

resource "fakecloud_tictactoe_game" "duel" {
  name = "x-vs-o"
}

resource "fakecloud_tictactoe_move" "opening" {
  game_id  = fakecloud_tictactoe_game.duel.id
  player   = "X"
  position = 4
}

# your second move: uncomment, pick an empty cell, apply — but only
# after O has played, or the server will reject it
#
# resource "fakecloud_tictactoe_move" "second" {
#   game_id  = fakecloud_tictactoe_game.duel.id
#   player   = "X"
#   position = 0
# }

output "game_id" {
  description = "Share this with player O"
  value       = fakecloud_tictactoe_game.duel.id
}
