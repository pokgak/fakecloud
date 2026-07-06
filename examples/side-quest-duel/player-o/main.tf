# The duel, player O — see ../README.md.

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
  description = "The duel board id player X shared with you"
  type        = number
}

# TODO: read the board, make your move
