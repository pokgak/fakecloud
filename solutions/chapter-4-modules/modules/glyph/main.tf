# The glyph module: one board plus a pattern of marks on it.
#
# Gotcha worth knowing: a module inherits the provider CONFIGURATION
# (the endpoint) from its caller, but NOT the source mapping. Without
# this required_providers block, Terraform would assume "fakecloud"
# means the default hashicorp/fakecloud — which doesn't exist — and
# fail. Every module that uses a non-HashiCorp provider needs this.
terraform {
  required_providers {
    fakecloud = {
      source  = "pokgak/fakecloud"
      version = "~> 0.3"
    }
  }
}

resource "fakecloud_tictactoe_board" "canvas" {
  name = var.name
}

# Chapter 2's lesson, now working for you inside an abstraction.
resource "fakecloud_tictactoe_move" "pixel" {
  for_each = toset([for c in var.cells : tostring(c)])
  board_id = fakecloud_tictactoe_board.canvas.id
  player   = var.player
  position = tonumber(each.key)
}
