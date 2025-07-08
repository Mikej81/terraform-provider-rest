// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-rest/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RestResource{}
var _ resource.ResourceWithImportState = &RestResource{}

func NewRestResource() resource.Resource {
	return &RestResource{}
}

// RestResource defines the resource implementation.
type RestResource struct {
	client *client.RestClient
}

// RestResourceModel describes the resource data model.
type RestResourceModel struct {
	Id              types.String            `tfsdk:"id"`
	Endpoint        types.String            `tfsdk:"endpoint"`
	Name            types.String            `tfsdk:"name"`
	Method          types.String            `tfsdk:"method"`
	Headers         map[string]types.String `tfsdk:"headers"`
	QueryParams     map[string]types.String `tfsdk:"query_params"`
	Body            types.String            `tfsdk:"body"`
	UpdateBody      types.String            `tfsdk:"update_body"`
	DestroyBody     types.String            `tfsdk:"destroy_body"`
	Response        types.String            `tfsdk:"response"`
	StatusCode      types.Int64             `tfsdk:"status_code"`
	ResponseHeaders types.Map               `tfsdk:"response_headers"`
	ResponseData    types.Map               `tfsdk:"response_data"`
	CreatedAt       types.String            `tfsdk:"created_at"`
	LastUpdated     types.String            `tfsdk:"last_updated"`
	Timeout         types.Int64             `tfsdk:"timeout"`
	Insecure        types.Bool              `tfsdk:"insecure"`
	RetryAttempts   types.Int64             `tfsdk:"retry_attempts"`
}

func (r *RestResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *RestResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "REST resource to create, read, update, and delete items via API with full HTTP method support.",

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
				MarkdownDescription: "The name of the item to be used for identification during read, update and delete operations.",
				Required:            true,
			},
			"method": schema.StringAttribute{
				MarkdownDescription: "The HTTP method to use for create operations (POST, PUT, PATCH). Default: POST.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("POST", "PUT", "PATCH"),
				},
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "Custom headers to include in requests.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"query_params": schema.MapAttribute{
				MarkdownDescription: "Query parameters to include in requests.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"body": schema.StringAttribute{
				MarkdownDescription: "The body for create requests. This can be a JSON object or any payload that the API expects.",
				Optional:            true,
			},
			"update_body": schema.StringAttribute{
				MarkdownDescription: "The body for update requests. If not specified, uses the same body as create.",
				Optional:            true,
			},
			"destroy_body": schema.StringAttribute{
				MarkdownDescription: "The body for delete requests. This can be a JSON object or any payload that the API expects.",
				Optional:            true,
			},
			"response": schema.StringAttribute{
				MarkdownDescription: "The response from the most recent API request.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "The HTTP status code from the most recent API request.",
				Computed:            true,
			},
			"response_headers": schema.MapAttribute{
				MarkdownDescription: "The response headers from the API request.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"response_data": schema.MapAttribute{
				MarkdownDescription: "The parsed response data as key-value pairs (for JSON responses).",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the resource was created.",
				Computed:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the resource was last updated.",
				Computed:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for requests in seconds.",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Disable SSL certificate verification.",
				Optional:            true,
			},
			"retry_attempts": schema.Int64Attribute{
				MarkdownDescription: "Number of retry attempts for failed requests.",
				Optional:            true,
			},
		},
	}
}

func (r *RestResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
}

// buildRequestOptions creates client.RequestOptions from resource model
func (r *RestResource) buildRequestOptions(ctx context.Context, data *RestResourceModel, method string, body string) client.RequestOptions {
	options := client.RequestOptions{
		Method:   method,
		Endpoint: data.Endpoint.ValueString(),
	}

	// Add body if provided
	if body != "" {
		options.Body = []byte(body)
	}

	// Add custom headers
	if data.Headers != nil {
		customHeaders := make(map[string]string)
		for key, value := range data.Headers {
			customHeaders[key] = value.ValueString()
		}
		options.Headers = customHeaders
	}

	// Add query parameters
	if data.QueryParams != nil {
		queryParams := make(map[string]string)
		for key, value := range data.QueryParams {
			queryParams[key] = value.ValueString()
		}
		options.QueryParams = queryParams
	}

	// Set timeout if provided
	if !data.Timeout.IsNull() {
		options.Timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	// Set retry attempts if provided
	if !data.RetryAttempts.IsNull() {
		options.Retries = int(data.RetryAttempts.ValueInt64())
	}

	return options
}

// processResponse handles the HTTP response and updates the model
func (r *RestResource) processResponse(ctx context.Context, response *client.Response, data *RestResourceModel) error {
	// Set response data
	data.StatusCode = types.Int64Value(int64(response.StatusCode))
	data.Response = types.StringValue(string(response.Body))

	// Set response headers
	responseHeaders := make(map[string]attr.Value)
	for key, values := range response.Headers {
		if len(values) > 0 {
			responseHeaders[key] = types.StringValue(values[0])
		}
	}
	headersMap, diags := types.MapValue(types.StringType, responseHeaders)
	if diags.HasError() {
		tflog.Warn(ctx, "failed to create response headers map", map[string]interface{}{
			"errors": diags.Errors(),
		})
		data.ResponseHeaders = types.MapNull(types.StringType)
	} else {
		data.ResponseHeaders = headersMap
	}

	// Parse JSON response data for dynamic output
	responseData := make(map[string]attr.Value)
	if len(response.Body) > 0 {
		var parsed map[string]interface{}
		if err := json.Unmarshal(response.Body, &parsed); err == nil {
			for key, value := range parsed {
				// Convert all values to strings for simplicity
				if value != nil {
					switch v := value.(type) {
					case string:
						responseData[key] = types.StringValue(v)
					case float64:
						responseData[key] = types.StringValue(fmt.Sprintf("%.0f", v))
					case bool:
						responseData[key] = types.StringValue(fmt.Sprintf("%t", v))
					default:
						// Convert complex types to JSON string
						if jsonBytes, err := json.Marshal(value); err == nil {
							responseData[key] = types.StringValue(string(jsonBytes))
						}
					}
				}
			}
		}
	}
	dataMap, diags := types.MapValue(types.StringType, responseData)
	if diags.HasError() {
		tflog.Warn(ctx, "failed to create response data map", map[string]interface{}{
			"errors": diags.Errors(),
		})
		data.ResponseData = types.MapNull(types.StringType)
	} else {
		data.ResponseData = dataMap
	}

	// Set timestamps
	currentTime := time.Now().UTC().Format(time.RFC3339)
	if data.CreatedAt.IsNull() || data.CreatedAt.IsUnknown() {
		data.CreatedAt = types.StringValue(currentTime)
	}
	data.LastUpdated = types.StringValue(currentTime)

	// Extract ID from response if it exists and it's not already set
	if data.Id.IsNull() || data.Id.IsUnknown() {
		if len(response.Body) > 0 {
			var parsed map[string]interface{}
			if err := json.Unmarshal(response.Body, &parsed); err == nil {
				if idValue, ok := parsed["id"].(string); ok {
					data.Id = types.StringValue(idValue)
					tflog.Debug(ctx, "extracted ID from response", map[string]interface{}{
						"id": idValue,
					})
					return nil
				}
			}
		}

		// Fallback ID generation only if no ID was found in response
		fallbackId := fmt.Sprintf("%s_%s", data.Endpoint.ValueString(), data.Name.ValueString())
		data.Id = types.StringValue(fallbackId)
		tflog.Debug(ctx, "generated fallback ID", map[string]interface{}{
			"id": fallbackId,
		})
	}

	return nil
}

func (r *RestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RestResourceModel

	// Read the planned state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default method to POST if not provided
	method := "POST"
	if !data.Method.IsNull() {
		method = data.Method.ValueString()
	}
	data.Method = types.StringValue(method)

	// Get request body
	requestBody := ""
	if !data.Body.IsNull() {
		requestBody = data.Body.ValueString()
	}

	// Build request options
	options := r.buildRequestOptions(ctx, &data, method, requestBody)

	tflog.Trace(ctx, "creating REST resource", map[string]interface{}{
		"method":   method,
		"endpoint": data.Endpoint.ValueString(),
	})

	// Make the request
	response, err := r.client.Do(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP Request Failed",
			fmt.Sprintf("Unable to send %s request to %s: %s", method, data.Endpoint.ValueString(), err),
		)
		return
	}

	// Check for successful status codes
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Received non-success response code: %d, Response: %s", response.StatusCode, string(response.Body)),
		)
		return
	}

	// Process response
	if err := r.processResponse(ctx, response, &data); err != nil {
		resp.Diagnostics.AddError("Response Processing Error", err.Error())
		return
	}

	tflog.Trace(ctx, "created REST resource", map[string]interface{}{
		"method":      method,
		"endpoint":    data.Endpoint.ValueString(),
		"status_code": response.StatusCode,
		"id":          data.Id.ValueString(),
	})

	// Save the resource state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RestResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build URL for GET request - typically append name to endpoint
	endpoint := data.Endpoint.ValueString()
	if !data.Name.IsNull() {
		endpoint = fmt.Sprintf("%s/%s", endpoint, data.Name.ValueString())
	}

	// Build request options for GET
	options := r.buildRequestOptions(ctx, &data, "GET", "")
	// Override endpoint for read operation (with name appended)
	options.Endpoint = endpoint

	tflog.Trace(ctx, "reading REST resource", map[string]interface{}{
		"endpoint": endpoint,
		"id":       data.Id.ValueString(),
	})

	// Make the request
	response, err := r.client.Do(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP Request Failed",
			fmt.Sprintf("Unable to send GET request to %s: %s", endpoint, err),
		)
		return
	}

	// Handle 404 as resource not found
	if response.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	// Check for other success status codes
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Received non-success response code: %d, Response: %s", response.StatusCode, string(response.Body)),
		)
		return
	}

	// Process the response using the same logic as Create/Update
	if err := r.processResponse(ctx, response, &data); err != nil {
		resp.Diagnostics.AddError("Response Processing Error", err.Error())
		return
	}

	// Check for drift in the response body if we have an expected structure
	if len(response.Body) > 0 {
		var parsed map[string]interface{}
		if err := json.Unmarshal(response.Body, &parsed); err == nil {
			// Check if the resource still exists by verifying it has the expected ID
			if currentId, ok := parsed["id"].(string); ok {
				expectedId := data.Id.ValueString()
				if currentId != expectedId {
					tflog.Warn(ctx, "detected ID drift", map[string]interface{}{
						"expected_id": expectedId,
						"current_id":  currentId,
					})
					// Update the ID to match what's on the server
					data.Id = types.StringValue(currentId)
				}
			}
		}
	}

	tflog.Trace(ctx, "read REST resource", map[string]interface{}{
		"endpoint":    endpoint,
		"status_code": response.StatusCode,
		"id":          data.Id.ValueString(),
	})

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RestResourceModel

	// Read the plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build URL for PUT/PATCH request - typically append name to endpoint
	endpoint := data.Endpoint.ValueString()
	if !data.Name.IsNull() {
		endpoint = fmt.Sprintf("%s/%s", endpoint, data.Name.ValueString())
	}

	// Determine update method - prefer PUT, but allow PATCH
	method := "PUT"
	if !data.Method.IsNull() && data.Method.ValueString() == "PATCH" {
		method = "PATCH"
	}

	// Get update body - prefer update_body, fallback to body
	requestBody := ""
	if !data.UpdateBody.IsNull() {
		requestBody = data.UpdateBody.ValueString()
	} else if !data.Body.IsNull() {
		requestBody = data.Body.ValueString()
	}

	// Build request options
	options := client.RequestOptions{
		Method:   method,
		Endpoint: endpoint,
	}

	if requestBody != "" {
		options.Body = []byte(requestBody)
	}

	// Add custom headers
	if data.Headers != nil {
		customHeaders := make(map[string]string)
		for key, value := range data.Headers {
			customHeaders[key] = value.ValueString()
		}
		options.Headers = customHeaders
	}

	// Add query parameters
	if data.QueryParams != nil {
		queryParams := make(map[string]string)
		for key, value := range data.QueryParams {
			queryParams[key] = value.ValueString()
		}
		options.QueryParams = queryParams
	}

	// Set timeout if provided
	if !data.Timeout.IsNull() {
		options.Timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	// Set retry attempts if provided
	if !data.RetryAttempts.IsNull() {
		options.Retries = int(data.RetryAttempts.ValueInt64())
	}

	tflog.Trace(ctx, "updating REST resource", map[string]interface{}{
		"method":   method,
		"endpoint": endpoint,
		"id":       data.Id.ValueString(),
	})

	// Make the request
	response, err := r.client.Do(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP Request Failed",
			fmt.Sprintf("Unable to send %s request to %s: %s", method, endpoint, err),
		)
		return
	}

	// Check for successful status codes
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Received non-success response code: %d, Response: %s", response.StatusCode, string(response.Body)),
		)
		return
	}

	// Process response
	if err := r.processResponse(ctx, response, &data); err != nil {
		resp.Diagnostics.AddError("Response Processing Error", err.Error())
		return
	}

	tflog.Trace(ctx, "updated REST resource", map[string]interface{}{
		"method":      method,
		"endpoint":    endpoint,
		"status_code": response.StatusCode,
		"id":          data.Id.ValueString(),
	})

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The ID from the import should be in the format "endpoint/name"
	parts := strings.Split(req.ID, "/")
	if len(parts) < 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'endpoint/name', got: %s", req.ID),
		)
		return
	}

	// Handle cases where endpoint might contain slashes
	name := parts[len(parts)-1]
	endpoint := "/" + strings.Join(parts[:len(parts)-1], "/")

	// Set the basic attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("endpoint"), endpoint)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	// Set default method
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("method"), "POST")...)

	tflog.Trace(ctx, "imported REST resource", map[string]interface{}{
		"endpoint": endpoint,
		"name":     name,
		"id":       req.ID,
	})
}

func (r *RestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RestResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build URL for DELETE request - typically append name to endpoint
	endpoint := data.Endpoint.ValueString()
	if !data.Name.IsNull() {
		endpoint = fmt.Sprintf("%s/%s", endpoint, data.Name.ValueString())
	}

	// Get destroy body if provided
	requestBody := ""
	if !data.DestroyBody.IsNull() {
		requestBody = data.DestroyBody.ValueString()
	}

	// Build request options
	options := client.RequestOptions{
		Method:   "DELETE",
		Endpoint: endpoint,
	}

	if requestBody != "" {
		options.Body = []byte(requestBody)
	}

	// Add custom headers
	if data.Headers != nil {
		customHeaders := make(map[string]string)
		for key, value := range data.Headers {
			customHeaders[key] = value.ValueString()
		}
		options.Headers = customHeaders
	}

	// Add query parameters
	if data.QueryParams != nil {
		queryParams := make(map[string]string)
		for key, value := range data.QueryParams {
			queryParams[key] = value.ValueString()
		}
		options.QueryParams = queryParams
	}

	// Set timeout if provided
	if !data.Timeout.IsNull() {
		options.Timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	// Set retry attempts if provided
	if !data.RetryAttempts.IsNull() {
		options.Retries = int(data.RetryAttempts.ValueInt64())
	}

	tflog.Trace(ctx, "deleting REST resource", map[string]interface{}{
		"endpoint": endpoint,
		"id":       data.Id.ValueString(),
	})

	// Make the request
	response, err := r.client.Do(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP Request Failed",
			fmt.Sprintf("Unable to send DELETE request to %s: %s", endpoint, err),
		)
		return
	}

	// Accept 200, 202, 204, and 404 as successful deletion
	acceptableStatusCodes := []int{200, 202, 204, 404}
	successful := false
	for _, code := range acceptableStatusCodes {
		if response.StatusCode == code {
			successful = true
			break
		}
	}

	if !successful {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Received non-success response code: %d, Response: %s", response.StatusCode, string(response.Body)),
		)
		return
	}

	tflog.Trace(ctx, "deleted REST resource", map[string]interface{}{
		"endpoint":    endpoint,
		"status_code": response.StatusCode,
		"id":          data.Id.ValueString(),
	})
}
