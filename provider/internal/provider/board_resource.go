package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pokgak/terraform-provider-fakecloud/internal/client"
)

var (
	_ resource.Resource                = &boardResource{}
	_ resource.ResourceWithConfigure   = &boardResource{}
	_ resource.ResourceWithImportState = &boardResource{}
)

type boardResource struct {
	client *client.Client
}

func NewBoardResource() resource.Resource {
	return &boardResource{}
}

type boardModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Mode       types.String `tfsdk:"mode"`
	Cells      types.List   `tfsdk:"cells"`
	NextPlayer types.String `tfsdk:"next_player"`
	Winner     types.String `tfsdk:"winner"`
}

func (r *boardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tictactoe_board"
}

func (r *boardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A tic-tac-toe board — the one primitive every fakecloud lesson is built on. " +
			"Mark cells by creating fakecloud_tictactoe_move resources; destroying the board clears it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Server-assigned board id.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the board, shown on the dashboard.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("freeplay"),
				Description: `"freeplay" (default): mark any empty cell any time — what the lessons use. ` +
					`"duel": the server referees a real game — X starts, turns alternate, the board locks on a win.`,
				Validators: []validator.String{
					stringvalidator.OneOf("freeplay", "duel"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cells": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Current board, 9 cells row by row; each is \"\", \"X\", or \"O\". Refreshed on plan/apply.",
			},
			"next_player": schema.StringAttribute{
				Computed:    true,
				Description: "In duel mode, whose turn it is; empty in freeplay or once the game is over.",
			},
			"winner": schema.StringAttribute{
				Computed:    true,
				Description: "\"X\", \"O\", \"draw\", or empty while no line is complete.",
			},
		},
	}
}

func (r *boardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func boardToModel(ctx context.Context, board client.Board) (boardModel, error) {
	cells, diags := types.ListValueFrom(ctx, types.StringType, board.Cells)
	if diags.HasError() {
		return boardModel{}, fmt.Errorf("converting cells: %v", diags.Errors())
	}
	return boardModel{
		ID:         types.Int64Value(board.ID),
		Name:       types.StringValue(board.Name),
		Mode:       types.StringValue(board.Mode),
		Cells:      cells,
		NextPlayer: types.StringValue(board.NextPlayer),
		Winner:     types.StringValue(board.Winner),
	}, nil
}

func (r *boardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan boardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	board, err := r.client.CreateBoard(plan.Name.ValueString(), plan.Mode.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create board", err.Error())
		return
	}

	model, err := boardToModel(ctx, board)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert board state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *boardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state boardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	board, err := r.client.GetBoard(state.ID.ValueInt64())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read board", err.Error())
		return
	}

	model, err := boardToModel(ctx, board)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert board state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Update is never called: "name" and "mode" require replacement and
// everything else is computed. It exists only to satisfy the
// resource.Resource interface.
func (r *boardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported operation", "fakecloud_tictactoe_board cannot be updated in place")
}

func (r *boardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state boardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteBoard(state.ID.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete board", err.Error())
	}
}

func (r *boardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("expected a numeric board id, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
