# Copyright (c) HashiCorp, Inc.


locals {
  resource_name = "object-name-1"
}

resource "rest_resource" "resource" {
  name           = local.resource_name # Define the name attribute here
  endpoint       = "/api/config"
  
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
