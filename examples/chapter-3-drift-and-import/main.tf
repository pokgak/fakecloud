# Chapter 3 — your workspace. Missions are in README.md.

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

# TODO(mission 1): a board + two managed X marks

# TODO(mission 4): the adopted stray
