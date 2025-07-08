# Terraform REST Provider Examples

This directory contains comprehensive examples demonstrating various use cases and features of the Terraform REST Provider.

## Directory Structure

- **`provider/`** - Basic provider configuration examples
- **`authentication/`** - Authentication method examples (Token, mTLS, PKCS12)
- **`data-sources/`** - Data source usage examples
- **`resources/`** - Resource management examples
- **`advanced-usage/`** - Advanced features and patterns
- **`real-world/`** - Production-ready integration examples

## Quick Start

### Basic Provider Setup

```hcl
terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}

provider "rest" {
  api_url   = "https://api.example.com"
  api_token = var.api_token
}
```

### Simple Resource Creation

```hcl
resource "rest_resource" "example" {
  name     = "my-resource"
  endpoint = "/api/v1/items"
  method   = "POST"
  
  body = jsonencode({
    name = "Example Item"
    type = "demo"
  })
}

output "resource_id" {
  value = rest_resource.example.response_data.id
}
```

## Authentication Examples

### Token Authentication
- **File**: `authentication/token-auth.tf`
- **Use Case**: Standard API token authentication
- **Features**: Custom headers, user management

### mTLS Certificate Authentication
- **File**: `authentication/mtls-auth.tf`
- **Use Case**: Mutual TLS authentication for secure APIs
- **Features**: Client certificates, secure configuration management

### PKCS12 Authentication
- **File**: `authentication/pkcs12-auth.tf`
- **Use Case**: Enterprise certificate bundles
- **Features**: PKCS12 certificates, enterprise policy management

## Advanced Usage Examples

### Dynamic Response Parsing
- **File**: `advanced-usage/dynamic-responses.tf`
- **Features**: 
  - Automatic JSON response parsing
  - Accessing nested response data
  - Using response data in dependent resources
  - Response headers and metadata

### Custom HTTP Operations
- **File**: `advanced-usage/custom-operations.tf`
- **Features**:
  - All HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
  - Custom headers and query parameters
  - Conditional operations
  - Complex request patterns

## Real-World Integration Examples

### API Gateway Management
- **File**: `real-world/api-gateway-management.tf`
- **Use Case**: Complete API gateway configuration
- **Features**:
  - Service and route creation
  - Plugin management (rate limiting, auth, CORS)
  - Consumer management
  - Health monitoring

### Monitoring and Alerting Setup
- **File**: `real-world/monitoring-setup.tf`
- **Use Case**: Comprehensive monitoring infrastructure
- **Features**:
  - Dashboard creation
  - Alert rule configuration
  - Data source management
  - Notification channels

## Key Features Demonstrated

### Dynamic Response Access
```hcl
# Automatic JSON parsing
output "user_email" {
  value = rest_resource.user.response_data.email
}

# Access response headers
output "rate_limit" {
  value = rest_resource.api_call.response_headers["X-RateLimit-Remaining"]
}

# Metadata access
output "created_at" {
  value = rest_resource.item.created_at
}
```

### Import Support
```bash
# Import existing resources
terraform import rest_resource.existing "api/users/user-123"
```

### Error Handling and Retries
```hcl
provider "rest" {
  api_url        = "https://api.example.com"
  api_token      = var.token
  retry_attempts = 5
  timeout        = 60
}
```

### Custom Request Bodies
```hcl
resource "rest_resource" "example" {
  # Different bodies for different operations
  body        = jsonencode({ action = "create" })
  update_body = jsonencode({ action = "update", version = 2 })
  destroy_body = jsonencode({ force = true })
}
```

## Running the Examples

1. **Navigate to an example directory**:
   ```bash
   cd examples/authentication
   ```

2. **Initialize Terraform**:
   ```bash
   terraform init
   ```

3. **Set required variables**:
   ```bash
   export TF_VAR_api_token="your-api-token"
   export TF_VAR_api_url="https://your-api.example.com"
   ```

4. **Plan and apply**:
   ```bash
   terraform plan
   terraform apply
   ```

## Variable Files

Most examples include `variables.tf` files with descriptions and defaults. Create a `terraform.tfvars` file to customize values:

```hcl
# terraform.tfvars
api_url = "https://your-api.example.com"
api_token = "your-secure-token"
application_name = "my-app"
environment = "production"
```

## Best Practices Demonstrated

1. **Security**:
   - Sensitive variable handling
   - Certificate management
   - SSL verification

2. **State Management**:
   - Import existing resources
   - Drift detection
   - Resource dependencies

3. **Error Handling**:
   - Retry configuration
   - Timeout management
   - Graceful error handling

4. **Production Readiness**:
   - Comprehensive monitoring
   - Security policies
   - Performance optimization

## Troubleshooting

### Common Issues

1. **Authentication Errors**:
   - Verify API tokens and certificates
   - Check header configurations
   - Validate SSL settings

2. **Timeout Issues**:
   - Increase timeout values
   - Configure retry attempts
   - Check network connectivity

3. **Response Parsing**:
   - Verify JSON response format
   - Check response data access patterns
   - Review error logs

### Debug Mode

Enable debug logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

## Documentation Tool Compatibility

The following files are used by the documentation generation tool:

* **provider/provider.tf** example file for the provider index page
* **data-sources/`full data source name`/data-source.tf** example file for the named data source page
* **resources/`full resource name`/resource.tf** example file for the named data source page

All other *.tf files are available for manual testing and can be run via the Terraform CLI.
