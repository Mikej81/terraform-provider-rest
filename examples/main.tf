# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    generic_rest = {
      source  = "local/generic-rest-provider"
      version = "0.1.0"
    }
  }
}

provider "generic_rest" {
  api_token        = ""
  api_token_header = ""
  api_url          = ""
}

resource "generic_rest_resource" "example" {
  name = "test-resource"
  data = {
    key1 = "value1"
    key2 = "value2"
  }
}
