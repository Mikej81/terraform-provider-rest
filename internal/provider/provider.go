// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-rest/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &restProvider{}
)

// ProviderData holds the configured REST client for data sources and resources.
type ProviderData struct {
	Client *client.RestClient
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &restProvider{
			version: version,
		}
	}
}

// restProvider is the provider implementation.
type restProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *restProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rest"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *restProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Required:    true,
				Description: "The base URL for the REST API.",
			},
			// Token Authentication
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The API token for authenticating requests to the REST API.",
			},
			"api_header": schema.StringAttribute{
				Optional:    true,
				Description: "The HTTP header name for the API token (default: 'Authorization').",
			},
			// Certificate Authentication
			"client_cert": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client certificate for mTLS authentication (PEM format).",
			},
			"client_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client private key for mTLS authentication (PEM format).",
			},
			"client_cert_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to client certificate file for mTLS authentication.",
			},
			"client_key_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to client private key file for mTLS authentication.",
			},
			// PKCS12 Authentication
			"pkcs12_bundle": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "PKCS12 certificate bundle for authentication (base64 encoded).",
			},
			"pkcs12_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to PKCS12 certificate bundle file for authentication.",
			},
			"pkcs12_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for PKCS12 certificate bundle.",
			},
			// General Options
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Default timeout for HTTP requests in seconds (default: 30).",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable SSL certificate verification (default: false).",
			},
			"retry_attempts": schema.Int64Attribute{
				Optional:    true,
				Description: "Default number of retry attempts for failed requests (default: 3).",
			},
			"max_idle_conns": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of idle HTTP connections (default: 100).",
			},
		},
	}
}

// Configure prepares a rest API client for data sources and resources.
func (p *restProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Extract provider configuration values
	var config struct {
		APIURL         types.String `tfsdk:"api_url"`
		APIToken       types.String `tfsdk:"api_token"`
		APIHeader      types.String `tfsdk:"api_header"`
		ClientCert     types.String `tfsdk:"client_cert"`
		ClientKey      types.String `tfsdk:"client_key"`
		ClientCertFile types.String `tfsdk:"client_cert_file"`
		ClientKeyFile  types.String `tfsdk:"client_key_file"`
		PKCS12Bundle   types.String `tfsdk:"pkcs12_bundle"`
		PKCS12File     types.String `tfsdk:"pkcs12_file"`
		PKCS12Password types.String `tfsdk:"pkcs12_password"`
		Timeout        types.Int64  `tfsdk:"timeout"`
		Insecure       types.Bool   `tfsdk:"insecure"`
		RetryAttempts  types.Int64  `tfsdk:"retry_attempts"`
		MaxIdleConns   types.Int64  `tfsdk:"max_idle_conns"`
	}

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check that required configuration values are provided
	if config.APIURL.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Configuration",
			"The API URL must be provided for the provider to function",
		)
		return
	}

	// Validate authentication configuration - only one method should be provided
	authMethods := 0
	
	// Token authentication
	if !config.APIToken.IsNull() {
		authMethods++
	}
	
	// Certificate authentication (either inline or file-based)
	if (!config.ClientCert.IsNull() && !config.ClientKey.IsNull()) || 
		(!config.ClientCertFile.IsNull() && !config.ClientKeyFile.IsNull()) {
		authMethods++
	}
	
	// PKCS12 authentication (either inline or file-based)
	if !config.PKCS12Bundle.IsNull() || !config.PKCS12File.IsNull() {
		authMethods++
	}

	if authMethods == 0 {
		resp.Diagnostics.AddError(
			"Missing Authentication",
			"At least one authentication method must be provided: api_token, client certificates (cert+key), or pkcs12 bundle",
		)
		return
	}

	if authMethods > 1 {
		resp.Diagnostics.AddError(
			"Multiple Authentication Methods",
			"Only one authentication method should be provided: api_token, client certificates, or pkcs12 bundle",
		)
		return
	}

	// Validate certificate authentication completeness
	if (!config.ClientCert.IsNull() && config.ClientKey.IsNull()) || 
		(config.ClientCert.IsNull() && !config.ClientKey.IsNull()) {
		resp.Diagnostics.AddError(
			"Incomplete Certificate Authentication",
			"Both client_cert and client_key must be provided together",
		)
		return
	}

	if (!config.ClientCertFile.IsNull() && config.ClientKeyFile.IsNull()) || 
		(config.ClientCertFile.IsNull() && !config.ClientKeyFile.IsNull()) {
		resp.Diagnostics.AddError(
			"Incomplete Certificate Authentication",
			"Both client_cert_file and client_key_file must be provided together",
		)
		return
	}

	// Set default values
	timeout := time.Duration(30) * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	retryAttempts := 3
	if !config.RetryAttempts.IsNull() {
		retryAttempts = int(config.RetryAttempts.ValueInt64())
	}

	maxIdleConns := 100
	if !config.MaxIdleConns.IsNull() {
		maxIdleConns = int(config.MaxIdleConns.ValueInt64())
	}

	apiHeader := "Authorization"
	if !config.APIHeader.IsNull() {
		apiHeader = config.APIHeader.ValueString()
	}

	insecure := false
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	// Create the REST client configuration
	clientConfig := client.Config{
		BaseURL:       config.APIURL.ValueString(),
		Timeout:       timeout,
		Insecure:      insecure,
		RetryAttempts: retryAttempts,
		MaxIdleConns:  maxIdleConns,
	}

	// Configure authentication based on the method chosen
	if !config.APIToken.IsNull() {
		// Token authentication
		clientConfig.Token = config.APIToken.ValueString()
		clientConfig.TokenHeader = apiHeader
	} else if !config.ClientCert.IsNull() && !config.ClientKey.IsNull() {
		// Certificate authentication (inline)
		clientConfig.ClientCert = config.ClientCert.ValueString()
		clientConfig.ClientKey = config.ClientKey.ValueString()
	} else if !config.ClientCertFile.IsNull() && !config.ClientKeyFile.IsNull() {
		// Certificate authentication (file-based)
		clientConfig.ClientCertFile = config.ClientCertFile.ValueString()
		clientConfig.ClientKeyFile = config.ClientKeyFile.ValueString()
	} else if !config.PKCS12Bundle.IsNull() {
		// PKCS12 authentication (inline)
		clientConfig.PKCS12Bundle = config.PKCS12Bundle.ValueString()
		if !config.PKCS12Password.IsNull() {
			clientConfig.PKCS12Password = config.PKCS12Password.ValueString()
		}
	} else if !config.PKCS12File.IsNull() {
		// PKCS12 authentication (file-based)
		clientConfig.PKCS12File = config.PKCS12File.ValueString()
		if !config.PKCS12Password.IsNull() {
			clientConfig.PKCS12Password = config.PKCS12Password.ValueString()
		}
	}

	// Create the REST client
	restClient, err := client.NewRestClient(clientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Configuration Error",
			"Unable to create REST client: "+err.Error(),
		)
		return
	}

	// Store the client in the provider data
	providerData := &ProviderData{
		Client: restClient,
	}
	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

// DataSources defines the data sources implemented in the provider.
func (p *restProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRestDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *restProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRestResource,
	}
}
