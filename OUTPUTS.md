# Outputs in terraform-provider-rest

This provider includes comprehensive computed outputs that provide detailed information about API responses, making it easy to use REST API data in your Terraform configurations.

## Available Outputs

### Core Response Outputs

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | The resource identifier (extracted from API response or generated) |
| `response` | `string` | The complete raw response body from the API |
| `status_code` | `number` | The HTTP status code returned by the API |
| `response_headers` | `map(string)` | HTTP response headers as key-value pairs |
| `response_data` | `map(string)` | Parsed JSON response data as key-value pairs |
| `created_at` | `string` | RFC3339 timestamp when the resource was created |
| `last_updated` | `string` | RFC3339 timestamp when the resource was last updated |

### Response Data Processing

The `response_data` output automatically parses JSON responses and converts all values to strings for consistent usage:

- **Strings**: Used as-is
- **Numbers**: Converted to string representation (e.g., `42` → `"42"`)
- **Booleans**: Converted to string representation (e.g., `true` → `"true"`)
- **Objects/Arrays**: Converted to JSON string representation

## Usage Examples

### Basic Output Usage

```hcl
resource "rest_resource" "user" {
  endpoint = "/users"
  name     = "john-doe"
  
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
    role  = "admin"
  })
}

# Access the outputs
output "user_id" {
  description = "The ID of the created user"
  value       = rest_resource.user.id
}

output "user_response" {
  description = "Complete API response"
  value       = rest_resource.user.response
}

output "user_status_code" {
  description = "HTTP status code"
  value       = rest_resource.user.status_code
}

output "user_created_at" {
  description = "When the user was created"
  value       = rest_resource.user.created_at
}
```

### Using Response Data

```hcl
resource "rest_resource" "api_key" {
  endpoint = "/api-keys"
  name     = "service-key"
  
  body = jsonencode({
    name        = "Service API Key"
    permissions = ["read", "write"]
    expires_in  = 3600
  })
}

# Extract specific values from response_data
output "api_key_token" {
  description = "The generated API key token"
  value       = rest_resource.api_key.response_data["token"]
  sensitive   = true
}

output "api_key_expires_at" {
  description = "When the API key expires"
  value       = rest_resource.api_key.response_data["expires_at"]
}

output "api_key_permissions" {
  description = "Assigned permissions (JSON string)"
  value       = rest_resource.api_key.response_data["permissions"]
}
```

### Using Response Headers

```hcl
resource "rest_resource" "deployment" {
  endpoint = "/deployments"
  name     = "app-v1"
  
  body = jsonencode({
    image    = "app:v1.0"
    replicas = 3
  })
}

output "deployment_location" {
  description = "Location header from deployment response"
  value       = rest_resource.deployment.response_headers["Location"]
}

output "deployment_etag" {
  description = "ETag header for caching"
  value       = rest_resource.deployment.response_headers["ETag"]
}
```

### Chaining Resources with Outputs

```hcl
# Create a project
resource "rest_resource" "project" {
  endpoint = "/projects"
  name     = "my-project"
  
  body = jsonencode({
    name        = "My Project"
    description = "A sample project"
  })
}

# Create a database using the project ID
resource "rest_resource" "database" {
  endpoint = "/databases"
  name     = "project-db"
  
  body = jsonencode({
    name       = "project-database"
    project_id = rest_resource.project.id
    engine     = "postgresql"
    version    = "13"
  })
}

# Create a user for the database
resource "rest_resource" "db_user" {
  endpoint = "/database-users"
  name     = "app-user"
  
  body = jsonencode({
    username    = "app"
    database_id = rest_resource.database.id
    permissions = ["SELECT", "INSERT", "UPDATE"]
  })
}

# Output the connection details
output "database_connection" {
  description = "Database connection information"
  value = {
    host     = rest_resource.database.response_data["host"]
    port     = rest_resource.database.response_data["port"]
    database = rest_resource.database.response_data["name"]
    username = rest_resource.db_user.response_data["username"]
    password = rest_resource.db_user.response_data["password"]
  }
  sensitive = true
}
```

### Complex Response Processing

```hcl
resource "rest_resource" "server" {
  endpoint = "/servers"
  name     = "web-server-1"
  
  body = jsonencode({
    name         = "Web Server 1"
    instance_type = "t3.medium"
    region       = "us-east-1"
  })
}

# Parse complex nested response data
locals {
  server_config = jsondecode(rest_resource.server.response_data["config"])
  server_tags   = jsondecode(rest_resource.server.response_data["tags"])
}

output "server_details" {
  description = "Processed server information"
  value = {
    id         = rest_resource.server.id
    public_ip  = rest_resource.server.response_data["public_ip"]
    private_ip = rest_resource.server.response_data["private_ip"]
    status     = rest_resource.server.response_data["status"]
    cpu_count  = local.server_config["cpu_count"]
    memory_gb  = local.server_config["memory_gb"]
    environment = local.server_tags["environment"]
    owner      = local.server_tags["owner"]
  }
}
```

### Conditional Logic with Outputs

```hcl
resource "rest_resource" "service" {
  endpoint = "/services"
  name     = "web-service"
  
  body = jsonencode({
    name = "Web Service"
    type = "web"
  })
}

# Conditional outputs based on response
output "service_url" {
  description = "Service URL (if available)"
  value = try(
    rest_resource.service.response_data["url"],
    "URL not available"
  )
}

output "service_health_check" {
  description = "Health check endpoint"
  value = format("%s/health", 
    rest_resource.service.response_data["base_url"]
  )
}

# Create additional resources conditionally
resource "rest_resource" "ssl_cert" {
  count = rest_resource.service.response_data["ssl_enabled"] == "true" ? 1 : 0
  
  endpoint = "/ssl-certificates"
  name     = "web-service-cert"
  
  body = jsonencode({
    service_id = rest_resource.service.id
    domain     = rest_resource.service.response_data["domain"]
  })
}
```

## Best Practices

### 1. Use Type Conversion Functions

Since `response_data` values are strings, convert them when needed:

```hcl
locals {
  port_number = tonumber(rest_resource.server.response_data["port"])
  is_enabled  = tobool(rest_resource.server.response_data["enabled"])
}
```

### 2. Handle Missing Values

Use `try()` or `lookup()` for safe access:

```hcl
output "optional_field" {
  value = try(rest_resource.api.response_data["optional_field"], "default_value")
}
```

### 3. Parse JSON Strings

For complex nested data:

```hcl
locals {
  metadata = jsondecode(rest_resource.service.response_data["metadata"])
}

output "parsed_metadata" {
  value = local.metadata
}
```

### 4. Use Sensitive Outputs

Mark sensitive data appropriately:

```hcl
output "api_secret" {
  value     = rest_resource.api_key.response_data["secret"]
  sensitive = true
}
```

### 5. Document Output Usage

Always provide clear descriptions:

```hcl
output "database_connection_string" {
  description = "PostgreSQL connection string for the application database"
  value = format("postgresql://%s:%s@%s:%s/%s",
    rest_resource.db_user.response_data["username"],
    rest_resource.db_user.response_data["password"],
    rest_resource.database.response_data["host"],
    rest_resource.database.response_data["port"],
    rest_resource.database.response_data["name"]
  )
  sensitive = true
}
```

## Common Patterns

### Service Discovery

```hcl
# Register a service
resource "rest_resource" "service_registration" {
  endpoint = "/service-registry"
  name     = "api-service"
  
  body = jsonencode({
    name = "API Service"
    port = 8080
  })
}

# Use discovered endpoints
output "service_endpoints" {
  description = "Discovered service endpoints"
  value = {
    health_check = "${rest_resource.service_registration.response_data["base_url"]}/health"
    metrics      = "${rest_resource.service_registration.response_data["base_url"]}/metrics"
    api_docs     = "${rest_resource.service_registration.response_data["base_url"]}/docs"
  }
}
```

### Configuration Management

```hcl
# Create configuration
resource "rest_resource" "app_config" {
  endpoint = "/configurations"
  name     = "app-settings"
  
  body = jsonencode({
    app_name     = "My Application"
    environment  = "production"
    feature_flags = {
      new_ui      = true
      beta_features = false
    }
  })
}

# Extract computed configuration
output "runtime_config" {
  description = "Runtime configuration for the application"
  value = {
    config_id      = rest_resource.app_config.id
    config_version = rest_resource.app_config.response_data["version"]
    cache_ttl      = rest_resource.app_config.response_data["cache_ttl"]
    last_updated   = rest_resource.app_config.last_updated
  }
}
```

## Troubleshooting

### Debug Output Values

Enable debug logging to see output processing:

```bash
export TF_LOG=DEBUG
terraform plan
```

### Check Response Structure

Use the raw response output to understand the API response format:

```hcl
output "debug_response" {
  value = rest_resource.api.response
}
```

### Validate Data Types

Use Terraform's type functions to validate and convert:

```hcl
output "validated_number" {
  value = can(tonumber(rest_resource.api.response_data["count"])) ? 
    tonumber(rest_resource.api.response_data["count"]) : 0
}
```

This comprehensive output system makes terraform-provider-rest highly versatile for integrating REST APIs into your Terraform infrastructure workflows.