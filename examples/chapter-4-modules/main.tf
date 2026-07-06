# Chapter 4 — your workspace. Missions are in README.md.
# The module itself goes in modules/glyph/ — you're writing it.

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

# TODO(mission 2): call your glyph module

# TODO(mission 3): the gallery — for_each over a map of glyphs

# TODO(mission 4): the moved block
