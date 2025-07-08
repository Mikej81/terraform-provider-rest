# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "1.0.8"
    }
  }
}

provider "rest" {
  api_token  = var.API_TOKEN
  api_header = var.API_TOKEN_HEADER
  api_url    = var.API_URL
}

data "rest_data" "example" {
  endpoint       = "/api/web/namespaces/system/tenant/settings"
  timeout        = 30   # Timeout in seconds for data source request
  insecure       = true # Disable SSL verification if needed
  retry_attempts = 3    # Number of retry attempts for data source request
}

locals {
  resource_name = "dc-cluster-group-1"
}

resource "rest_resource" "resource" {
  name           = local.resource_name # Define the name attribute here
  endpoint       = "/api/config/namespaces/system/dc_cluster_groups"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  timeout        = 30   # Timeout in seconds for resource request
  insecure       = true # Disable SSL verification if needed
  retry_attempts = 3    # Number of retry attempts for resource request

  body = jsonencode({
    metadata = {
      name    = local.resource_name
      disable = false
    }
    spec = {
      type = {
        data_plane_mesh = {}
      }
    }
  })
  destroy_body = jsonencode({
    fail_if_referred = true,
    name             = local.resource_name,
    namespace        = "system"
  })
}

output "data_response" {
  value = data.rest_data.example.response
}

output "data_status_code" {
  value = data.rest_data.example.status_code
}

output "data_parsed_data" {
  value = data.rest_data.example.parsed_data
}

output "resource_response" {
  value = rest_resource.resource.response
}

output "resource_status_code" {
  value = rest_resource.resource.status_code
}
