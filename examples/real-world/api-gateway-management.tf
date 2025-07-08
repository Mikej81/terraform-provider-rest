# Real-World Example: API Gateway Management
# This example demonstrates managing API gateway configurations through REST APIs

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}

provider "rest" {
  api_url   = var.gateway_api_url
  api_token = var.admin_token
  api_header = "X-Admin-Token"
  
  timeout        = 60
  retry_attempts = 3
}

# Create API gateway service
resource "rest_resource" "api_service" {
  name     = var.service_name
  endpoint = "/api/v1/services"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name     = var.service_name
    protocol = "https"
    host     = var.service_host
    port     = var.service_port
    path     = var.service_path
    retries  = 5
    connect_timeout = 60000
    write_timeout   = 60000
    read_timeout    = 60000
    tags = var.service_tags
  })
}

# Create route for the service
resource "rest_resource" "api_route" {
  name     = "${var.service_name}-route"
  endpoint = "/api/v1/routes"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name      = "${var.service_name}-route"
    protocols = ["https", "http"]
    methods   = ["GET", "POST", "PUT", "DELETE", "PATCH"]
    hosts     = var.route_hosts
    paths     = var.route_paths
    service = {
      id = rest_resource.api_service.response_data.id
    }
    strip_path      = false
    preserve_host   = false
    regex_priority  = 0
    path_handling   = "v1"
    tags           = var.route_tags
  })
}

# Add rate limiting plugin
resource "rest_resource" "rate_limit_plugin" {
  name     = "${var.service_name}-rate-limit"
  endpoint = "/api/v1/plugins"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = "rate-limiting"
    service = {
      id = rest_resource.api_service.response_data.id
    }
    config = {
      minute         = var.rate_limit_per_minute
      hour          = var.rate_limit_per_hour
      day           = var.rate_limit_per_day
      month         = var.rate_limit_per_month
      year          = var.rate_limit_per_year
      hide_client_headers = false
      fault_tolerant      = true
    }
    enabled = true
    tags    = ["rate-limiting", "security"]
  })
}

# Add authentication plugin
resource "rest_resource" "auth_plugin" {
  count = var.enable_auth ? 1 : 0
  
  name     = "${var.service_name}-auth"
  endpoint = "/api/v1/plugins"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = var.auth_plugin_type
    service = {
      id = rest_resource.api_service.response_data.id
    }
    config = var.auth_config
    enabled = true
    tags    = ["authentication", "security"]
  })
}

# Add CORS plugin for browser support
resource "rest_resource" "cors_plugin" {
  count = var.enable_cors ? 1 : 0
  
  name     = "${var.service_name}-cors"
  endpoint = "/api/v1/plugins"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = "cors"
    service = {
      id = rest_resource.api_service.response_data.id
    }
    config = {
      origins         = var.cors_origins
      methods         = ["GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"]
      headers         = var.cors_headers
      exposed_headers = var.cors_exposed_headers
      credentials     = var.cors_credentials
      max_age         = 3600
      preflight_continue = false
    }
    enabled = true
    tags    = ["cors", "browser-support"]
  })
}

# Create consumer for API access
resource "rest_resource" "api_consumer" {
  count = length(var.consumers)
  
  name     = var.consumers[count.index].username
  endpoint = "/api/v1/consumers"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    username    = var.consumers[count.index].username
    custom_id   = var.consumers[count.index].custom_id
    tags        = var.consumers[count.index].tags
  })
}

# Add API key for each consumer
resource "rest_resource" "consumer_api_key" {
  count = length(var.consumers)
  
  name     = "${var.consumers[count.index].username}-key"
  endpoint = "/api/v1/consumers/${rest_resource.api_consumer[count.index].response_data.id}/key-auth"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    key  = var.consumers[count.index].api_key
    tags = ["api-key", "auth"]
  })
}

# Query service health status
data "rest_data" "service_health" {
  endpoint = "/api/v1/services/${rest_resource.api_service.response_data.id}/health"
  method   = "GET"
  
  headers = {
    "Accept" = "application/json"
  }
}

# Get service statistics
data "rest_data" "service_stats" {
  endpoint = "/api/v1/services/${rest_resource.api_service.response_data.id}/stats"
  method   = "GET"
  
  query_params = {
    "period" = "1h"
    "metrics" = "requests,latency,errors"
  }
}

# Outputs
output "service_info" {
  description = "API service information"
  value = {
    id       = rest_resource.api_service.response_data.id
    name     = rest_resource.api_service.response_data.name
    protocol = rest_resource.api_service.response_data.protocol
    host     = rest_resource.api_service.response_data.host
    port     = rest_resource.api_service.response_data.port
    created_at = rest_resource.api_service.created_at
  }
}

output "route_info" {
  description = "API route information"
  value = {
    id        = rest_resource.api_route.response_data.id
    name      = rest_resource.api_route.response_data.name
    protocols = rest_resource.api_route.response_data.protocols
    hosts     = rest_resource.api_route.response_data.hosts
    paths     = rest_resource.api_route.response_data.paths
  }
}

output "plugins" {
  description = "Configured plugins"
  value = {
    rate_limit = {
      id      = rest_resource.rate_limit_plugin.response_data.id
      enabled = rest_resource.rate_limit_plugin.response_data.enabled
    }
    auth = var.enable_auth ? {
      id      = rest_resource.auth_plugin[0].response_data.id
      enabled = rest_resource.auth_plugin[0].response_data.enabled
    } : null
    cors = var.enable_cors ? {
      id      = rest_resource.cors_plugin[0].response_data.id
      enabled = rest_resource.cors_plugin[0].response_data.enabled
    } : null
  }
}

output "consumers" {
  description = "Created consumers and their API keys"
  value = [
    for i, consumer in var.consumers : {
      id       = rest_resource.api_consumer[i].response_data.id
      username = rest_resource.api_consumer[i].response_data.username
      api_key_id = rest_resource.consumer_api_key[i].response_data.id
    }
  ]
  sensitive = true
}

output "service_health" {
  description = "Service health status"
  value = {
    status = data.rest_data.service_health.response_data.status
    checks = data.rest_data.service_health.response_data.checks
  }
}

output "service_statistics" {
  description = "Service performance statistics"
  value = {
    total_requests   = data.rest_data.service_stats.response_data.total_requests
    success_rate     = data.rest_data.service_stats.response_data.success_rate
    avg_latency      = data.rest_data.service_stats.response_data.avg_latency
    error_rate       = data.rest_data.service_stats.response_data.error_rate
  }
}

# Variables
variable "gateway_api_url" {
  description = "API Gateway management URL"
  type        = string
}

variable "admin_token" {
  description = "Admin token for API Gateway"
  type        = string
  sensitive   = true
}

variable "service_name" {
  description = "Name of the API service"
  type        = string
}

variable "service_host" {
  description = "Backend service host"
  type        = string
}

variable "service_port" {
  description = "Backend service port"
  type        = number
  default     = 80
}

variable "service_path" {
  description = "Backend service path"
  type        = string
  default     = "/"
}

variable "service_tags" {
  description = "Tags for the service"
  type        = list(string)
  default     = []
}

variable "route_hosts" {
  description = "Route host names"
  type        = list(string)
}

variable "route_paths" {
  description = "Route paths"
  type        = list(string)
}

variable "route_tags" {
  description = "Tags for the route"
  type        = list(string)
  default     = []
}

variable "rate_limit_per_minute" {
  description = "Rate limit per minute"
  type        = number
  default     = 100
}

variable "rate_limit_per_hour" {
  description = "Rate limit per hour"
  type        = number
  default     = 1000
}

variable "rate_limit_per_day" {
  description = "Rate limit per day"
  type        = number
  default     = 10000
}

variable "rate_limit_per_month" {
  description = "Rate limit per month"
  type        = number
  default     = null
}

variable "rate_limit_per_year" {
  description = "Rate limit per year"
  type        = number
  default     = null
}

variable "enable_auth" {
  description = "Enable authentication plugin"
  type        = bool
  default     = true
}

variable "auth_plugin_type" {
  description = "Type of authentication plugin"
  type        = string
  default     = "key-auth"
}

variable "auth_config" {
  description = "Authentication plugin configuration"
  type        = any
  default     = {}
}

variable "enable_cors" {
  description = "Enable CORS plugin"
  type        = bool
  default     = false
}

variable "cors_origins" {
  description = "CORS allowed origins"
  type        = list(string)
  default     = ["*"]
}

variable "cors_headers" {
  description = "CORS allowed headers"
  type        = list(string)
  default     = ["Accept", "Accept-Version", "Content-Length", "Content-MD5", "Content-Type", "Date", "X-Auth-Token"]
}

variable "cors_exposed_headers" {
  description = "CORS exposed headers"
  type        = list(string)
  default     = []
}

variable "cors_credentials" {
  description = "CORS credentials support"
  type        = bool
  default     = false
}

variable "consumers" {
  description = "API consumers"
  type = list(object({
    username  = string
    custom_id = string
    api_key   = string
    tags      = list(string)
  }))
  default = []
}