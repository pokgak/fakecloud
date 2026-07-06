# Side quest: tic-tac-toe over Terraform — player O.
#
# You didn't create the game, so it isn't in your state; look it up with
# the data source instead using the id player X shared with you. Then add
# one fakecloud_tictactoe_move block per turn and apply.
#
# `terraform plan` re-reads the data source, so it doubles as a
# "has X moved yet?" check — the board and next_player outputs update.

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

variable "game_id" {
  description = "The game id player X shared with you"
  type        = number
  default     = 1
}

data "fakecloud_tictactoe_game" "duel" {
  id = var.game_id
}

resource "fakecloud_tictactoe_move" "first" {
  game_id  = data.fakecloud_tictactoe_game.duel.id
  player   = "O"
  position = 0
}

output "board" {
  value = data.fakecloud_tictactoe_game.duel.board
}

output "status" {
  value = (data.fakecloud_tictactoe_game.duel.winner != ""
    ? "game over: ${data.fakecloud_tictactoe_game.duel.winner}"
  : "next up: ${data.fakecloud_tictactoe_game.duel.next_player}")
}
