# Chapter 5, player O — missions are in ../README.md.
# X's board id arrives out of band; take it as a variable.

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

# TODO(mission 2): read X's board, then try to claim its nameplate

# TODO(mission 5): the reformed version — read, don't own
