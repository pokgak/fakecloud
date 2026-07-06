package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pokgak/terraform-provider-fakecloud/internal/client"
)

var (
	_ resource.Resource                = &moveResource{}
	_ resource.ResourceWithConfigure   = &moveResource{}
	_ resource.ResourceWithImportState = &moveResource{}
)

type moveResource struct {
	client *client.Client
}

func NewMoveResource() resource.Resource {
	return &moveResource{}
}

type moveModel struct {
	ID       types.Int64  `tfsdk:"id"`
	BoardID  types.Int64  `tfsdk:"board_id"`
	Player   types.String `tfsdk:"player"`
	Position types.Int64  `tfsdk:"position"`
}

func (r *moveResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tictactoe_move"
}

func (r *moveResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A mark on a board. Creating it plays the move; destroying it takes the move back. " +
			"On a duel board the server also rejects the apply if it is not your turn.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Server-assigned move id.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"board_id": schema.Int64Attribute{
				Required:    true,
				Description: "The board this move belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"player": schema.StringAttribute{
				Required:    true,
				Description: `"X" or "O". On a duel board, X always starts.`,
				Validators: []validator.String{
					stringvalidator.OneOf("X", "O"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"position": schema.Int64Attribute{
				Required:    true,
				Description: "Cell to claim, 0-8, row by row (4 is the center).",
				Validators: []validator.Int64{
					int64validator.Between(0, 8),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *moveResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *moveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan moveModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	move, err := r.client.CreateMove(plan.BoardID.ValueInt64(), plan.Player.ValueString(), plan.Position.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Move rejected", err.Error())
		return
	}

	plan.ID = types.Int64Value(move.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *moveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state moveModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	move, err := r.client.GetMove(state.ID.ValueInt64())
	if err != nil {
		if client.IsNotFound(err) {
			// Move gone — e.g. removed in the dashboard, or the board
			// was deleted. Drift: the next plan will offer to replay it.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read move", err.Error())
		return
	}

	state.BoardID = types.Int64Value(move.BoardID)
	state.Player = types.StringValue(move.Player)
	state.Position = types.Int64Value(move.Position)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is never called: every configurable attribute requires replacement.
func (r *moveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported operation", "fakecloud_tictactoe_move cannot be updated in place")
}

func (r *moveResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state moveModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMove(state.ID.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete move", err.Error())
	}
}

// ImportState lets you adopt a mark made outside Terraform (e.g. by clicking
// a cell on the dashboard): terraform import fakecloud_tictactoe_move.NAME ID
func (r *moveResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("expected a numeric move id, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
