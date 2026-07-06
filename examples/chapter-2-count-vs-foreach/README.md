# Chapter 2 — count, for_each, and the footguns

One resource block can manage many resources. There are two ways, and the
difference is a classic Terraform footgun. The board makes it visible: you
are going to watch marks get torn off cells you never meant to touch.

## Missions

**1. Paint the diagonal with `count`.**
Create a board, then a `list(number)` variable defaulting to `[0, 4, 8]`,
then ONE move block that creates a mark per list element. Hints: `count`,
`count.index`, and `length()`. Apply, then look at `terraform state list` —
note the `[0]`, `[1]`, `[2]` addresses. Each mark's *identity* is its
position in the list.

**2. Fire the footgun.**
You want to remove just the mark at cell 0. So remove the *first* element:

```
terraform plan -var 'x_cells=[4, 8]'
```

Read the plan carefully before applying. You asked for one removal — how
many operations did Terraform plan, and why? Then apply it **watching the
dashboard, not the terminal**. The end state is correct, but the
intermediate states were real. With real infrastructure this is how "remove
one server from the list" becomes "recreate most of the fleet".

**3. Rewrite with `for_each`.**
`terraform destroy`, then convert the block so each mark's identity is its
*cell* rather than its list position. Hints: `for_each` wants a set of
strings, so you'll need `toset()`, a `for` expression with `tostring()`, and
`tonumber(each.key)` on the way back out. Check `state list` again — the
addresses changed shape.

**4. Fire again.**
Same experiment: drop the first element and plan. This time the plan should
be exactly one operation.

## Check yourself

- Mission 2's plan: **2 to add, 3 to destroy** (for a one-element removal!).
- Mission 4's plan: **0 to add, 0 to change, 1 to destroy**.

## Footgun lore for the road

- `for_each` keys must be strings known at plan time.
- `toset()` deduplicates and drops order — if duplicates matter, `for_each`
  over a map instead.
- `count` is still right for genuinely interchangeable copies
  (`count = var.replicas`) and the conditional trick
  (`count = var.enabled ? 1 : 0`).
- Migrating live `count` state to `for_each` without rebuilding is what
  `terraform state mv` and `moved` blocks are for — chapter 4 uses one.
