# Chapter 5, player X — missions are in ../README.md.

terraform {
  required_providers {
    fakecloud = {
      source  = "pokgak/fakecloud"
      version = "~> 0.3"
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

# TODO(mission 1): a board, a smug nameplate, and outputs for O

# TODO(mission 5): the truce
# TODO(mission 6): the tripwire
# TODO(mission 7): the CBD footgun
