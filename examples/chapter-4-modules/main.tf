# ================================================================
# Chapter 4 — Modules
# ================================================================
#
# By now you've written the same three blocks a few times: a board,
# some marks, some outputs. Modules bundle that pattern up so you can
# stamp it out with one block. Look inside modules/glyph/ — it's just
# the Terraform you already know, with variables as its inputs and
# outputs as its return values.

terraform {
  required_providers {
    fakecloud = {
      source = "pokgak/fakecloud"
    }
  }
}

provider "fakecloud" {
  endpoint = "http://localhost:8000"
}

# ================================================================
# PART 1 — one module call
# ================================================================
# One extra step for module configs: run `terraform get` first to
# install local modules. (Normally `terraform init` does this, but
# init also wants to download providers from the registry, which our
# dev-override provider isn't in. `terraform get` installs modules
# only.)
#
# Then `terraform apply` — a board named "smiley" appears with the
# pattern already drawn. One block, six resources.
#
# Then look at `terraform state list`: the module's resources live at
# addresses like module.smiley.fakecloud_tictactoe_move.pixel["0"] —
# the module name is part of the resource's identity. (Remember that;
# it's the footgun in exercise 3.)

module "smiley" {
  source = "./modules/glyph"
  name   = "smiley"
  player = "O"
  cells  = [0, 2, 6, 7, 8] # eyes + mouth
}

# You consume a module through its outputs:
output "smiley_board" {
  value = module.smiley.board_id
}

# ================================================================
# PART 2 — a gallery: for_each over the module itself
# ================================================================
# Modules compose with chapter 2: for_each works on module blocks too.
# Uncomment, apply, and watch a whole gallery of boards appear in the
# console from one block. THIS is the moment modules click — reuse at
# a scale you can see.

# variable "glyphs" {
#   type = map(list(number))
#   default = {
#     cross  = [0, 2, 4, 6, 8]
#     frame  = [0, 1, 2, 3, 5, 6, 7, 8]
#     arrow  = [1, 3, 4, 5, 7]
#   }
# }
#
# module "gallery" {
#   source   = "./modules/glyph"
#   for_each = var.glyphs
#   name     = each.key
#   cells    = each.value
# }
#
# output "gallery_boards" {
#   value = { for name, glyph in module.gallery : name => glyph.board_id }
# }

# --- Exercises ----------------------------------------------------
# 1. Apply part 1, read `terraform state list`, and find the pixel
#    marks' full addresses.
# 2. Uncomment part 2 and apply. Now change the arrow's cells and
#    apply again — only that one board changes (for_each keys again).
# 3. THE FOOTGUN: rename `module "smiley"` to `module "smile"` (and
#    its references), run `terraform get` again (module installs are
#    per-block-name), then plan. Terraform wants to DESTROY six
#    resources and create
#    six "new" ones — the address changed, so as far as state is
#    concerned these are different resources. The fix is a moved
#    block, which migrates state instead of rebuilding the world:
#
#      moved {
#        from = module.smiley
#        to   = module.smile
#      }
#
#    Add it, re-plan: "0 to add, 0 to destroy" plus a note that
#    resources were moved. Same trick works when renaming plain
#    resources or switching count -> for_each.
# 4. `terraform apply -target=module.gallery["cross"]` — surgical
#    applies work at module granularity too (use sparingly; -target
#    is for emergencies, not workflows).
