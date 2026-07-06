package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pokgak/terraform-provider-fakecloud/internal/client"
)

// DefaultEndpoint is the hosted fakecloud. Point the provider elsewhere
// (e.g. a local `wrangler dev` on http://localhost:8787) with the endpoint
// attribute or FAKECLOUD_ENDPOINT.
const DefaultEndpoint = "https://fakecloud.pokgak.workers.dev"

var _ provider.Provider = &fakecloudProvider{}

type fakecloudProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &fakecloudProvider{version: version}
	}
}

type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Sandbox  types.String `tfsdk:"sandbox"`
}

func (p *fakecloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fakecloud"
	resp.Version = p.version
}

func (p *fakecloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage resources in fakecloud, a pretend cloud for learning Terraform. " +
			"Create a playground on the fakecloud website, keep its dashboard open, and watch every apply happen live.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the fakecloud server. Defaults to the FAKECLOUD_ENDPOINT " +
					"environment variable, then the hosted fakecloud (" + DefaultEndpoint + "). " +
					"Running locally? `wrangler dev` serves http://localhost:8787.",
			},
			"sandbox": schema.StringAttribute{
				Optional: true,
				Description: "Your playground id, shown on its dashboard. Falls back to the " +
					"FAKECLOUD_SANDBOX environment variable. Every learner (and each duel opponent " +
					"pair) works inside one sandbox; the id is the key, so share it only on purpose.",
			},
		},
	}
}

func (p *fakecloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := config.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("FAKECLOUD_ENDPOINT")
	}
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	endpoint = strings.TrimRight(endpoint, "/")

	sandbox := config.Sandbox.ValueString()
	if sandbox == "" {
		sandbox = os.Getenv("FAKECLOUD_SANDBOX")
	}
	if sandbox != "" {
		endpoint = endpoint + "/s/" + sandbox
	}

	c := client.New(endpoint)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *fakecloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBoardResource,
		NewMoveResource,
		NewNameplateResource,
	}
}

func (p *fakecloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewBoardDataSource,
	}
}
