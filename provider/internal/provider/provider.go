package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pokgak/terraform-provider-fakecloud/internal/client"
)

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
}

func (p *fakecloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fakecloud"
	resp.Version = p.version
}

func (p *fakecloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage resources in fakecloud, a pretend cloud for learning Terraform. " +
			"Open the fakecloud dashboard in a browser to watch every apply happen live.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the fakecloud server. Defaults to the FAKECLOUD_ENDPOINT " +
					"environment variable, then http://localhost:8000.",
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
		endpoint = "http://localhost:8000"
	}

	c := client.New(endpoint)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *fakecloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVMResource,
		NewGameResource,
		NewMoveResource,
	}
}

func (p *fakecloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGameDataSource,
	}
}
