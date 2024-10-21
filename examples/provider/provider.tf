# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}

provider "rest" {
  api_token  = var.API_TOKEN
  api_header = var.API_TOKEN_HEADER
  api_url    = var.API_URL
}
