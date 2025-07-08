---
page_title: "rest Provider"
subcategory: ""
description: |-
  The REST provider enables interaction with REST APIs through Terraform. It supports multiple authentication methods including token-based, client certificates, and PKCS12 bundles.
---

# REST Provider

This REST provider allows Terraform to interact with REST APIs, providing full CRUD operations with state management, drift detection, and dynamic response parsing.

## How to Use This Provider

Add the REST provider to your Terraform configuration:

```terraform
terraform {
  required_providers {
    rest = {
      source  = "Mikej81/rest"
      version = "1.0.5"
    }
  }
}

provider "rest" {
  api_url = "https://your-api.example.com"
  # Add authentication - see examples below
}
```

Then run `terraform init` to download the provider.

## What Makes This Provider Different

- **Multiple Authentication Methods**: Token, client certificates (mTLS), and PKCS12 support
- **Complete HTTP Method Support**: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- **Dynamic Response Parsing**: Automatic JSON parsing with accessible key-value outputs
- **Import/Export Support**: Full Terraform import and export functionality
- **Robust Error Handling**: Exponential backoff retry logic with configurable attempts
- **Drift Detection**: Automatic detection and handling of configuration drift

## Authentication Methods

### Token Authentication

```terraform
provider "rest" {
  api_url    = "https://api.example.com"
  api_token  = var.api_token
  api_header = "Authorization" # Optional, defaults to "Authorization"
}
```

### Client Certificate Authentication (mTLS)

```terraform
# Using inline PEM content
provider "rest" {
  api_url     = "https://secure-api.example.com"
  client_cert = file("client.pem")
  client_key  = file("client-key.pem")
}

# Using file paths
provider "rest" {
  api_url          = "https://secure-api.example.com"
  client_cert_file = "/path/to/client.pem"
  client_key_file  = "/path/to/client-key.pem"
}
```

### PKCS12 Certificate Authentication

```terraform
# Using file path
provider "rest" {
  api_url         = "https://enterprise-api.example.com"
  pkcs12_file     = "/secure/certs/client.p12"
  pkcs12_password = var.cert_password
}

# Using base64-encoded content
provider "rest" {
  api_url         = "https://enterprise-api.example.com"
  pkcs12_bundle   = var.pkcs12_base64_content
  pkcs12_password = var.cert_password
}
```

## Schema

### Required

- `api_url` (String) The base URL for the REST API.

### Optional

**Authentication Options (choose one method):**

- `api_token` (String, Sensitive) The API token for authenticating requests
- `api_header` (String) The HTTP header name for the API token (default: "Authorization")
- `client_cert` (String, Sensitive) Client certificate for mTLS authentication (PEM format)
- `client_key` (String, Sensitive) Client private key for mTLS authentication (PEM format)
- `client_cert_file` (String) Path to client certificate file for mTLS authentication
- `client_key_file` (String) Path to client private key file for mTLS authentication
- `pkcs12_bundle` (String, Sensitive) PKCS12 certificate bundle (base64 encoded)
- `pkcs12_file` (String) Path to PKCS12 certificate bundle file
- `pkcs12_password` (String, Sensitive) Password for PKCS12 certificate bundle

**Connection Options:**

- `timeout` (Number) Default timeout for HTTP requests in seconds (default: 30)
- `insecure` (Boolean) Disable SSL certificate verification (default: false)
- `retry_attempts` (Number) Default number of retry attempts for failed requests (default: 3)
- `max_idle_conns` (Number) Maximum number of idle HTTP connections (default: 100)
