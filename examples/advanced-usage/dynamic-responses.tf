# Dynamic Response Parsing Example
# This example demonstrates how to use dynamic response data parsing features

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}

provider "rest" {
  api_url   = "https://api.example.com"
  api_token = var.api_token
}

# Example: Create a complex resource with nested response data
resource "rest_resource" "application" {
  name     = "my-application"
  endpoint = "/api/v1/applications"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = "My Application"
    type = "web"
    configuration = {
      replicas = 3
      resources = {
        memory = "512Mi"
        cpu    = "250m"
      }
      environment = {
        NODE_ENV = "production"
        PORT     = "3000"
      }
    }
    metadata = {
      labels = {
        app     = "my-application"
        version = "1.0.0"
        tier    = "frontend"
      }
    }
  })
}

# Example: Use response data to create dependent resources
resource "rest_resource" "application_config" {
  name     = "config-${rest_resource.application.response_data.id}"
  endpoint = "/api/v1/configurations"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    application_id = rest_resource.application.response_data.id
    config_name    = "production-config"
    settings = {
      database_url = "postgresql://prod-db:5432/app"
      redis_url    = "redis://prod-redis:6379"
      api_base_url = rest_resource.application.response_data.endpoints.api
    }
  })
}

# Example: Query application status using dynamic data
data "rest_data" "application_status" {
  endpoint = "/api/v1/applications/${rest_resource.application.response_data.id}/status"
  method   = "GET"
  
  headers = {
    "Accept" = "application/json"
  }
}

# Example: Create monitoring based on application metrics
resource "rest_resource" "monitoring_alert" {
  name     = "alert-${rest_resource.application.response_data.id}"
  endpoint = "/api/v1/alerts"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name         = "High CPU Usage - ${rest_resource.application.response_data.name}"
    application_id = rest_resource.application.response_data.id
    metric       = "cpu_usage"
    threshold    = 80
    notification = {
      email = ["ops@example.com"]
      slack = rest_resource.application.response_data.metadata.slack_channel
    }
    conditions = {
      duration = "5m"
      severity = "warning"
    }
  })
}

# Outputs demonstrating dynamic response access
output "application_details" {
  description = "Application details from API response"
  value = {
    id          = rest_resource.application.response_data.id
    name        = rest_resource.application.response_data.name
    status      = rest_resource.application.response_data.status
    created_at  = rest_resource.application.created_at
    last_updated = rest_resource.application.last_updated
  }
}

output "application_endpoints" {
  description = "Application endpoints from response"
  value = {
    api_endpoint    = rest_resource.application.response_data.endpoints.api
    health_endpoint = rest_resource.application.response_data.endpoints.health
    metrics_endpoint = rest_resource.application.response_data.endpoints.metrics
  }
}

output "application_metadata" {
  description = "Application metadata from response"
  value = {
    namespace = rest_resource.application.response_data.metadata.namespace
    labels    = rest_resource.application.response_data.metadata.labels
    annotations = rest_resource.application.response_data.metadata.annotations
  }
}

output "current_status" {
  description = "Current application status"
  value = {
    health        = data.rest_data.application_status.response_data.health
    replicas      = data.rest_data.application_status.response_data.replicas
    ready_replicas = data.rest_data.application_status.response_data.ready_replicas
    uptime        = data.rest_data.application_status.response_data.uptime
  }
}

output "response_headers" {
  description = "Response headers from the API"
  value = {
    rate_limit_remaining = rest_resource.application.response_headers["X-RateLimit-Remaining"]
    rate_limit_reset     = rest_resource.application.response_headers["X-RateLimit-Reset"]
    api_version         = rest_resource.application.response_headers["X-API-Version"]
  }
}

# Variables
variable "api_token" {
  description = "API token for authentication"
  type        = string
  sensitive   = true
}