# Chapter 2 — your workspace. Missions are in README.md.

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

# TODO(mission 1): a board, an x_cells variable, and ONE move block
#                  that draws a mark per element using count

# TODO(mission 3): the same block, rebuilt with for_each
