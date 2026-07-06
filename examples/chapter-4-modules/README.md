# Chapter 4 — Modules

By now you've written the same blocks a few times: a board, some marks.
Modules bundle a pattern up so you can stamp it out with one block — and
this time you're going to *write* the module, not just call one.

## Missions

**1. Build a `glyph` module.**
Create `modules/glyph/` with the classic three files:

- `variables.tf` — the module's inputs: a board `name`, a `player`
  (add a `validation` block that only allows X or O), and a list of `cells`.
- `main.tf` — one board plus its marks. Inside a module you write ordinary
  Terraform; chapter 2's `for_each` technique is exactly what the marks
  want. **One gotcha**: a module inherits the provider *configuration* from
  its caller, but not the *source mapping* — without its own
  `required_providers` block naming `pokgak/fakecloud`, Terraform assumes
  `hashicorp/fakecloud` and fails.
- `outputs.tf` — expose the board's `id` (and `cells` if you like).
  Outputs are the ONLY thing a caller can see.

**2. Call it.**
In this directory's `main.tf`, a single `module` block with
`source = "./modules/glyph"`. A smiley is `cells = [0, 2, 6, 7, 8]`. Local
modules need installing — `terraform init` does that too (notice its
"Initializing modules..." line). Then apply: one block, six resources.
Look at their addresses in `terraform state list`.

**3. A gallery.**
`for_each` works on module blocks too. Make a `map(list(number))` variable
of glyphs — try `cross = [0, 2, 4, 6, 8]`, `frame = [0, 1, 2, 3, 5, 6, 7, 8]`,
`arrow = [1, 3, 4, 5, 7]` — and stamp them all from one module block. Watch
the dashboard when you apply; this is the moment modules click. Reference an
individual instance's output with `module.<name>["<key>"]`.

**4. The rename footgun.**
Rename your mission-2 module block (and its references), run
`terraform init` again (module installs are per-block-name), and plan. Read it: the
module name is part of every resource's address, so Terraform wants to
destroy and recreate all of them. The cure is a `moved` block —
`from`/`to` at the module address level. Add one and re-plan.

## Check yourself

- Mission 3: one apply, and the whole gallery appears on the dashboard.
- Mission 4 before the `moved` block: **6 to add, 6 to destroy**.
  After: **0 to add, 0 to change, 0 to destroy** plus "moved" notes.
- `terraform apply -target='module.gallery["cross"]'` works at module
  granularity — for emergencies, not workflows.
