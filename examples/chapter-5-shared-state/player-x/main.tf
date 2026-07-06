# Chapter 5, player X — missions are in ../README.md.

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

# TODO(mission 1): a board, a smug nameplate, and outputs for O

# TODO(mission 5): the truce
# TODO(mission 6): the tripwire
# TODO(mission 7): the CBD footgun
