// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RestDataSource{}

func NewRestDataSource() datasource.DataSource {
	return &RestDataSource{}
}

// RestDataSource defines the data source implementation.
type RestDataSource struct {
	client  *http.Client
	baseURL string
	token   string
	header  string
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
				MarkdownDescription: "The HTTP method to use (GET, POST, PUT, DELETE).",
				Optional:            true,
				Computed:            true,
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

	apiClient, ok := req.ProviderData.(*APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = &http.Client{}
	d.baseURL = apiClient.BaseURL
	d.token = apiClient.Token
	d.header = apiClient.Header
}

func (d *RestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RestDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...) // Read the config into data
	if resp.Diagnostics.HasError() {
		return
	}

	// Default method to GET if not provided
	method := "GET"
	if !data.Method.IsNull() {
		method = data.Method.ValueString()
	}

	// Build the full URL by combining base URL and endpoint
	reqURL := fmt.Sprintf("%s%s", d.baseURL, data.Endpoint.ValueString())

	// Create request body if applicable
	var requestBody *strings.Reader
	if !data.Body.IsNull() && (method == "POST" || method == "PUT") {
		requestBody = strings.NewReader(data.Body.ValueString())
	} else {
		requestBody = strings.NewReader("")
	}

	// Configure HTTP client with timeout and insecure options
	client := &http.Client{}
	if !data.Timeout.IsNull() {
		client.Timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}
	if !data.Insecure.IsNull() && data.Insecure.ValueBool() {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = transport
	}

	// Make the request to the specified endpoint
	httpReq, err := http.NewRequest(method, reqURL, requestBody)
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	// Add the default header for authentication
	httpReq.Header.Set(d.header, d.token)
	httpReq.Header.Set("Content-Type", "application/json")

	// Add custom headers if provided
	if data.Headers != nil {
		for key, value := range data.Headers {
			httpReq.Header.Set(key, value.ValueString())
		}
	}

	// Retry logic
	retryAttempts := 1
	if !data.RetryAttempts.IsNull() {
		retryAttempts = int(data.RetryAttempts.ValueInt64())
	}

	var httpResp *http.Response
	for i := 0; i < retryAttempts; i++ {
		httpResp, err = client.Do(httpReq)
		if err == nil {
			break
		}
		tflog.Warn(ctx, fmt.Sprintf("Request attempt %d failed: %s", i+1, err))
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send %s request: %s", method, err))
		return
	}
	defer httpResp.Body.Close()

	// Set the status code in the response model
	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))

	// Read the response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Unable to read response body: %s", err))
		return
	}

	data.Response = types.StringValue(string(body))
	data.Id = types.StringValue(reqURL)

	// Parse the JSON response body into dynamic attributes
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response body as JSON: %s", err))
		return
	}

	parsedData := make(map[string]types.String)
	for key, value := range parsed {
		if strValue, ok := value.(string); ok {
			parsedData[key] = types.StringValue(strValue)
		} else {
			parsedData[key] = types.StringValue(fmt.Sprintf("%v", value))
		}
	}
	data.ParsedData = parsedData

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a REST data source", map[string]interface{}{
		"endpoint":    reqURL,
		"status_code": httpResp.StatusCode,
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Write the data to state
}
