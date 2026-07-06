# Chapter 5, player O — missions are in ../README.md.
# X's board id arrives out of band; take it as a variable.

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

variable "board_id" {
  description = "Player X's board id"
  type        = number
}

# TODO(mission 2): read X's board, then try to claim its nameplate

# TODO(mission 5): the reformed version — read, don't own
