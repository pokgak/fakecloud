# Chapter 6, the opponent (run me FIRST) — missions in ../README.md.

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

# TODO(mission 1): their board + an output of its id
