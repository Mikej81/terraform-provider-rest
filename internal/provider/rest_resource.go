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
	CreateMethod    types.String            `tfsdk:"create_method"`
	ReadMethod      types.String            `tfsdk:"read_method"`
	UpdateMethod    types.String            `tfsdk:"update_method"`
	DeleteMethod    types.String            `tfsdk:"delete_method"`
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
	// Conditional operations
	ExpectedStatus types.List   `tfsdk:"expected_status"`
	OnSuccess      types.String `tfsdk:"on_success"`
	OnFailure      types.String `tfsdk:"on_failure"`
	FailOnStatus   types.List   `tfsdk:"fail_on_status"`
	RetryOnStatus  types.List   `tfsdk:"retry_on_status"`
	// Drift detection configuration
	IgnoreFields   types.List `tfsdk:"ignore_fields"`
	DriftDetection types.Bool `tfsdk:"drift_detection"`
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
				MarkdownDescription: "The HTTP method to use for create operations (POST, PUT, PATCH). Default: POST. DEPRECATED: Use create_method, read_method, update_method, delete_method instead.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("POST", "PUT", "PATCH"),
				},
			},
			"create_method": schema.StringAttribute{
				MarkdownDescription: "The HTTP method to use for create operations (POST, PUT, PATCH). Default: POST.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("POST", "PUT", "PATCH"),
				},
			},
			"read_method": schema.StringAttribute{
				MarkdownDescription: "The HTTP method to use for read operations (GET, POST, HEAD). Default: GET.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "POST", "HEAD"),
				},
			},
			"update_method": schema.StringAttribute{
				MarkdownDescription: "The HTTP method to use for update operations (PUT, PATCH, POST). Default: PUT.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("PUT", "PATCH", "POST"),
				},
			},
			"delete_method": schema.StringAttribute{
				MarkdownDescription: "The HTTP method to use for delete operations (DELETE, POST, PUT). Default: DELETE.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("DELETE", "POST", "PUT"),
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
			"expected_status": schema.ListAttribute{
				MarkdownDescription: "List of expected HTTP status codes for successful operations. If specified, only these status codes will be considered successful.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"on_success": schema.StringAttribute{
				MarkdownDescription: "Action to take on successful response: 'continue' (default), 'stop', or 'retry'.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("continue", "stop", "retry"),
				},
			},
			"on_failure": schema.StringAttribute{
				MarkdownDescription: "Action to take on failed response: 'fail' (default), 'continue', or 'retry'.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("fail", "continue", "retry"),
				},
			},
			"fail_on_status": schema.ListAttribute{
				MarkdownDescription: "List of HTTP status codes that should be treated as failures, even if they would normally be considered successful.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"retry_on_status": schema.ListAttribute{
				MarkdownDescription: "List of HTTP status codes that should trigger a retry, in addition to the standard retryable codes.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"ignore_fields": schema.ListAttribute{
				MarkdownDescription: "List of field names in the API response to ignore during drift detection. Useful for server-side metadata fields like 'created_at', 'updated_at', 'etag', etc.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"drift_detection": schema.BoolAttribute{
				MarkdownDescription: "Enable drift detection for response body changes. When enabled, compares planned vs actual response to detect configuration drift. Default: true.",
				Optional:            true,
				Computed:            true,
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

// checkStatusCode evaluates if a status code should be considered successful based on conditional rules
func (r *RestResource) checkStatusCode(ctx context.Context, statusCode int, data *RestResourceModel) (bool, string) {
	// Check if this status should be treated as a failure
	if !data.FailOnStatus.IsNull() {
		failStatuses := make([]int64, 0, len(data.FailOnStatus.Elements()))
		diags := data.FailOnStatus.ElementsAs(ctx, &failStatuses, false)
		if !diags.HasError() {
			for _, failStatus := range failStatuses {
				if int(failStatus) == statusCode {
					action := "fail"
					if !data.OnFailure.IsNull() {
						action = data.OnFailure.ValueString()
					}
					return false, action
				}
			}
		}
	}

	// Check expected status codes if specified
	if !data.ExpectedStatus.IsNull() {
		expectedStatuses := make([]int64, 0, len(data.ExpectedStatus.Elements()))
		diags := data.ExpectedStatus.ElementsAs(ctx, &expectedStatuses, false)
		if !diags.HasError() {
			for _, expectedStatus := range expectedStatuses {
				if int(expectedStatus) == statusCode {
					action := "continue"
					if !data.OnSuccess.IsNull() {
						action = data.OnSuccess.ValueString()
					}
					return true, action
				}
			}
			// If expected statuses are specified but this code isn't in the list, treat as failure
			action := "fail"
			if !data.OnFailure.IsNull() {
				action = data.OnFailure.ValueString()
			}
			return false, action
		}
	}

	// Default behavior: 2xx codes are successful
	if statusCode >= 200 && statusCode < 300 {
		action := "continue"
		if !data.OnSuccess.IsNull() {
			action = data.OnSuccess.ValueString()
		}
		return true, action
	}

	// Non-2xx codes are failures by default
	action := "fail"
	if !data.OnFailure.IsNull() {
		action = data.OnFailure.ValueString()
	}
	return false, action
}

// shouldRetryOnStatus checks if the status code should trigger a retry
func (r *RestResource) shouldRetryOnStatus(ctx context.Context, statusCode int, data *RestResourceModel) bool {
	// Check custom retry status codes
	if !data.RetryOnStatus.IsNull() {
		retryStatuses := make([]int64, 0, len(data.RetryOnStatus.Elements()))
		diags := data.RetryOnStatus.ElementsAs(ctx, &retryStatuses, false)
		if !diags.HasError() {
			for _, retryStatus := range retryStatuses {
				if int(retryStatus) == statusCode {
					return true
				}
			}
		}
	}

	// Standard retryable status codes (429, 5xx)
	return statusCode == 429 || (statusCode >= 500 && statusCode < 600)
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

	// Resolve the HTTP method for create operation
	method := r.resolveMethodForOperation(&data, "create")

	// Set computed values for backward compatibility and visibility
	data.CreateMethod = types.StringValue(method)
	if data.Method.IsNull() {
		data.Method = types.StringValue(method)
	}

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

	// Check status code using conditional logic
	isSuccess, action := r.checkStatusCode(ctx, response.StatusCode, &data)

	if !isSuccess {
		switch action {
		case "continue":
			tflog.Warn(ctx, "continuing despite failed status code", map[string]interface{}{
				"status_code": response.StatusCode,
				"response":    string(response.Body),
			})
		case "retry":
			if r.shouldRetryOnStatus(ctx, response.StatusCode, &data) {
				resp.Diagnostics.AddError(
					"API Error - Retryable",
					fmt.Sprintf("Received retryable response code: %d, Response: %s", response.StatusCode, string(response.Body)),
				)
				return
			}
			fallthrough
		default: // "fail"
			resp.Diagnostics.AddError(
				"API Error",
				fmt.Sprintf("Received non-success response code: %d, Response: %s", response.StatusCode, string(response.Body)),
			)
			return
		}
	} else if action == "stop" {
		tflog.Info(ctx, "stopping processing due to success action", map[string]interface{}{
			"status_code": response.StatusCode,
		})
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

	// Resolve the HTTP method for read operation
	method := r.resolveMethodForOperation(&data, "read")

	// Set computed value for visibility
	data.ReadMethod = types.StringValue(method)

	// Build request options with resolved method
	options := r.buildRequestOptions(ctx, &data, method, "")
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

	// Perform drift detection if enabled
	if err := r.performDriftDetection(ctx, &data, response); err != nil {
		tflog.Warn(ctx, "drift detection warning", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue execution - drift detection errors are warnings, not failures
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

	// Resolve the HTTP method for update operation
	method := r.resolveMethodForOperation(&data, "update")

	// Set computed value for visibility
	data.UpdateMethod = types.StringValue(method)

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

	// Resolve the HTTP method for delete operation
	method := r.resolveMethodForOperation(&data, "delete")

	// Set computed value for visibility
	data.DeleteMethod = types.StringValue(method)

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

// performDriftDetection compares the current API response with the expected state
// to detect configuration drift while ignoring server-side metadata fields
func (r *RestResource) performDriftDetection(ctx context.Context, data *RestResourceModel, response *client.Response) error {
	// Skip drift detection if disabled
	if !data.DriftDetection.IsNull() && !data.DriftDetection.ValueBool() {
		return nil
	}

	// Default to enabled if not specified
	driftEnabled := true
	if !data.DriftDetection.IsNull() {
		driftEnabled = data.DriftDetection.ValueBool()
	}
	if !driftEnabled {
		return nil
	}

	// Parse the current response body
	if len(response.Body) == 0 {
		return nil
	}

	var currentData map[string]interface{}
	if err := json.Unmarshal(response.Body, &currentData); err != nil {
		// Not JSON or malformed - skip drift detection
		return nil
	}

	// Get the list of fields to ignore during drift detection
	ignoreFields := make(map[string]bool)

	// Add default fields that are commonly server-managed
	defaultIgnoreFields := []string{
		"id", "created_at", "updated_at", "last_modified", "etag",
		"version", "revision", "timestamp", "last_updated_at",
		"created_by", "updated_by", "modified_by", "owner_id",
		"_id", "_created", "_updated", "_modified", "_version",
		"createdAt", "updatedAt", "lastModified", "lastUpdated",
		"href", "self", "links", "_links", "meta", "_meta",
	}

	for _, field := range defaultIgnoreFields {
		ignoreFields[field] = true
	}

	// Add user-specified ignore fields
	if !data.IgnoreFields.IsNull() {
		userIgnoreFields := make([]string, 0, len(data.IgnoreFields.Elements()))
		diags := data.IgnoreFields.ElementsAs(ctx, &userIgnoreFields, false)
		if !diags.HasError() {
			for _, field := range userIgnoreFields {
				ignoreFields[field] = true
			}
		}
	}

	// Parse the expected response body if we have one
	var expectedData map[string]interface{}
	if !data.Body.IsNull() && data.Body.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.Body.ValueString()), &expectedData); err != nil {
			// Expected body is not JSON - compare raw strings
			return r.compareRawResponse(ctx, data, string(response.Body), ignoreFields)
		}
	} else {
		// No expected body to compare against - just update computed fields
		return r.updateComputedFields(ctx, data, currentData, ignoreFields)
	}

	// Compare the structured data
	driftDetected := r.compareStructuredData(ctx, expectedData, currentData, ignoreFields, "")

	if driftDetected {
		tflog.Warn(ctx, "configuration drift detected", map[string]interface{}{
			"resource_id": data.Id.ValueString(),
			"endpoint":    data.Endpoint.ValueString(),
		})
	}

	// Always update computed fields regardless of drift
	return r.updateComputedFields(ctx, data, currentData, ignoreFields)
}

// compareStructuredData recursively compares two JSON structures, ignoring specified fields
func (r *RestResource) compareStructuredData(ctx context.Context, expected, current map[string]interface{}, ignoreFields map[string]bool, path string) bool {
	driftDetected := false

	// Check for missing or changed fields in expected data
	for key, expectedValue := range expected {
		fieldPath := key
		if path != "" {
			fieldPath = path + "." + key
		}

		// Skip ignored fields
		if ignoreFields[key] {
			continue
		}

		currentValue, exists := current[key]
		if !exists {
			tflog.Debug(ctx, "field missing in current response", map[string]interface{}{
				"field":    fieldPath,
				"expected": expectedValue,
			})
			driftDetected = true
			continue
		}

		// Compare values based on type
		if !r.valuesEqual(expectedValue, currentValue) {
			tflog.Debug(ctx, "field value drift detected", map[string]interface{}{
				"field":    fieldPath,
				"expected": expectedValue,
				"current":  currentValue,
			})
			driftDetected = true
		}

		// Recursively check nested objects
		if expectedMap, ok := expectedValue.(map[string]interface{}); ok {
			if currentMap, ok := currentValue.(map[string]interface{}); ok {
				if r.compareStructuredData(ctx, expectedMap, currentMap, ignoreFields, fieldPath) {
					driftDetected = true
				}
			}
		}
	}

	return driftDetected
}

// compareRawResponse compares raw string responses
func (r *RestResource) compareRawResponse(ctx context.Context, data *RestResourceModel, currentResponse string, ignoreFields map[string]bool) error {
	expectedResponse := data.Body.ValueString()

	if expectedResponse != currentResponse {
		tflog.Debug(ctx, "raw response drift detected", map[string]interface{}{
			"expected_length": len(expectedResponse),
			"current_length":  len(currentResponse),
		})
	}

	return nil
}

// updateComputedFields updates computed fields in the state based on the current response
func (r *RestResource) updateComputedFields(ctx context.Context, data *RestResourceModel, currentData map[string]interface{}, ignoreFields map[string]bool) error {
	// Update the ID if it exists in the response and is different
	if currentId, ok := currentData["id"]; ok {
		if idStr, ok := currentId.(string); ok {
			expectedId := data.Id.ValueString()
			if idStr != expectedId && expectedId != "" {
				tflog.Debug(ctx, "updating resource ID from response", map[string]interface{}{
					"old_id": expectedId,
					"new_id": idStr,
				})
				data.Id = types.StringValue(idStr)
			} else if expectedId == "" {
				// Set ID if it wasn't set before
				data.Id = types.StringValue(idStr)
			}
		}
	}

	// Set drift_detection to true if not explicitly set
	if data.DriftDetection.IsNull() {
		data.DriftDetection = types.BoolValue(true)
	}

	return nil
}

// valuesEqual compares two interface{} values for equality, handling different JSON number types
func (r *RestResource) valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle numeric comparisons (JSON unmarshalling can create float64 or int)
	if aNum, aOk := a.(float64); aOk {
		if bNum, bOk := b.(float64); bOk {
			return aNum == bNum
		}
		if bNum, bOk := b.(int); bOk {
			return aNum == float64(bNum)
		}
	}
	if aNum, aOk := a.(int); aOk {
		if bNum, bOk := b.(float64); bOk {
			return float64(aNum) == bNum
		}
		if bNum, bOk := b.(int); bOk {
			return aNum == bNum
		}
	}

	// Direct comparison for other types
	return a == b
}

// resolveMethodForOperation determines the HTTP method to use for a given operation
// Supports backward compatibility with the legacy 'method' field
func (r *RestResource) resolveMethodForOperation(data *RestResourceModel, operation string) string {
	switch operation {
	case "create":
		if !data.CreateMethod.IsNull() && !data.CreateMethod.IsUnknown() {
			return data.CreateMethod.ValueString()
		}
		// Backward compatibility: use legacy method field for create
		if !data.Method.IsNull() && !data.Method.IsUnknown() {
			return data.Method.ValueString()
		}
		return "POST" // Default for create

	case "read":
		if !data.ReadMethod.IsNull() && !data.ReadMethod.IsUnknown() {
			return data.ReadMethod.ValueString()
		}
		// Backward compatibility: some APIs might use the same method for all operations
		if !data.Method.IsNull() && !data.Method.IsUnknown() {
			method := data.Method.ValueString()
			// Only use legacy method for read if it's a valid read method
			if method == "GET" || method == "POST" || method == "HEAD" {
				return method
			}
		}
		return "GET" // Default for read

	case "update":
		if !data.UpdateMethod.IsNull() && !data.UpdateMethod.IsUnknown() {
			return data.UpdateMethod.ValueString()
		}
		// Backward compatibility: use PATCH if legacy method was PATCH, otherwise PUT
		if !data.Method.IsNull() && !data.Method.IsUnknown() {
			method := data.Method.ValueString()
			if method == "PATCH" {
				return "PATCH"
			}
		}
		return "PUT" // Default for update

	case "delete":
		if !data.DeleteMethod.IsNull() && !data.DeleteMethod.IsUnknown() {
			return data.DeleteMethod.ValueString()
		}
		// Backward compatibility: use legacy method if it's a valid delete method
		if !data.Method.IsNull() && !data.Method.IsUnknown() {
			method := data.Method.ValueString()
			if method == "DELETE" || method == "POST" || method == "PUT" {
				return method
			}
		}
		return "DELETE" // Default for delete

	default:
		return "GET" // Safe default
	}
}
