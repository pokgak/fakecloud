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
  # Paste your playground id — its dashboard's "Connect Terraform" panel
  # has the exact block to copy. (Or export FAKECLOUD_SANDBOX instead.)
  sandbox = "your-sandbox-id"

  # Running fakecloud locally (cd server && npx wrangler dev)? Add:
  # endpoint = "http://localhost:8787"
}

# TODO(mission 2): call your glyph module

# TODO(mission 3): the gallery — for_each over a map of glyphs

# TODO(mission 4): the moved block
