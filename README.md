# Terraform REST Provider

## What is this provider?

The Terraform REST Provider bridges the gap between Terraform and REST APIs, allowing you to manage any REST-based resource as if it were a native Terraform resource. Instead of writing custom scripts or manual API calls, you can declaratively manage your REST resources alongside your infrastructure.

**Think of it as a universal translator** between Terraform and any REST API - whether you're managing users in a SaaS platform, configurations in a monitoring system, or resources in a custom application.

## Why use this provider?

Many organizations have internal APIs, SaaS platforms, or services that don't have dedicated Terraform providers. This provider fills that gap by:

- **Eliminating manual API management** - No more curl scripts or custom tooling
- **Providing real infrastructure as code** - REST resources get the same lifecycle management as your servers
- **Enabling drift detection** - Automatically detect when someone changes things outside of Terraform
- **Supporting complex workflows** - Chain API calls, use outputs from one resource in another

## Real-world examples

- **User Management**: Create users in your company's identity system
- **Configuration Management**: Deploy application settings to microservices
- **Monitoring Setup**: Configure alerts and dashboards in observability platforms
- **CI/CD Integration**: Manage build configurations and deployment settings
- **Third-party Integrations**: Set up webhooks, API keys, and service configurations

## Key Features

### **Complete Resource Lifecycle**
Create, read, update, and delete REST resources with full Terraform state management. Your API resources get the same `terraform plan`, `terraform apply`, and `terraform destroy` workflow as any other resource.

### **Flexible Authentication**
Supports the authentication methods you're already using:
- **API Tokens**: Bearer tokens, API keys, or custom headers
- **Certificate Authentication**: Client certificates (mTLS) for secure environments
- **PKCS12 Bundles**: Enterprise certificate formats

### **Universal HTTP Support**
Works with any REST API using standard HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS). Configure different methods for each operation (create, read, update, delete) to match your API's requirements.

### **Smart Response Handling**
Automatically parses JSON responses so you can access individual fields in your Terraform configurations. No more parsing JSON in bash scripts!

### **Production-Ready Features**
- **Automatic Retries**: Handles temporary network issues and API rate limits
- **Drift Detection**: Detects when resources change outside Terraform
- **Import Support**: Bring existing API resources under Terraform management
- **Error Recovery**: Exponential backoff and configurable retry logic

## Quick Start

### 1. Add the provider to your Terraform configuration

```hcl
terraform {
  required_providers {
    rest = {
      source  = "Mikej81/rest"
      version = "1.0.5"
    }
  }
}
```

### 2. Initialize Terraform

```bash
terraform init
```

### 3. Configure the provider

```hcl
provider "rest" {
  api_url = "https://api.example.com"
  # Add authentication (see examples below)
}
```

### 4. Create your first resource

```hcl
resource "rest_resource" "my_user" {
  name     = "john-doe"
  endpoint = "/users"
  
  # Use specific methods for different operations
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
  })
}
```

### 5. Apply your configuration

```bash
terraform plan
terraform apply
```

That's it! Your REST resource is now managed by Terraform.

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
  
  # New method configuration (recommended)
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  # Legacy method configuration (deprecated but still supported)
  # method = "POST"  # Only affects create operation
  
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
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PATCH"  # Use PATCH for partial updates
  delete_method = "DELETE"
  
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

- **create_method**: HTTP method for create operations (POST, PUT, PATCH)
- **read_method**: HTTP method for read operations (GET, POST, HEAD)
- **update_method**: HTTP method for update operations (PUT, PATCH, POST)
- **delete_method**: HTTP method for delete operations (DELETE, POST, PUT)
- **method**: HTTP method for create operations (DEPRECATED - use create_method instead)
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
