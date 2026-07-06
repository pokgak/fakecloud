# Chapter 3 — Drift & import

Terraform only knows about resources in **its state file**. The real world
has other actors: coworkers clicking consoles, scripts, autoscalers. In this
chapter the fakecloud dashboard plays the coworker who clicks things.

## Missions

**1. Set the stage.**
A board and two X marks, all managed by Terraform (you learned every piece
of this in chapter 1). Apply, and confirm `terraform state list` shows all
three.

**2. Unmanaged is not drift.**
On the dashboard, **click an empty cell** to place an O by hand. Then run
`terraform plan`. Surprised by the answer? Terraform doesn't scan your
cloud — it only refreshes the resources it already tracks. Your O is
*unmanaged infrastructure*: invisible to plan, and `terraform destroy` won't
touch it either. (This is why teams adopt "everything goes through
Terraform" rules.)

**3. Now cause real drift.**
Click one of the X marks **Terraform manages** to remove it out-of-band,
then plan again. This time Terraform notices — the refresh finds a managed
resource missing and offers to recreate it. Apply to heal the board.

**4. Adopt the stray.**
Bring the hand-clicked O under management. The dashboard shows every mark's
id and the exact import command. Write a resource block whose attributes
match reality *exactly* (which cell did you click? which player?), then:

- the classic way: `terraform import <address> <id>`
- the declarative way (1.5+): an `import` block, then just apply — and try
  `terraform plan -generate-config-out=generated.tf` to have Terraform write
  the block for you.

## Check yourself

- Mission 2's plan: **no resource changes** — the stray O appears nowhere.
- Mission 3's plan: **1 to add**.
- After mission 4: `terraform plan` shows **no changes**. If it wants to
  modify or replace your adopted mark, your block doesn't match reality —
  fix the block, never apply a mismatch.
- Adoption is forever: `terraform destroy` now takes the O down with
  everything else.
