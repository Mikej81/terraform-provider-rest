// Copyright (c) HashiCorp, Inc.

package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RestResource{}

func NewRestResource() resource.Resource {
	return &RestResource{}
}

// RestResource defines the resource implementation.
type RestResource struct {
	client  *http.Client
	baseURL string
	token   string
	header  string
}

// RestResourceModel describes the resource data model.
type RestResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Endpoint      types.String `tfsdk:"endpoint"`
	Name          types.String `tfsdk:"name"`
	Response      types.String `tfsdk:"response"`
	StatusCode    types.Int64  `tfsdk:"status_code"`
	Body          types.String `tfsdk:"body"`
	DestroyBody   types.String `tfsdk:"destroy_body"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	Insecure      types.Bool   `tfsdk:"insecure"`
	RetryAttempts types.Int64  `tfsdk:"retry_attempts"`
}

func (r *RestResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *RestResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "REST resource to create, read, update, and delete items via API.",

		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The API endpoint to send the request to (relative to the base URL).",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier for the created resource.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the item to be used for identification during update and delete operations.",
				Required:            true,
			},
			"body": schema.StringAttribute{
				MarkdownDescription: "The body for the request. This can be a JSON object or any payload that the API expects.",
				Optional:            true,
			},
			"destroy_body": schema.StringAttribute{
				MarkdownDescription: "The body for the destroy. This can be a JSON object or any payload that the API expects.",
				Optional:            true,
			},
			"response": schema.StringAttribute{
				MarkdownDescription: "The response from the API request.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "The HTTP status code from the API request.",
				Computed:            true,
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

func (r *RestResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = &http.Client{}
	r.baseURL = apiClient.BaseURL
	r.token = apiClient.Token
	r.header = apiClient.Header
}

// Working, commenting out to test new mods
// func (r *RestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
// 	var data RestResourceModel

// 	// Read the planned state (attributes from HCL)
// 	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// Build the full URL using baseURL and endpoint provided by the user
// 	reqURL := fmt.Sprintf("%s%s", r.baseURL, data.Endpoint.ValueString())

// 	// Use the `Body` value from the plan
// 	requestBody := data.Body.ValueString()
// 	if requestBody == "" {
// 		requestBody = "{}" // Default to empty JSON object if no body is provided
// 	}

// 	tflog.Trace(ctx, "tried to create a REST resource", map[string]interface{}{
// 		"endpoint":    reqURL,
// 		"requestBody": requestBody,
// 	})

// 	// Create the HTTP POST request
// 	httpReq, err := http.NewRequest("POST", reqURL, bytes.NewBuffer([]byte(requestBody)))
// 	if err != nil {
// 		resp.Diagnostics.AddError("Request Creation Error", fmt.Sprintf("Unable to create request: %s", err))
// 		return
// 	}

// 	// Set headers
// 	httpReq.Header.Set(r.header, r.token)
// 	httpReq.Header.Set("Content-Type", "application/json")

// 	// Send the HTTP request
// 	httpResp, err := r.client.Do(httpReq)
// 	if err != nil {
// 		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send POST request: %s", err))
// 		return
// 	}
// 	defer httpResp.Body.Close()

// 	// Read the response body
// 	body, err := ioutil.ReadAll(httpResp.Body)
// 	if err != nil {
// 		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Unable to read response body: %s", err))
// 		return
// 	}

// 	// Set the HTTP status code in the resource state
// 	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))

// 	// Check if the response is successful
// 	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
// 		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Received non-success response code: %d, Response: %s", httpResp.StatusCode, string(body)))
// 		return
// 	}

// 	// Save the full response body
// 	data.Response = types.StringValue(string(body))

// 	// Extract `id` from response if it exists
// 	var parsed map[string]interface{}
// 	if err := json.Unmarshal(body, &parsed); err == nil {
// 		if idValue, ok := parsed["id"].(string); ok {
// 			data.Id = types.StringValue(idValue)
// 		} else {
// 			data.Id = types.StringValue(reqURL) // Fallback to request URL as ID
// 		}
// 	} else {
// 		data.Id = types.StringValue(reqURL) // Fallback to request URL as ID if parsing fails
// 	}

// 	// Write logs using the tflog package
// 	tflog.Trace(ctx, "created a REST resource", map[string]interface{}{
// 		"endpoint":    reqURL,
// 		"status_code": httpResp.StatusCode,
// 	})

// 	// Save the resource state
// 	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
// }

func (r *RestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RestResourceModel

	// Read the planned state (attributes from HCL)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the full URL using baseURL and endpoint provided by the user
	reqURL := fmt.Sprintf("%s%s", r.baseURL, data.Endpoint.ValueString())

	// Use the `Body` value from the plan
	requestBody := data.Body.ValueString()
	if requestBody == "" {
		requestBody = "{}" // Default to empty JSON object if no body is provided
	}

	tflog.Trace(ctx, "tried to create a REST resource", map[string]interface{}{
		"endpoint":    reqURL,
		"requestBody": requestBody,
	})

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

	// Create the HTTP POST request
	httpReq, err := http.NewRequest("POST", reqURL, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	// Set headers
	httpReq.Header.Set(r.header, r.token)
	httpReq.Header.Set("Content-Type", "application/json")

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send POST request: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Unable to read response body: %s", err))
		return
	}

	// Set the HTTP status code in the resource state
	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))

	// Check if the response is successful
	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Received non-success response code: %d, Response: %s", httpResp.StatusCode, string(body)))
		return
	}

	// Save the full response body
	data.Response = types.StringValue(string(body))

	// Extract `id` from response if it exists
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if idValue, ok := parsed["id"].(string); ok {
			data.Id = types.StringValue(idValue)
		} else {
			data.Id = types.StringValue(reqURL) // Fallback to request URL as ID
		}
	} else {
		data.Id = types.StringValue(reqURL) // Fallback to request URL as ID if parsing fails
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a REST resource", map[string]interface{}{
		"endpoint":    reqURL,
		"status_code": httpResp.StatusCode,
	})

	// Save the resource state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Write the data to state
}

func (r *RestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RestResourceModel

	// Debugging: Log state retrieval
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("State Retrieval Error", "Error while retrieving state")
		return
	}

	// Construct the URL using baseURL, endpoint, and object name
	url := fmt.Sprintf("%s%s/%s", r.baseURL, data.Endpoint.ValueString(), data.Name.ValueString())

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

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", fmt.Sprintf("Unable to create request: %s", err))
		resp.Diagnostics.AddWarning("Debug URL", fmt.Sprintf("Attempted URL: %s", url))
		return
	}

	httpReq.Header.Set(r.header, r.token)

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send GET request: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Set the status code in the response model
	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Received non-200 response code: %d", httpResp.StatusCode))
		return
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Unable to read response body: %s", err))
		return
	}

	data.Response = types.StringValue(string(body))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save the read data back into the state
}

func (r *RestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RestResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read Terraform plan data
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the full URL using baseURL and endpoint provided by the user
	reqURL := fmt.Sprintf("%s%s/%s", r.baseURL, data.Endpoint.ValueString(), data.Name.ValueString())

	// Use the `Body` value from the plan
	requestBody := data.Body.ValueString()
	if requestBody == "" {
		requestBody = "{}" // Default to empty JSON object if no body is provided
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

	// Create the HTTP PUT request
	httpReq, err := http.NewRequest("PUT", reqURL, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	httpReq.Header.Set(r.header, r.token)
	httpReq.Header.Set("Content-Type", "application/json")

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send PUT request: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Set the status code in the response model
	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Received non-200 response code: %d", httpResp.StatusCode))
		return
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Unable to read response body: %s", err))
		return
	}

	data.Response = types.StringValue(string(body))

	tflog.Trace(ctx, "updated a REST resource", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save updated data into Terraform state
}

func (r *RestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RestResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read Terraform prior state data
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the full URL using baseURL, endpoint, and name provided by the user
	reqURL := fmt.Sprintf("%s%s/%s", r.baseURL, data.Endpoint.ValueString(), data.Name.ValueString())

	// Use the `DestroyBody` value from the plan
	destroyRequestBody := data.DestroyBody.ValueString()
	if destroyRequestBody == "" {
		destroyRequestBody = "{}" // Default to empty JSON object if no body is provided
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

	httpReq, err := http.NewRequest("DELETE", reqURL, bytes.NewBuffer([]byte(destroyRequestBody)))
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	httpReq.Header.Set(r.header, r.token)
	httpReq.Header.Set("Content-Type", "application/json")

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send DELETE request: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Set the status code in the response model
	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Received non-success response code: %d", httpResp.StatusCode))
		return
	}

	tflog.Trace(ctx, "deleted a REST resource", map[string]interface{}{
		"name":        data.Name.ValueString(),
		"status_code": httpResp.StatusCode,
	})
}
