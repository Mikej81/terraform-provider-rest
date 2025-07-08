---
page_title: "rest_data Data Source - rest"
subcategory: ""
description: |-
  Retrieves data from REST API endpoints with support for all HTTP methods, custom headers, and dynamic response parsing.
---

# rest_data (Data Source)

The `rest_data` data source enables retrieval of data from REST API endpoints with full HTTP method support, custom headers, query parameters, and automatic JSON response parsing.

## Features

- **All HTTP Methods**: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS support
- **Dynamic Response Parsing**: Automatic JSON parsing with accessible key-value outputs
- **Custom Headers**: Add custom HTTP headers for authentication or API requirements
- **Query Parameters**: URL query parameter support
- **Flexible Configuration**: Override provider defaults for timeout, retry, and SSL settings

## Example Usage

### Basic GET Request

```terraform
data "rest_data" "user_info" {
  endpoint = "/api/users/123"
  method   = "GET"
}

output "user_name" {
  value = data.rest_data.user_info.response_data.name
}
```

### POST Request with Body

```terraform
data "rest_data" "search_results" {
  endpoint = "/api/search"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
    "Accept"       = "application/json"
  }
  
  body = jsonencode({
    query = "terraform"
    limit = 10
    filters = {
      category = "infrastructure"
    }
  })
}

output "search_count" {
  value = data.rest_data.search_results.response_data.total_count
}
```

### Request with Custom Headers and Query Parameters

```terraform
data "rest_data" "api_status" {
  endpoint = "/api/health"
  method   = "GET"
  
  headers = {
    "Accept"         = "application/json"
    "X-API-Version"  = "v2"
  }
  
  query_params = {
    "include" = "metrics"
    "format"  = "json"
  }
  
  timeout        = 30
  retry_attempts = 3
}

output "api_health" {
  value = data.rest_data.api_status.response_data.status
}
```

## Schema

### Required

- `endpoint` (String) The API endpoint to send the request to (relative to the base URL)

### Optional

- `method` (String) The HTTP method to use (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS). Default: GET
- `headers` (Map of String) Custom HTTP headers to include in the request
- `query_params` (Map of String) URL query parameters to include in the request
- `body` (String) The request body (for POST, PUT, PATCH methods)
- `timeout` (Number) Timeout for the request in seconds. Overrides provider default
- `retry_attempts` (Number) Number of retry attempts for the request. Overrides provider default  
- `insecure` (Boolean) Disable SSL certificate verification. Overrides provider default

### Read-Only (Computed)

- `id` (String) The identifier for the request
- `response` (String) The raw response body from the API request
- `status_code` (Number) The HTTP status code from the API request
- `response_data` (Map of String) Parsed JSON response as key-value pairs for dynamic access
- `response_headers` (Map of String) HTTP response headers as key-value pairs

## Accessing Response Data

The `response_data` attribute automatically parses JSON responses, making individual fields accessible:

```terraform
data "rest_data" "api_info" {
  endpoint = "/api/info"
}

# Access nested response data
output "api_version" {
  value = data.rest_data.api_info.response_data.version
}

output "api_features" {
  value = data.rest_data.api_info.response_data.features
}

# Access response headers
output "rate_limit_remaining" {
  value = data.rest_data.api_info.response_headers["X-RateLimit-Remaining"]
}
```
