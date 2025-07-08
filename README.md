# Terraform REST Provider

## Overview

The **Terraform REST Provider** enables seamless interaction with REST APIs through Infrastructure as Code. This provider supports comprehensive CRUD operations, multiple authentication methods, and advanced features for managing REST-based resources within your Terraform configurations.

## Features

- **Complete CRUD Operations**: Create, read, update, and delete resources via REST APIs with full state management
- **Multiple Authentication Methods**: 
  - Token-based authentication with configurable headers
  - Client certificate authentication (mTLS) with PEM format
  - PKCS12 certificate bundle support
- **Advanced HTTP Support**: All standard HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
- **Dynamic Response Parsing**: Automatic JSON response parsing with accessible key-value outputs
- **Robust Error Handling**: Exponential backoff retry logic with configurable attempts
- **Import/Export Support**: Full Terraform import and export functionality
- **Drift Detection**: Automatic detection and handling of configuration drift
- **Flexible Configuration**: Custom headers, query parameters, timeouts, and SSL options

## Installation

To install the Generic REST Provider, add the following configuration to your Terraform file:

```hcl
terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}
```

Run the `terraform init` command to download and install the provider.

## Provider Configuration

The provider supports multiple authentication methods. Choose one that fits your API requirements:

### Token Authentication
```hcl
provider "rest" {
  api_url    = "https://api.example.com"
  api_token  = "your-api-token"
  api_header = "Authorization" # Optional, defaults to "Authorization"
}
```

### Certificate Authentication (PEM format)
```hcl
provider "rest" {
  api_url     = "https://api.example.com"
  client_cert = file("client.pem")
  client_key  = file("client-key.pem")
}
```

### Certificate Authentication (file paths)
```hcl
provider "rest" {
  api_url          = "https://api.example.com"
  client_cert_file = "/path/to/client.pem"
  client_key_file  = "/path/to/client-key.pem"
}
```

### PKCS12 Certificate Bundle
```hcl
provider "rest" {
  api_url         = "https://api.example.com"
  pkcs12_file     = "/path/to/certificate.p12"
  pkcs12_password = "certificate-password"
}
```

### Configuration Options

- **api_url** (required): Base URL for the REST API
- **timeout**: Request timeout in seconds (default: 30)
- **insecure**: Disable SSL certificate verification (default: false)
- **retry_attempts**: Number of retry attempts for failed requests (default: 3)
- **max_idle_conns**: Maximum idle HTTP connections (default: 100)

## Example Usage

### Data Source

Retrieve data from an API endpoint:

```hcl
data "rest_data" "user" {
  endpoint = "/api/users/123"
  method   = "GET"
  headers = {
    "Accept"       = "application/json"
    "Content-Type" = "application/json"
  }
}

# Access parsed response data
output "user_name" {
  value = data.rest_data.user.response_data.name
}

output "user_email" {
  value = data.rest_data.user.response_data.email
}
```

### Basic Resource Management

```hcl
resource "rest_resource" "user" {
  name     = "john-doe"
  endpoint = "/api/users"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
    role  = "admin"
  })
  
  # Optional: Custom update body
  update_body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
    role  = "admin"
    active = true
  })
}

# Access dynamic response data
output "user_id" {
  value = rest_resource.user.response_data.id
}

output "created_at" {
  value = rest_resource.user.created_at
}
```

### Advanced Resource with Custom Headers and Query Parameters

```hcl
resource "rest_resource" "api_configuration" {
  name     = "prod-config"
  endpoint = "/api/v2/configurations"
  method   = "POST"
  
  headers = {
    "Content-Type"    = "application/json"
    "X-Custom-Header" = "production"
  }
  
  query_params = {
    "environment" = "production"
    "validate"    = "true"
  }
  
  body = jsonencode({
    name = "production-config"
    settings = {
      timeout     = 30
      retry_count = 3
      debug       = false
    }
  })
  
  timeout        = 60
  retry_attempts = 5
}
```

### Import Existing Resources

```bash
# Import existing resources using endpoint/name format
terraform import rest_resource.existing_user "api/users/existing-user"
```

## Resource Attributes

### Computed Attributes

All resources provide these computed attributes for accessing response data:

- **response_data**: Parsed JSON response as key-value pairs (e.g., `response_data.id`, `response_data.name`)
- **response**: Raw response body
- **status_code**: HTTP response status code
- **response_headers**: HTTP response headers as key-value pairs
- **created_at**: Timestamp when resource was created
- **last_updated**: Timestamp when resource was last updated

### Configurable Attributes

- **method**: HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
- **headers**: Custom HTTP headers as key-value pairs
- **query_params**: URL query parameters as key-value pairs
- **body**: Request body for create operations
- **update_body**: Optional separate body for update operations
- **destroy_body**: Optional body for delete operations
- **timeout**: Request timeout in seconds
- **retry_attempts**: Number of retry attempts
- **insecure**: Disable SSL verification

## Authentication Methods

### Token Authentication
Best for APIs using bearer tokens, API keys, or custom token headers.

```hcl
provider "rest" {
  api_url    = "https://api.example.com"
  api_token  = var.api_token
  api_header = "X-API-Key"  # or "Authorization", "Bearer", etc.
}
```

### mTLS Certificate Authentication
For APIs requiring client certificate authentication:

```hcl
provider "rest" {
  api_url     = "https://secure-api.example.com"
  client_cert = file("${path.module}/certs/client.pem")
  client_key  = file("${path.module}/certs/client-key.pem")
}
```

### PKCS12 Certificates
For enterprise environments using PKCS12 certificate bundles:

```hcl
provider "rest" {
  api_url         = "https://enterprise-api.example.com"
  pkcs12_file     = "/secure/certs/client.p12"
  pkcs12_password = var.cert_password
}
```

## Best Practices

### Security
- Use Terraform variables or environment variables for sensitive data
- Enable SSL verification in production (`insecure = false`)
- Store certificates in secure locations with proper permissions
- Rotate API tokens and certificates regularly

### Performance
- Configure appropriate timeouts for your API response times
- Set retry attempts based on your API's reliability
- Use connection pooling with `max_idle_conns` for high-throughput scenarios

### State Management
- Leverage import functionality to bring existing resources under management
- Use computed attributes to access dynamic response data
- Monitor drift detection warnings in Terraform output

## Advanced Features

### Drift Detection
The provider automatically detects when resources change outside of Terraform and updates the state accordingly.

### Dynamic Response Access
Access any field from JSON responses directly:
```hcl
output "api_version" {
  value = rest_resource.api.response_data.version
}

output "response_headers" {
  value = rest_resource.api.response_headers
}
```

### Import Support
Import existing API resources into Terraform management:
```bash
terraform import rest_resource.existing_api_config "api/v1/config/existing-config"
```

## Contributing

Contributions are welcome! If you would like to contribute to this provider, please feel free to open an issue or a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
