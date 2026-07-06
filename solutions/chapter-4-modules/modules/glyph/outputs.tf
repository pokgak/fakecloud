# Outputs are the only values a module exposes back to its caller.
# Everything else inside is encapsulated — the caller can't reach
# module.NAME.fakecloud_tictactoe_board.canvas directly.

output "board_id" {
  description = "The board this glyph was drawn on"
  value       = fakecloud_tictactoe_board.canvas.id
}

output "cells" {
  description = "The board's current cells"
  value       = fakecloud_tictactoe_board.canvas.cells
}
