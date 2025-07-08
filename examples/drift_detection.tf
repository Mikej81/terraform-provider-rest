# Example: REST resource with drift detection configuration
# This demonstrates handling of server-side metadata fields

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

# Example 1: Basic resource with default drift detection
# This will ignore common server-side fields automatically
resource "rest_resource" "user_basic" {
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

# Example 2: Resource with custom ignore fields
# Additional fields to ignore during drift detection
resource "rest_resource" "user_custom_ignore" {
  endpoint = "/users"
  name     = "jane-doe"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name  = "Jane Doe"
    email = "jane@example.com"
    role  = "user"
  })
  
  # Ignore additional server-side metadata fields
  ignore_fields = [
    "last_login",           # Server tracks last login time
    "login_count",          # Server increments login counter
    "profile_views",        # Server tracks profile views
    "account_status",       # Server may change based on activity
    "external_id",          # Server generates external references
    "computed_score"        # Server calculates dynamic scores
  ]
}

# Example 3: Resource with drift detection disabled
# Use when server response changes frequently but doesn't indicate drift
resource "rest_resource" "analytics_data" {
  endpoint = "/analytics/datasets"
  name     = "user-engagement"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name        = "User Engagement Dataset"
    description = "Tracks user engagement metrics"
    enabled     = true
  })
  
  # Disable drift detection for analytics data that changes constantly
  drift_detection = false
}

# Example 4: Resource expecting server to add metadata
# Common scenario where API adds fields like id, timestamps, etc.
resource "rest_resource" "blog_post" {
  endpoint = "/blog/posts"
  name     = "terraform-provider-guide"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    title   = "Terraform Provider Development Guide"
    content = "A comprehensive guide to building Terraform providers..."
    author  = "DevOps Team"
    tags    = ["terraform", "devops", "infrastructure"]
    status  = "draft"
  })
  
  # The API will add these fields, but they shouldn't trigger drift:
  # - id: "post-12345"
  # - created_at: "2025-01-15T10:30:00Z"
  # - updated_at: "2025-01-15T10:30:00Z"
  # - slug: "terraform-provider-guide"
  # - word_count: 1500
  # - reading_time: 7
  # - view_count: 0
  
  # These will be automatically ignored by default ignore list
}

# Example 5: Advanced configuration with conditional operations and drift detection
resource "rest_resource" "deployment" {
  endpoint = "/deployments"
  name     = "api-v2-deployment"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name        = "API v2 Deployment"
    image       = "api:v2.1.0"
    replicas    = 3
    environment = "production"
  })
  
  # Expected successful deployment status codes
  expected_status = [200, 201, 202]
  
  # Retry on deployment-related temporary failures
  retry_on_status = [503, 429]
  
  # Ignore deployment-specific server metadata
  ignore_fields = [
    "deployment_id",     # Server generates unique deployment ID
    "started_at",        # Server tracks deployment start time
    "duration",          # Server calculates deployment duration
    "previous_version",  # Server tracks version history
    "rollback_url",      # Server provides rollback endpoint
    "health_status",     # Server monitors deployment health
    "resource_usage"     # Server tracks CPU/memory usage
  ]
  
  # Enable detailed drift detection for critical deployments
  drift_detection = true
  
  timeout = 300  # 5 minutes for deployment timeout
  retry_attempts = 3
}

# Outputs to verify the behavior
output "user_basic_response" {
  description = "Response from basic user creation (includes server metadata)"
  value       = rest_resource.user_basic.response
}

output "blog_post_id" {
  description = "Server-generated blog post ID"
  value       = rest_resource.blog_post.id
}

output "deployment_status" {
  description = "Deployment status and metadata"
  value = {
    id            = rest_resource.deployment.id
    status_code   = rest_resource.deployment.status_code
    last_updated  = rest_resource.deployment.last_updated
  }
}