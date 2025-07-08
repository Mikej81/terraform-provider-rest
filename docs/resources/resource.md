---
page_title: "rest_resource Resource - rest"
subcategory: ""
description: |-
  Manages REST API resources with full CRUD operations, dynamic response parsing, and import support.
---

# rest_resource (Resource)

The `rest_resource` provides comprehensive REST API resource management with support for all HTTP methods, custom headers, query parameters, and dynamic response data access.

## Features

- **Full CRUD Operations**: Complete create, read, update, delete lifecycle management
- **HTTP Method Support**: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- **Dynamic Response Access**: Automatic JSON parsing with accessible key-value outputs
- **Import Support**: Import existing API resources into Terraform state
- **Drift Detection**: Automatic detection of external changes
- **Custom Bodies**: Separate request bodies for create, update, and delete operations

## Example Usage

### Basic Resource

```terraform
resource "rest_resource" "user" {
  name     = "john-doe"
  endpoint = "/api/users"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
    role  = "admin"
  })
}

# Access response data
output "user_id" {
  value = rest_resource.user.response_data.id
}
```

### Advanced Resource with Custom Operations

```terraform
resource "rest_resource" "api_config" {
  name     = "production-config"
  endpoint = "/api/configurations"
  method   = "POST"
  
  headers = {
    "Content-Type"    = "application/json"
    "X-Environment"   = "production"
  }
  
  query_params = {
    "validate" = "true"
    "backup"   = "true"
  }
  
  # Create body
  body = jsonencode({
    name = "production-config"
    settings = {
      timeout = 30
      debug   = false
    }
  })
  
  # Update body (optional, different from create)
  update_body = jsonencode({
    name = "production-config"
    settings = {
      timeout = 30
      debug   = false
      version = "2.0"
    }
  })
  
  # Delete body (optional)
  destroy_body = jsonencode({
    force = true
    backup = false
  })
  
  timeout        = 60
  retry_attempts = 5
}
```

### Import Existing Resources

```bash
# Import using endpoint/name format
terraform import rest_resource.existing_user "api/users/existing-user-id"
```

## Schema

### Required

- `endpoint` (String) The API endpoint to send the request to (relative to the base URL)
- `name` (String) The name of the item used for identification during read, update, and delete operations

### Optional

- `method` (String) The HTTP method for create operations (POST, PUT, PATCH). Default: POST
- `headers` (Map of String) Custom HTTP headers to include in requests
- `query_params` (Map of String) URL query parameters to include in requests
- `body` (String) The request body for create operations (JSON or any payload format)
- `update_body` (String) Optional separate request body for update operations
- `destroy_body` (String) Optional request body for delete operations
- `timeout` (Number) Timeout for requests in seconds. Overrides provider default
- `retry_attempts` (Number) Number of retry attempts for requests. Overrides provider default
- `insecure` (Boolean) Disable SSL certificate verification. Overrides provider default

### Read-Only (Computed)

- `id` (String) The identifier for the created resource
- `response` (String) The raw response body from the most recent API request
- `status_code` (Number) The HTTP status code from the most recent API request
- `response_data` (Map of String) Parsed JSON response as key-value pairs for dynamic access
- `response_headers` (Map of String) HTTP response headers as key-value pairs
- `created_at` (String) Timestamp when the resource was created
- `last_updated` (String) Timestamp when the resource was last updated

## Accessing Dynamic Response Data

The `response_data` attribute automatically parses JSON responses, making individual fields accessible:

```terraform
resource "rest_resource" "api_call" {
  # ... configuration ...
}

# Access specific response fields
output "resource_id" {
  value = rest_resource.api_call.response_data.id
}

output "resource_status" {
  value = rest_resource.api_call.response_data.status
}

output "nested_field" {
  value = rest_resource.api_call.response_data.metadata
}
```

## Import

Resources can be imported using the format `endpoint/name`:

```bash
terraform import rest_resource.example "api/v1/users/user-123"
```

This will import the resource with:
- `endpoint` = "/api/v1/users"  
- `name` = "user-123"
