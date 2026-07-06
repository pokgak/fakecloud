package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pokgak/terraform-provider-fakecloud/internal/client"
)

var (
	_ datasource.DataSource              = &boardDataSource{}
	_ datasource.DataSourceWithConfigure = &boardDataSource{}
)

// boardDataSource lets a player who did NOT create the board (and therefore
// doesn't have it in their state) look it up by id.
type boardDataSource struct {
	client *client.Client
}

func NewBoardDataSource() datasource.DataSource {
	return &boardDataSource{}
}

func (d *boardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tictactoe_board"
}

func (d *boardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a board by id — for the opponent who joined a duel they didn't create, " +
			"or for printing the board with an output block.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
				Description: "Board id, as shown on the dashboard.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"mode": schema.StringAttribute{
				Computed: true,
			},
			"cells": schema.ListAttribute{
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
			"nameplate_text": schema.StringAttribute{
				Computed:    true,
				Description: "Text of the board's nameplate, if one exists.",
			},
		},
	}
}

func (d *boardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *boardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config boardModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	board, err := d.client.GetBoard(config.ID.ValueInt64())
	if err != nil {
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
