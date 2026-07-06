# Chapter 1 — Terraform basics

Get yourself a playground: open the hosted fakecloud (see the repo README),
sign in with GitHub, click **+ new playground**, and keep its dashboard open
next to your terminal. Everything you do here changes real state you can
watch there. (Prefer offline? `cd server && npm install && npx wrangler dev`
runs the identical thing at <http://localhost:8787>, no sign-in needed.)

You'll write all the Terraform yourself — this course gives you hints, not
answers. (An answer key lives on the `solutions` branch if you're truly
stuck, but the marks you draw yourself stick better.)

## Missions

**1. Install the provider.**
`main.tf` already declares which provider this configuration needs — run
`terraform init` and watch Terraform download `pokgak/fakecloud` from the
Terraform Registry. Two things appeared: a `.terraform/` directory (the
provider binary lives there; never commit it) and `.terraform.lock.hcl`
(pinned checksums so your teammates get the *same* provider — do commit
that one).

**2. Point Terraform at your playground.**
Add a `provider` block: a `sandbox` (your playground id — the dashboard's
"Connect Terraform" panel shows it), and an `endpoint` only if you're not
using the hosted default (locally that's `http://localhost:8787`).

**3. Create a board.**
The resource type is `fakecloud_tictactoe_board`, and it needs one argument:
a `name`. Run `terraform plan` first and read what it intends before you
`terraform apply`. Watch the dashboard as you do.

**4. Put an X in the center.**
The resource type is `fakecloud_tictactoe_move` and it takes a `board_id`, a
`player`, and a `position` (cells are numbered 0–8, row by row). Don't
hardcode the board id — *reference* the board resource's `id` attribute so
Terraform knows the move depends on it.

**5. Parameterize.**
Add a `variable` for O's position with a `default`, use it in a second move,
and override it once with `terraform apply -var ...`.

**6. Report back.**
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
