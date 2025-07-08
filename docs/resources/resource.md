---
page_title: "rest_resource Resource - rest"
subcategory: ""
description: |-
  Manages REST API resources with full CRUD operations, dynamic response parsing, and import support.
---

# rest_resource (Resource)

The `rest_resource` is your main tool for creating and managing resources through REST APIs. Think of it as a way to treat any API endpoint as a Terraform resource with full lifecycle management.

**What this means**: Instead of making API calls manually, you declare what you want, and Terraform handles the create, read, update, and delete operations for you.

## What Can This Resource Do?

### **Complete Lifecycle Management**

- **Create**: `terraform apply` calls your API to create new resources
- **Read**: Terraform periodically checks if resources still exist and match your configuration
- **Update**: Changes to your configuration trigger API calls to update resources
- **Delete**: `terraform destroy` removes resources through your API

### **Universal API Support**

- Works with any HTTP method your API supports
- Handles custom headers (authentication, content types, etc.)
- Supports query parameters for APIs that need them
- Automatic JSON response parsing so you can use response data

### **Smart Change Detection**

- Detects when someone changes resources outside of Terraform
- Ignores server-added metadata (timestamps, IDs) that shouldn't trigger changes
- Keeps your Terraform state in sync with reality

### **Production Features**

- Import existing resources without recreating them
- Use different request bodies for create, update, and delete operations
- Override provider settings per resource (timeouts, retries, etc.)

## Example Usage

### **Simple Resource Creation**

```terraform
resource "rest_resource" "user" {
  name     = "john-doe"        # Unique identifier for this resource
  endpoint = "/api/users"       # API endpoint to call
  method   = "POST"             # HTTP method for creation
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  # The data to send when creating the user
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
    role  = "admin"
  })
}

# Use the API response in other resources or outputs
output "user_details" {
  value = {
    id       = rest_resource.user.response_data.id
    name     = rest_resource.user.response_data.name
    api_url  = "https://api.example.com/users/${rest_resource.user.response_data.id}"
  }
}
```

**What happens when you run `terraform apply`**:

1. Terraform calls `POST /api/users` with your JSON body
2. Your API creates the user and returns user data (including an ID)
3. Terraform stores the response in state
4. You can reference the response data in other resources

**What happens on subsequent runs**:

1. Terraform calls `GET /api/users/{name}` to check if the user still exists
2. If it exists and matches your configuration, no changes are made
3. If it's changed, Terraform will update it
4. If it's deleted, Terraform will recreate it

### **Advanced Resource Configuration**

```terraform
resource "rest_resource" "api_config" {
  name     = "production-config"
  endpoint = "/api/configurations"
  method   = "POST"
  
  # Custom headers for your API's requirements
  headers = {
    "Content-Type"    = "application/json"
    "X-Environment"   = "production"
    "X-Request-ID"    = "config-${random_uuid.config_id.result}"
  }
  
  # Query parameters are added to the URL
  query_params = {
    "validate" = "true"     # Validates config before creating
    "backup"   = "true"     # Creates backup of old config
    "notify"   = "false"    # Disables notifications
  }
  
  # Request body for creating the configuration
  body = jsonencode({
    name = "production-config"
    settings = {
      timeout = 30
      debug   = false
      region  = "us-east-1"
    }
  })
  
  # Different body for updates (optional)
  # Use this when update operations need different data
  update_body = jsonencode({
    name = "production-config"
    settings = {
      timeout = 30
      debug   = false
      region  = "us-east-1"
      version = "2.0"        # Only set on updates
    }
  })
  
  # Body for delete operations (optional)
  # Some APIs need data when deleting resources
  destroy_body = jsonencode({
    force  = true
    backup = false
  })
  
  # Override provider defaults for this resource
  timeout        = 60   # Wait longer for config operations
  retry_attempts = 5    # Retry more for critical configs
}

# Generate a unique ID for this configuration
resource "random_uuid" "config_id" {}

# Create a monitoring rule for this configuration
resource "rest_resource" "config_monitor" {
  name     = "monitor-${rest_resource.api_config.response_data.id}"
  endpoint = "/api/monitors"
  
  body = jsonencode({
    config_id = rest_resource.api_config.response_data.id
    type      = "config_check"
    interval  = 300  # Check every 5 minutes
  })
  
  # Only create monitor after config is created
  depends_on = [rest_resource.api_config]
}
```

**Key concepts explained**:

- **`name`**: How Terraform identifies this resource for read/update/delete operations
- **`update_body`**: Use when your API needs different data for updates vs creates
- **`destroy_body`**: Use when your API needs data to delete resources (like force flags)
- **`depends_on`**: Ensures resources are created in the right order

### **Importing Existing Resources**

Bring existing API resources under Terraform management without recreating them:

```bash
# Import a user that already exists in your API
terraform import rest_resource.existing_user "api/users/existing-user-id"

# Import a configuration
terraform import rest_resource.prod_config "api/configurations/prod-config"
```

**After importing, create the Terraform configuration**:

```terraform
resource "rest_resource" "existing_user" {
  name     = "existing-user-id"  # Must match the imported name
  endpoint = "/api/users"
  
  # This body should match the current state of the resource
  body = jsonencode({
    name  = "Existing User"
    email = "existing@example.com"
    role  = "user"
  })
}
```

**Import workflow**:

1. Run `terraform import` to bring the resource into state
2. Run `terraform show` to see the current state
3. Write Terraform configuration that matches the current state
4. Run `terraform plan` to verify no changes are needed
5. Start managing the resource normally

> **Pro tip**: Use `terraform show` after importing to see what the current state looks like, then write your configuration to match.

## Configuration Reference

### Required Settings

**`endpoint`** (String)

- The API endpoint to send requests to (relative to the provider's `api_url`)
- Examples: `"/users"`, `"/api/v1/configurations"`, `"/projects/{project_id}/settings"`

**`name`** (String)

- Unique identifier for this resource within your API
- Used for read, update, and delete operations
- Should be stable and unique (like a username, slug, or ID)
- Examples: `"john-doe"`, `"prod-config"`, `"project-123"`

### Optional Settings

**`method`** (String)

- HTTP method for creating the resource
- Default: `"POST"`
- Common values: `"POST"`, `"PUT"`, `"PATCH"`

**`headers`** (Map of String)

- Custom HTTP headers to include in all requests
- Examples: `{"Content-Type" = "application/json", "X-Custom-Header" = "value"}`

**`query_params`** (Map of String)

- URL query parameters to include in all requests
- Examples: `{"validate" = "true", "format" = "json"}`

**`body`** (String)

- Request body for create operations (usually JSON)
- Use `jsonencode()` for JSON data
- Example: `jsonencode({name = "John", email = "john@example.com"})`

**`update_body`** (String)

- Optional separate request body for update operations
- Use when your API needs different data for updates vs creates
- If not specified, uses `body` for updates

**`destroy_body`** (String)

- Optional request body for delete operations
- Use when your API needs data to delete resources (like force flags)

**Performance/Reliability Settings** (Override provider defaults)

- **`timeout`** (Number) - Request timeout in seconds
- **`retry_attempts`** (Number) - Number of retry attempts
- **`insecure`** (Boolean) - Skip SSL certificate verification

### Response Data (Read-Only)

**`id`** (String)

- The unique identifier for the created resource
- Extracted from API response or generated automatically

**`response_data`** (Map of String)

- Parsed JSON response as key-value pairs
- **This is the most useful attribute** - access specific fields like `response_data.id`
- Example: If API returns `{"id": "123", "name": "John"}`, you can use `response_data.id` and `response_data.name`

**`response`** (String)

- Raw response body from the most recent API request
- Useful for debugging or when response isn't JSON

**`status_code`** (Number)

- HTTP status code from the most recent API request
- Useful for debugging or conditional logic

**`response_headers`** (Map of String)

- HTTP response headers as key-value pairs
- Useful for accessing pagination info, rate limits, etc.

**`created_at`** (String)

- RFC3339 timestamp when the resource was created in Terraform
- This is when Terraform created it, not when your API created it

**`last_updated`** (String)

- RFC3339 timestamp when the resource was last updated in Terraform

## Working with API Responses

### **Understanding `response_data`**

The `response_data` attribute is your key to accessing API response data in Terraform:

```terraform
resource "rest_resource" "user" {
  name     = "john-doe"
  endpoint = "/users"
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
  })
}

# If your API returns:
# {
#   "id": "user-123",
#   "name": "John Doe",
#   "email": "john@example.com",
#   "created_at": "2025-01-15T10:30:00Z",
#   "profile": {
#     "avatar_url": "https://example.com/avatars/john.jpg",
#     "bio": "Software Engineer"
#   }
# }

# You can access individual fields:
output "user_id" {
  value = rest_resource.user.response_data.id
  # Result: "user-123"
}

output "user_profile" {
  value = {
    avatar = rest_resource.user.response_data.profile.avatar_url
    bio    = rest_resource.user.response_data.profile.bio
  }
}
```

### **Chaining Resources**

Use response data from one resource to create another:

```terraform
# Create a project
resource "rest_resource" "project" {
  name     = "my-project"
  endpoint = "/projects"
  body = jsonencode({
    name        = "My Project"
    description = "A sample project"
  })
}

# Create a database for the project
resource "rest_resource" "database" {
  name     = "project-database"
  endpoint = "/databases"
  body = jsonencode({
    name       = "my-project-db"
    project_id = rest_resource.project.response_data.id  # Use the project ID
    size       = "small"
  })
}

# Create a database user
resource "rest_resource" "db_user" {
  name     = "app-user"
  endpoint = "/database-users"
  body = jsonencode({
    username    = "app_user"
    database_id = rest_resource.database.response_data.id
    permissions = ["SELECT", "INSERT", "UPDATE"]
  })
}

# Output connection information
output "database_connection" {
  value = {
    host     = rest_resource.database.response_data.host
    port     = rest_resource.database.response_data.port
    database = rest_resource.database.response_data.name
    username = rest_resource.db_user.response_data.username
  }
  sensitive = true
}
```

## Complete Example: User Management System

```terraform
# Configure the provider
provider "rest" {
  api_url   = "https://api.example.com"
  api_token = var.api_token
}

# Create a user
resource "rest_resource" "user" {
  name     = "john-doe"
  endpoint = "/users"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
    role  = "developer"
  })
}

# Create a project owned by the user
resource "rest_resource" "project" {
  name     = "johns-project"
  endpoint = "/projects"
  
  body = jsonencode({
    name        = "John's Project"
    description = "A sample project"
    owner_id    = rest_resource.user.response_data.id
  })
}

# Add the user to a team
resource "rest_resource" "team_membership" {
  name     = "john-team-membership"
  endpoint = "/team-memberships"
  
  body = jsonencode({
    user_id = rest_resource.user.response_data.id
    team_id = "team-developers"
    role    = "member"
  })
}

# Outputs for other modules or debugging
output "user_info" {
  value = {
    id       = rest_resource.user.response_data.id
    name     = rest_resource.user.response_data.name
    email    = rest_resource.user.response_data.email
    api_url  = "https://api.example.com/users/${rest_resource.user.response_data.id}"
  }
}

output "project_info" {
  value = {
    id       = rest_resource.project.response_data.id
    name     = rest_resource.project.response_data.name
    owner_id = rest_resource.project.response_data.owner_id
    url      = "https://app.example.com/projects/${rest_resource.project.response_data.id}"
  }
}
```

**What this example demonstrates**:

- Basic resource creation with API responses
- Resource chaining (using user ID in project creation)
- Proper output structuring for reusability
- Real-world resource relationships

## Troubleshooting Tips

### **Common Issues**

**Problem: Resource not found during reads**

```terraform
# Solution: Make sure your 'name' field matches what your API expects
resource "rest_resource" "user" {
  name     = "john-doe"  # This must match your API's identifier
  endpoint = "/users"
  # ...
}
```

**Problem: Authentication errors**

```terraform
# Solution: Check your provider configuration
provider "rest" {
  api_url    = "https://api.example.com"
  api_token  = var.api_token  # Make sure this variable is set
  api_header = "Authorization"  # Make sure this matches your API
}
```

**Problem: Timeout errors**

```terraform
# Solution: Increase timeout for slow operations
resource "rest_resource" "slow_operation" {
  name     = "slow-resource"
  endpoint = "/slow-endpoint"
  timeout  = 300  # 5 minutes
  # ...
}
```

### **Debugging**

```bash
# Enable debug logging to see HTTP requests/responses
export TF_LOG=DEBUG
terraform plan

# Look for lines containing "rest_resource" in the output
# You'll see the actual HTTP requests being made
```

### **Testing Your Configuration**

```terraform
# Start with a data source to test connectivity
data "rest_data" "health_check" {
  endpoint = "/health"
  method   = "GET"
}

output "api_status" {
  value = data.rest_data.health_check.response_data
}
```

Once the data source works, you know your authentication and connectivity are correct.

## Need More Help?

- **[Data Source Documentation](../data-sources/data.md)**: Learn about fetching data from APIs
- **[Provider Configuration](../index.md)**: Complete authentication and configuration guide
- **[Examples](../../examples/)**: Real-world usage patterns
- **[Drift Detection](../../DRIFT_DETECTION.md)**: Understanding automatic change detection

> **Pro tip**: Start simple with a basic resource, then gradually add features like custom headers, query parameters, and different request bodies as you become more comfortable with the provider.

## Common Patterns and Best Practices

### **Naming Resources**

```terraform
# Good: Descriptive and unique
resource "rest_resource" "prod_api_key" {
  name = "production-api-key-${var.service_name}"
  # ...
}

# Good: Use environment prefixes
resource "rest_resource" "user" {
  name = "${var.environment}-user-${var.username}"
  # ...
}

# Bad: Generic names that might conflict
resource "rest_resource" "key" {
  name = "key"
  # ...
}
```

### **Handling Different HTTP Methods**

```terraform
# Most APIs use POST for creation
resource "rest_resource" "post_example" {
  name     = "my-resource"
  endpoint = "/resources"
  method   = "POST"  # Default
  body     = jsonencode({data = "value"})
}

# Some APIs use PUT for creation
resource "rest_resource" "put_example" {
  name     = "my-resource"
  endpoint = "/resources"
  method   = "PUT"
  body     = jsonencode({data = "value"})
}

# PATCH for partial updates
resource "rest_resource" "patch_example" {
  name     = "my-resource"
  endpoint = "/resources"
  method   = "PATCH"
  body     = jsonencode({data = "value"})
}
```

### **Error Handling**

```terraform
# Increase retries for unreliable APIs
resource "rest_resource" "unreliable_api" {
  name           = "my-resource"
  endpoint       = "/resources"
  retry_attempts = 10
  timeout        = 120
  body           = jsonencode({data = "value"})
}

# Skip SSL verification for development
resource "rest_resource" "dev_api" {
  name     = "dev-resource"
  endpoint = "/resources"
  insecure = true  # Only for development!
  body     = jsonencode({data = "value"})
}
```

### **Import Format**

Import existing resources using the format `endpoint/name`:

```bash
terraform import rest_resource.example "api/v1/users/user-123"
```

**This imports**:

- `endpoint` = `"/api/v1/users"`
- `name` = `"user-123"`

**Then create matching configuration**:

```terraform
resource "rest_resource" "example" {
  name     = "user-123"
  endpoint = "/api/v1/users"
  body     = jsonencode({
    # Match the current state of the resource
  })
}
```

> **Remember**: After importing, your Terraform configuration must match the current state of the resource, or Terraform will try to "fix" it on the next apply.
