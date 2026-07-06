# The duel, player X — see ../README.md.

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

# TODO: the duel board, your opening move, and outputs for your opponent
