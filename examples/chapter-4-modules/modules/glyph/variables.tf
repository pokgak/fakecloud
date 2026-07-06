# A module's variables are its API — the only way callers can affect
# what it creates.

variable "name" {
  description = "Board name, shown on the dashboard"
  type        = string
}

variable "player" {
  description = "Which mark to draw the glyph with"
  type        = string
  default     = "X"

  validation {
    condition     = contains(["X", "O"], var.player)
    error_message = "player must be \"X\" or \"O\"."
  }
}

variable "cells" {
  description = "Cells to mark (0-8, row by row)"
  type        = list(number)
}
