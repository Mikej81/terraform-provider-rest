# Token Authentication Example
# This example demonstrates how to use token-based authentication with the REST provider

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}

# Provider configuration with token authentication
provider "rest" {
  api_url    = "https://api.example.com"
  api_token  = var.api_token
  api_header = "X-API-Key" # Custom header name for the token
  
  timeout        = 30
  retry_attempts = 3
}

# Example: Create a user via REST API
resource "rest_resource" "user" {
  name     = "john-doe"
  endpoint = "/api/v1/users"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    username = "john.doe"
    email    = "john@example.com"
    role     = "user"
    active   = true
  })
}

# Example: Retrieve user information
data "rest_data" "user_info" {
  endpoint = "/api/v1/users/${rest_resource.user.response_data.id}"
  method   = "GET"
}

# Outputs
output "created_user_id" {
  description = "The ID of the created user"
  value       = rest_resource.user.response_data.id
}

output "user_status" {
  description = "The status of the user"
  value       = data.rest_data.user_info.response_data.status
}

# Variables
variable "api_token" {
  description = "API token for authentication"
  type        = string
  sensitive   = true
}