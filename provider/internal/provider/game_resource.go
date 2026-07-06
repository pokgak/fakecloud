package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pokgak/terraform-provider-fakecloud/internal/client"
)

var (
	_ resource.Resource                = &gameResource{}
	_ resource.ResourceWithConfigure   = &gameResource{}
	_ resource.ResourceWithImportState = &gameResource{}
)

type gameResource struct {
	client *client.Client
}

func NewGameResource() resource.Resource {
	return &gameResource{}
}

type gameModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Board      types.List   `tfsdk:"board"`
	NextPlayer types.String `tfsdk:"next_player"`
	Winner     types.String `tfsdk:"winner"`
}

func (r *gameResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tictactoe_game"
}

func (r *gameResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A tic-tac-toe board. Share its id with an opponent pointing at the same fakecloud " +
			"and play by applying fakecloud_tictactoe_move resources. Destroying the game clears the board.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Server-assigned game id — share this with your opponent.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the game, shown on the dashboard.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"board": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Current board, 9 cells row by row; each is \"\", \"X\", or \"O\". Refreshed on plan/apply.",
			},
			"next_player": schema.StringAttribute{
				Computed:    true,
				Description: "Whose turn it is (empty once the game is over).",
			},
			"winner": schema.StringAttribute{
				Computed:    true,
				Description: "\"X\", \"O\", \"draw\", or empty while the game is in progress.",
			},
		},
	}
}

func (r *gameResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func gameToModel(ctx context.Context, game client.Game) (gameModel, error) {
	board, diags := types.ListValueFrom(ctx, types.StringType, game.Board)
	if diags.HasError() {
		return gameModel{}, fmt.Errorf("converting board: %v", diags.Errors())
	}
	return gameModel{
		ID:         types.Int64Value(game.ID),
		Name:       types.StringValue(game.Name),
		Board:      board,
		NextPlayer: types.StringValue(game.NextPlayer),
		Winner:     types.StringValue(game.Winner),
	}, nil
}

func (r *gameResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan gameModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	game, err := r.client.CreateGame(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create game", err.Error())
		return
	}

	model, err := gameToModel(ctx, game)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert game state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *gameResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gameModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	game, err := r.client.GetGame(state.ID.ValueInt64())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read game", err.Error())
		return
	}

	model, err := gameToModel(ctx, game)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert game state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Update is never called: "name" requires replacement and everything else is
// computed. It exists only to satisfy the resource.Resource interface.
func (r *gameResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported operation", "fakecloud_tictactoe_game cannot be updated in place")
}

func (r *gameResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gameModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGame(state.ID.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete game", err.Error())
	}
}

func (r *gameResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("expected a numeric game id, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
