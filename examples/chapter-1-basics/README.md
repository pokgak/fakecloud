# Chapter 1 — Terraform basics

Run the fakecloud server (`cd server && go run .`) and open
<http://localhost:8000> next to your terminal. Everything you do here changes
real state you can watch there.

You'll write all the Terraform yourself — this course gives you hints, not
answers. (An answer key lives on the `solutions` branch if you're truly
stuck, but the marks you draw yourself stick better.)

## Missions

**1. Point Terraform at fakecloud.**
`main.tf` already declares the provider's *source*. Add a `provider`
block that configures it — the only setting is `endpoint`, and the server
prints the URL when it starts.

**2. Create a board.**
The resource type is `fakecloud_tictactoe_board`, and it needs one argument:
a `name`. Run `terraform plan` first and read what it intends before you
`terraform apply`. Watch the dashboard as you do.

**3. Put an X in the center.**
The resource type is `fakecloud_tictactoe_move` and it takes a `board_id`, a
`player`, and a `position` (cells are numbered 0–8, row by row). Don't
hardcode the board id — *reference* the board resource's `id` attribute so
Terraform knows the move depends on it.

**4. Parameterize.**
Add a `variable` for O's position with a `default`, use it in a second move,
and override it once with `terraform apply -var ...`.

**5. Report back.**
Add `output` blocks for the board's `id` and its `cells` attribute — `cells`
is *computed*: the server derives it, Terraform just reads it.

## Check yourself

- The dashboard shows one board with two marks, and the activity feed logged
  each create as a green `+`.
- `terraform apply` a second time says **No changes** — Terraform converges,
  it doesn't re-run.
- `terraform state list` shows exactly three resources. That file is
  Terraform's memory; chapter 3 abuses it.

## Think about it

Change your X's `position` and plan again. Why does Terraform say
**must be replaced** instead of updating in place? (Hint: can a mark slide
across a board, or can it only be taken back and played again?)

Finish with `terraform destroy` and watch everything vanish in reverse
dependency order.
