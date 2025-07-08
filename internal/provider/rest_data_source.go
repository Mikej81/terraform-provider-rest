// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-rest/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RestDataSource{}

func NewRestDataSource() datasource.DataSource {
	return &RestDataSource{}
}

// RestDataSource defines the data source implementation.
type RestDataSource struct {
	client *client.RestClient
}

// RestDataSourceModel describes the data source data model.
type RestDataSourceModel struct {
	Id            types.String            `tfsdk:"id"`
	Endpoint      types.String            `tfsdk:"endpoint"`
	Method        types.String            `tfsdk:"method"`
	Headers       map[string]types.String `tfsdk:"headers"`
	Body          types.String            `tfsdk:"body"`
	Response      types.String            `tfsdk:"response"`
	StatusCode    types.Int64             `tfsdk:"status_code"`
	ParsedData    map[string]types.String `tfsdk:"parsed_data"`
	Timeout       types.Int64             `tfsdk:"timeout"`
	Insecure      types.Bool              `tfsdk:"insecure"`
	RetryAttempts types.Int64             `tfsdk:"retry_attempts"`
}

func (d *RestDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data"
}

func (d *RestDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "REST data source to make API requests.",

		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The API endpoint to send the request to (relative to the base URL).",
				Required:            true,
			},
			"method": schema.StringAttribute{
				MarkdownDescription: "The HTTP method to use (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"),
				},
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "Custom headers to include in the request.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"body": schema.StringAttribute{
				MarkdownDescription: "The request body, used with POST or PUT methods.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier for the request.",
				Computed:            true,
			},
			"response": schema.StringAttribute{
				MarkdownDescription: "The response from the request.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "The HTTP status code from the request.",
				Computed:            true,
			},
			"parsed_data": schema.MapAttribute{
				MarkdownDescription: "Parsed JSON attributes from the response body.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for the request in seconds.",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Disable SSL certificate verification.",
				Optional:            true,
			},
			"retry_attempts": schema.Int64Attribute{
				MarkdownDescription: "Number of retry attempts for the request.",
				Optional:            true,
			},
		},
	}
}

func (d *RestDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
}

func (d *RestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RestDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default method to GET if not provided
	method := "GET"
	if !data.Method.IsNull() {
		method = data.Method.ValueString()
	}

	// Prepare request options
	requestOptions := client.RequestOptions{
		Method:   method,
		Endpoint: data.Endpoint.ValueString(),
	}

	// Add request body if provided
	if !data.Body.IsNull() {
		requestOptions.Body = []byte(data.Body.ValueString())
	}

	// Add custom headers if provided
	if data.Headers != nil {
		customHeaders := make(map[string]string)
		for key, value := range data.Headers {
			customHeaders[key] = value.ValueString()
		}
		requestOptions.Headers = customHeaders
	}

	// Set timeout if provided
	if !data.Timeout.IsNull() {
		requestOptions.Timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	// Set retry attempts if provided
	if !data.RetryAttempts.IsNull() {
		requestOptions.Retries = int(data.RetryAttempts.ValueInt64())
	}

	// Make the request using the REST client
	response, err := d.client.Do(ctx, requestOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP Request Failed",
			fmt.Sprintf("Unable to send %s request to %s: %s", method, data.Endpoint.ValueString(), err),
		)
		return
	}

	// Set the status code and response body
	data.StatusCode = types.Int64Value(int64(response.StatusCode))
	data.Response = types.StringValue(string(response.Body))
	data.Id = types.StringValue(fmt.Sprintf("%s_%s", method, data.Endpoint.ValueString()))

	// Set the computed method value
	data.Method = types.StringValue(method)

	// Parse the JSON response body into dynamic attributes
	if len(response.Body) > 0 {
		var parsed map[string]interface{}
		if err := json.Unmarshal(response.Body, &parsed); err != nil {
			// Log the error but don't fail - not all responses are JSON
			tflog.Debug(ctx, "Response body is not valid JSON", map[string]interface{}{
				"endpoint": data.Endpoint.ValueString(),
				"error":    err.Error(),
			})
		} else {
			parsedData := make(map[string]types.String)
			for key, value := range parsed {
				if strValue, ok := value.(string); ok {
					parsedData[key] = types.StringValue(strValue)
				} else {
					parsedData[key] = types.StringValue(fmt.Sprintf("%v", value))
				}
			}
			data.ParsedData = parsedData
		}
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a REST data source", map[string]interface{}{
		"method":      method,
		"endpoint":    data.Endpoint.ValueString(),
		"status_code": response.StatusCode,
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
