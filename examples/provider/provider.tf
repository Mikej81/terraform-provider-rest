# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "1.0.8"
    }
  }
}

# Example 1: Token Authentication
provider "rest" {
  alias      = "token_auth"
  api_url    = var.API_URL
  api_token  = var.API_TOKEN
  api_header = "Authorization" # Optional, defaults to "Authorization"
  
  # Optional connection settings
  timeout        = 30
  retry_attempts = 3
  insecure       = false
}

# Example 2: Client Certificate Authentication (mTLS)
provider "rest" {
  alias       = "cert_auth"
  api_url     = var.API_URL
  client_cert = file("${path.module}/certs/client.pem")
  client_key  = file("${path.module}/certs/client-key.pem")
  
  timeout = 60
}

# Example 3: PKCS12 Certificate Authentication
provider "rest" {
  alias           = "pkcs12_auth"
  api_url         = var.API_URL
  pkcs12_file     = var.PKCS12_FILE_PATH
  pkcs12_password = var.PKCS12_PASSWORD
  
  retry_attempts = 5
}
