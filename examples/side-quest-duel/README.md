# ★ Side quest: the duel

Everything so far used `freeplay` boards. A board with **`mode = "duel"`**
is refereed by the server: X always starts, turns alternate, and the board
locks once someone completes a line. Applying out of turn *fails with the
referee's reason* — a failed apply just means "wait for your opponent".

Grab a friend pointing at the same fakecloud (or two terminals), keep the
dashboard open, and play one `terraform apply` at a time.

Cells: `0 1 2` / `3 4 5` / `6 7 8`.

## Player X (in `player-x/`)

You create the duel board (don't forget the `mode`!) and open with a move.
Output the board id and hand it to your opponent. Each turn after that: one
more move block, apply. Want to see the state of play from the CLI? The
board resource's `cells`, `next_player`, and `winner` attributes make good
outputs.

## Player O (in `player-o/`)

The board isn't in your state and you must NOT create it — chapter 5 taught
you what co-owning ends in. Take the id as a variable, read the board with
the data source, and hang your moves off
`data.fakecloud_tictactoe_board.duel.id`. Bonus: `terraform plan` re-reads
the data source, so it doubles as a "has X moved yet?" check.

## House rules

- One move block per turn. Trying two in one apply gets the second rejected.
- `terraform destroy -target` on a move takes it back (the server allows it
  in any order — house-rule it as "undo your last move only" if you're
  honorable).
- Rematch: whoever owns the board destroys it; moves everywhere become
  drift and vanish from both states on the next refresh.
