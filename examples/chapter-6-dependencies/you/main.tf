# Chapter 6, you — missions are in ../README.md.

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

# TODO(mission 2): your board, a nameplate, two marks — references only

# TODO(mission 3): the ceremonial mark that waits for the plaque

# TODO(mission 4): the away game — a move on the opponent's board
