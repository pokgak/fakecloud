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
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pokgak/terraform-provider-fakecloud/internal/client"
)

var (
	_ resource.Resource                = &nameplateResource{}
	_ resource.ResourceWithConfigure   = &nameplateResource{}
	_ resource.ResourceWithImportState = &nameplateResource{}
)

// nameplateResource is the only fakecloud resource that supports in-place
// updates — changing text is a yellow ~ in the plan, not a replace. That
// also makes it the demo object for two state files fighting over one
// resource (chapter 5).
type nameplateResource struct {
	client *client.Client
}

func NewNameplateResource() resource.Resource {
	return &nameplateResource{}
}

type nameplateModel struct {
	ID      types.Int64  `tfsdk:"id"`
	BoardID types.Int64  `tfsdk:"board_id"`
	Text    types.String `tfsdk:"text"`
}

func (r *nameplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameplate"
}

func (r *nameplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A plaque attached to a board, shown on the dashboard. Each board holds at most one; " +
			"if it already exists, import it instead of creating another. text updates in place.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Server-assigned nameplate id.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"board_id": schema.Int64Attribute{
				Required:    true,
				Description: "The board this plaque hangs on.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"text": schema.StringAttribute{
				Required:    true,
				Description: "What the plaque says. Changing it is an in-place update.",
			},
		},
	}
}

func (r *nameplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *nameplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nameplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plate, err := r.client.CreateNameplate(plan.BoardID.ValueInt64(), plan.Text.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create nameplate", err.Error())
		return
	}

	plan.ID = types.Int64Value(plate.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *nameplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nameplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plate, err := r.client.GetNameplate(state.ID.ValueInt64())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read nameplate", err.Error())
		return
	}

	state.BoardID = types.Int64Value(plate.BoardID)
	state.Text = types.StringValue(plate.Text)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *nameplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan nameplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateNameplate(plan.ID.ValueInt64(), plan.Text.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update nameplate", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *nameplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nameplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteNameplate(state.ID.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete nameplate", err.Error())
	}
}

// ImportState is how a second state file adopts an existing nameplate —
// which is exactly the setup for the chapter 5 apply war.
func (r *nameplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("expected a numeric nameplate id, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
