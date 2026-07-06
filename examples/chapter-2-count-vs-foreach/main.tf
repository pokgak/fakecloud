# Chapter 2 — your workspace. Missions are in README.md.

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

# TODO(mission 1): a board, an x_cells variable, and ONE move block
#                  that draws a mark per element using count

# TODO(mission 3): the same block, rebuilt with for_each
