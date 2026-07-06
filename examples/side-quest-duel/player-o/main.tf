# The duel, player O — see ../README.md.

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

variable "board_id" {
  description = "The duel board id player X shared with you"
  type        = number
}

# TODO: read the board, make your move
