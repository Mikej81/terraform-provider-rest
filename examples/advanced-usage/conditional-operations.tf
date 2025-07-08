# Advanced Example: Conditional Operations Based on Response Codes
# This example demonstrates how to handle different response codes with custom actions

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}

provider "rest" {
  api_url   = var.api_url
  api_token = var.api_token
  
  timeout        = 30
  retry_attempts = 3
}

# Example 1: Only accept specific status codes as successful
resource "rest_resource" "strict_validation" {
  name     = "strict-user"
  endpoint = "/api/v1/users"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
  })
  
  # Only accept 201 Created as successful
  expected_status = [201]
  
  # If we get any other status, fail the operation
  on_failure = "fail"
}

# Example 2: Continue on specific "error" codes
resource "rest_resource" "flexible_handling" {
  name     = "flexible-item"
  endpoint = "/api/v1/items"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = "Test Item"
    type = "demo"
  })
  
  # Accept both success and conflict as "successful"
  expected_status = [200, 201, 409]
  
  # Continue processing even on failure
  on_failure = "continue"
}

# Example 3: Custom retry logic
resource "rest_resource" "custom_retry" {
  name     = "retry-item"
  endpoint = "/api/v1/queue/items"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    action = "process"
    priority = "high"
  })
  
  # Retry on rate limiting and service unavailable
  retry_on_status = [429, 503, 423] # 423 = Locked
  
  # Also retry when we get these statuses as the failure action
  expected_status = [200, 201, 202]
  on_failure = "retry"
  
  retry_attempts = 5
}

# Example 4: Treat normally successful codes as failures
resource "rest_resource" "strict_requirements" {
  name     = "strict-validation"
  endpoint = "/api/v1/validation"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    data = "test-data"
    validate = true
  })
  
  # Treat 202 Accepted as a failure (we want immediate processing)
  fail_on_status = [202, 204]
  
  # Only 200 OK and 201 Created are acceptable
  expected_status = [200, 201]
  
  on_failure = "fail"
}

# Example 5: Stop processing after success
resource "rest_resource" "stop_on_success" {
  name     = "one-time-action"
  endpoint = "/api/v1/actions/trigger"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    action = "initialize"
    once = true
  })
  
  # Stop any further processing after success
  on_success = "stop"
  
  # Accept multiple success codes
  expected_status = [200, 201, 204]
}

# Example 6: Complex conditional logic
resource "rest_resource" "complex_handling" {
  name     = "complex-workflow"
  endpoint = "/api/v1/workflows"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
    "X-Workflow-Type" = "conditional"
  }
  
  body = jsonencode({
    workflow = {
      name = "conditional-example"
      steps = [
        {
          type = "validation"
          required = true
        },
        {
          type = "processing"
          optional = false
        }
      ]
    }
  })
  
  # Accept normal success codes
  expected_status = [200, 201, 202]
  
  # Continue on most failures to allow manual intervention
  on_failure = "continue"
  
  # But fail immediately on auth/permission issues
  fail_on_status = [401, 403, 404]
  
  # Retry on temporary issues
  retry_on_status = [429, 503, 504, 520, 521, 522, 523, 524]
  
  retry_attempts = 3
  timeout = 60
}

# Query status of conditional operations
data "rest_data" "operation_status" {
  endpoint = "/api/v1/operations/${rest_resource.complex_handling.response_data.id}/status"
  method   = "GET"
  
  # Only succeed if the operation is complete
  expected_status = [200]
  
  # Retry if still processing
  retry_on_status = [202, 423]
}

# Outputs to demonstrate response handling
output "strict_validation" {
  description = "Result of strict validation"
  value = {
    status_code = rest_resource.strict_validation.status_code
    response_data = rest_resource.strict_validation.response_data
  }
}

output "flexible_handling" {
  description = "Result of flexible handling"
  value = {
    status_code = rest_resource.flexible_handling.status_code
    success = rest_resource.flexible_handling.status_code >= 200 && rest_resource.flexible_handling.status_code < 300
    response_data = rest_resource.flexible_handling.response_data
  }
}

output "custom_retry_result" {
  description = "Result after custom retry logic"
  value = {
    status_code = rest_resource.custom_retry.status_code
    attempts_made = "See logs for retry attempts"
    response_data = rest_resource.custom_retry.response_data
  }
}

output "operation_status" {
  description = "Status of the complex operation"
  value = {
    workflow_id = rest_resource.complex_handling.response_data.id
    status = data.rest_data.operation_status.response_data.status
    last_updated = data.rest_data.operation_status.response_data.updated_at
  }
}

# Variables
variable "api_url" {
  description = "API base URL"
  type        = string
}

variable "api_token" {
  description = "API authentication token"
  type        = string
  sensitive   = true
}