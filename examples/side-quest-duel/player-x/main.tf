# The duel, player X — see ../README.md.

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

# TODO: the duel board, your opening move, and outputs for your opponent
