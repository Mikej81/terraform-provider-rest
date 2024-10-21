# Copyright (c) HashiCorp, Inc.

data "rest_data" "example" {
  endpoint       = "/api/web/namespaces/system/tenant/settings"
  timeout        = 30   # Timeout in seconds for data source request
  insecure       = true # Disable SSL verification if needed
  retry_attempts = 3    # Number of retry attempts for data source request
}
