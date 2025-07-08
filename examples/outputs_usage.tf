# Example: Comprehensive output usage with terraform-provider-rest
# This demonstrates all available outputs and common usage patterns

terraform {
  required_providers {
    rest = {
      source  = "registry.terraform.io/your-org/rest"
      version = "1.0.8"
    }
  }
}

provider "rest" {
  base_url = "https://api.example.com"
  headers = {
    "Authorization" = "Bearer ${var.api_token}"
    "Content-Type"  = "application/json"
  }
}

# Example 1: Basic output usage
resource "rest_resource" "user" {
  endpoint = "/users"
  name     = "john-doe"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
    role  = "admin"
  })
}

# Basic outputs
output "user_id" {
  description = "The ID of the created user"
  value       = rest_resource.user.id
}

output "user_status_code" {
  description = "HTTP status code from user creation"
  value       = rest_resource.user.status_code
}

output "user_created_at" {
  description = "When the user was created"
  value       = rest_resource.user.created_at
}

output "user_last_updated" {
  description = "When the user was last updated"
  value       = rest_resource.user.last_updated
}

# Example 2: Response data extraction
resource "rest_resource" "api_key" {
  endpoint = "/api-keys"
  name     = "service-key"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name        = "Service API Key"
    permissions = ["read", "write"]
    expires_in  = 3600
  })
}

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

# Example 3: Response headers usage
resource "rest_resource" "deployment" {
  endpoint = "/deployments"
  name     = "app-v1"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
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

output "deployment_server" {
  description = "Server header information"
  value       = rest_resource.deployment.response_headers["Server"]
}

# Example 4: Resource chaining with outputs
resource "rest_resource" "project" {
  endpoint = "/projects"
  name     = "my-project"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name        = "My Project"
    description = "A sample project"
  })
}

resource "rest_resource" "database" {
  endpoint = "/databases"
  name     = "project-db"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name       = "project-database"
    project_id = rest_resource.project.id
    engine     = "postgresql"
    version    = "13"
  })
}

resource "rest_resource" "db_user" {
  endpoint = "/database-users"
  name     = "app-user"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    username    = "app"
    database_id = rest_resource.database.id
    permissions = ["SELECT", "INSERT", "UPDATE"]
  })
}

# Chained outputs
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

# Example 5: Complex response processing
resource "rest_resource" "server" {
  endpoint = "/servers"
  name     = "web-server-1"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
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

# Example 6: Conditional logic with outputs
resource "rest_resource" "service" {
  endpoint = "/services"
  name     = "web-service"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name = "Web Service"
    type = "web"
  })
}

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

# Conditional resource creation
resource "rest_resource" "ssl_cert" {
  count = rest_resource.service.response_data["ssl_enabled"] == "true" ? 1 : 0
  
  endpoint = "/ssl-certificates"
  name     = "web-service-cert"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    service_id = rest_resource.service.id
    domain     = rest_resource.service.response_data["domain"]
  })
}

# Example 7: Service discovery pattern
resource "rest_resource" "service_registration" {
  endpoint = "/service-registry"
  name     = "api-service"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name = "API Service"
    port = 8080
  })
}

output "service_endpoints" {
  description = "Discovered service endpoints"
  value = {
    health_check = "${rest_resource.service_registration.response_data["base_url"]}/health"
    metrics      = "${rest_resource.service_registration.response_data["base_url"]}/metrics"
    api_docs     = "${rest_resource.service_registration.response_data["base_url"]}/docs"
  }
}

# Example 8: Configuration management
resource "rest_resource" "app_config" {
  endpoint = "/configurations"
  name     = "app-settings"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    app_name     = "My Application"
    environment  = "production"
    feature_flags = {
      new_ui      = true
      beta_features = false
    }
  })
}

output "runtime_config" {
  description = "Runtime configuration for the application"
  value = {
    config_id      = rest_resource.app_config.id
    config_version = rest_resource.app_config.response_data["version"]
    cache_ttl      = rest_resource.app_config.response_data["cache_ttl"]
    last_updated   = rest_resource.app_config.last_updated
  }
}

# Example 9: Type conversion and validation
resource "rest_resource" "metrics" {
  endpoint = "/metrics"
  name     = "app-metrics"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    metric_name = "cpu_usage"
    threshold   = 80
  })
}

locals {
  # Type conversions
  threshold_number = tonumber(rest_resource.metrics.response_data["threshold"])
  is_enabled       = tobool(rest_resource.metrics.response_data["enabled"])
  
  # Safe access with defaults
  sample_rate = try(
    tonumber(rest_resource.metrics.response_data["sample_rate"]),
    1.0
  )
}

output "metric_configuration" {
  description = "Processed metric configuration"
  value = {
    threshold    = local.threshold_number
    enabled      = local.is_enabled
    sample_rate  = local.sample_rate
    alert_url    = rest_resource.metrics.response_data["alert_url"]
  }
}

# Example 10: Error handling and debugging
resource "rest_resource" "debug_example" {
  endpoint = "/debug-endpoint"
  name     = "debug-test"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    test_field = "test_value"
  })
}

# Debug outputs
output "debug_raw_response" {
  description = "Raw response for debugging"
  value       = rest_resource.debug_example.response
}

output "debug_all_headers" {
  description = "All response headers"
  value       = rest_resource.debug_example.response_headers
}

output "debug_all_data" {
  description = "All parsed response data"
  value       = rest_resource.debug_example.response_data
}

# Validation example
output "validated_number" {
  description = "Validated numeric field"
  value = can(tonumber(rest_resource.debug_example.response_data["count"])) ? tonumber(rest_resource.debug_example.response_data["count"]) : 0
}

# Example 11: Multi-step workflow
resource "rest_resource" "workflow_step1" {
  endpoint = "/workflows"
  name     = "data-processing"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    type = "data_processing"
    input_source = "s3://my-bucket/data/"
  })
}

resource "rest_resource" "workflow_step2" {
  endpoint = "/workflows/${rest_resource.workflow_step1.id}/steps"
  name     = "validation-step"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    step_type = "validation"
    depends_on = rest_resource.workflow_step1.id
  })
}

output "workflow_status" {
  description = "Complete workflow status"
  value = {
    workflow_id = rest_resource.workflow_step1.id
    step1_status = rest_resource.workflow_step1.response_data["status"]
    step2_status = rest_resource.workflow_step2.response_data["status"]
    step1_created = rest_resource.workflow_step1.created_at
    step2_created = rest_resource.workflow_step2.created_at
    total_duration = rest_resource.workflow_step2.response_data["total_duration"]
  }
}