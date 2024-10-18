// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &genericrestProvider{}
)

// APIClient is a simple REST client for interacting with the API.
type APIClient struct {
	BaseURL string
	Token   string
	Header  string
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &genericrestProvider{
			version: version,
		}
	}
}

// genericrestProvider is the provider implementation.
type genericrestProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *genericrestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "genericrest"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *genericrestProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Required:    true,
				Description: "The API token for authenticating requests to the REST API.",
			},
			"api_url": schema.StringAttribute{
				Required:    true,
				Description: "The base URL for the REST API.",
			},
			"api_header": schema.StringAttribute{
				Optional:    true,
				Description: "Additional headers to include in API requests.",
			},
		},
	}
}

// Configure prepares a genericrest API client for data sources and resources.
func (p *genericrestProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Extract provider configuration values
	var config struct {
		APIToken  types.String `tfsdk:"api_token"`
		APIURL    types.String `tfsdk:"api_url"`
		APIHeader types.String `tfsdk:"api_header"`
	}

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check that required configuration values are provided
	if config.APIToken.IsNull() || config.APIURL.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Configuration",
			"The API token and API URL must be provided for the provider to function",
		)
		return
	}

	// Create the API client
	apiClient := &APIClient{
		BaseURL: config.APIURL.ValueString(),
		Token:   config.APIToken.ValueString(),
		Header:  config.APIHeader.ValueString(),
	}

	// Store the client in the provider data
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

// DataSources defines the data sources implemented in the provider.
func (p *genericrestProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRestDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *genericrestProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRestResource,
	}
}
