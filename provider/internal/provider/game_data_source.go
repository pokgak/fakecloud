package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pokgak/terraform-provider-fakecloud/internal/client"
)

var (
	_ datasource.DataSource              = &gameDataSource{}
	_ datasource.DataSourceWithConfigure = &gameDataSource{}
)

// gameDataSource lets the player who did NOT create the game (and therefore
// doesn't have it in their state) look up the board by id.
type gameDataSource struct {
	client *client.Client
}

func NewGameDataSource() datasource.DataSource {
	return &gameDataSource{}
}

func (d *gameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tictactoe_game"
}

func (d *gameDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a tic-tac-toe game by id — handy for the opponent who joined a game " +
			"they didn't create, or for printing the board with an output block.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
				Description: "Game id, as shown on the dashboard.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"board": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Current board, 9 cells row by row.",
			},
			"next_player": schema.StringAttribute{
				Computed: true,
			},
			"winner": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *gameDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *gameDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config gameModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	game, err := d.client.GetGame(config.ID.ValueInt64())
	if err != nil {
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
