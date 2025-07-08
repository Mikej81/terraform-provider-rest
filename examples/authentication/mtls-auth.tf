# mTLS Certificate Authentication Example
# This example demonstrates client certificate authentication (mutual TLS)

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}

# Provider configuration with client certificate authentication
provider "rest" {
  api_url     = "https://secure-api.example.com"
  client_cert = file("${path.module}/certs/client.pem")
  client_key  = file("${path.module}/certs/client-key.pem")
  
  timeout = 60
  insecure = false # Ensure SSL verification is enabled for security
}

# Example: Secure configuration management
resource "rest_resource" "secure_config" {
  name     = "production-config"
  endpoint = "/api/v1/configurations"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
    "X-Environment" = "production"
  }
  
  body = jsonencode({
    name = "prod-config"
    settings = {
      encryption_enabled = true
      audit_logging     = true
      max_connections   = 1000
      timeout_seconds   = 30
    }
    security = {
      tls_version = "1.3"
      cipher_suites = [
        "TLS_AES_256_GCM_SHA384",
        "TLS_CHACHA20_POLY1305_SHA256"
      ]
    }
  })
}

# Example: Query secure endpoint with certificate auth
data "rest_data" "security_status" {
  endpoint = "/api/v1/security/status"
  method   = "GET"
  
  headers = {
    "Accept" = "application/json"
  }
}

# Outputs
output "config_id" {
  description = "The ID of the created configuration"
  value       = rest_resource.secure_config.response_data.id
}

output "security_level" {
  description = "Current security level"
  value       = data.rest_data.security_status.response_data.security_level
}

output "certificate_expires" {
  description = "Certificate expiration date"
  value       = data.rest_data.security_status.response_data.cert_expiry
}