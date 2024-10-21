# Generic REST Provider for Terraform

## Overview

The **Generic REST Provider** for Terraform enables users to interact with REST APIs seamlessly, allowing them to manage resources through simple CRUD operations. This provider is designed to integrate with any REST API, making it easy to automate the creation, reading, updating, and deletion of resources in your infrastructure.

By leveraging the power of Terraform, you can bring your REST-based integrations under infrastructure-as-code management, providing version control, state tracking, and a consistent way to manage your external services.

## Features

- **CRUD Operations**: Provides a unified interface for creating, reading, updating, and deleting resources through REST APIs.
- **Customizable Requests**: Supports various configuration options such as custom headers, request bodies, and query parameters.
- **Timeouts and Retries**: Includes customizable request timeouts, retry attempts, and options to skip SSL verification for enhanced flexibility.
- **Dynamic Resource Handling**: Automatically manage the state of resources to ensure infrastructure consistency and eliminate configuration drift.
- **Supports Authentication**: Integrate API tokens and headers for secure access to authenticated endpoints.

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

You can configure the provider to connect to your REST API by setting up the required parameters:

```hcl
provider "rest" {
  api_token  = var.API_TOKEN
  api_header = var.API_TOKEN_HEADER
  api_url    = var.API_URL
}
```

- **api_token**: The token used for API authentication.
- **api_header**: The HTTP header key for passing the API token.
- **api_url**: The base URL for the REST API.

## Example Usage

### Data Source

To retrieve data from an API endpoint:

```hcl
data "rest_data" "example" {
  endpoint      = "/api/web/namespaces/system/tenant/settings"
  timeout       = 30  # Timeout in seconds for data source request
  insecure      = true # Disable SSL verification if needed
  retry_attempts = 3   # Number of retry attempts for data source request
}

output "data_response" {
  value = data.rest_data.example.response
}
```

### Resource Creation

To create or manage a resource using the provider:

```hcl
locals {
  resource_name = "dc-cluster-group-1"
}

resource "rest_resource" "resource" {
  name          = local.resource_name
  endpoint      = "/api/config/namespaces/system/dc_cluster_groups"
  timeout       = 30  # Timeout in seconds for resource request
  insecure      = true # Disable SSL verification if needed
  retry_attempts = 3   # Number of retry attempts for resource request

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

output "resource_response" {
  value = rest_resource.resource.response
}
```

## Configuration Options

- **Timeout**: Set a custom timeout (in seconds) for each request to prevent long wait times. Defaults can be overridden in both the data source and resource blocks.
- **Insecure**: Set to `true` to disable SSL certificate verification, useful in development environments with self-signed certificates.
- **Retry Attempts**: Specify the number of retry attempts if the initial request fails. This helps handle transient network issues or API rate limits.

## Use Cases

- **Third-Party Integration**: Easily integrate with third-party services that expose REST APIs to automate configurations and management.
- **Custom Resource Management**: Manage non-standard resources by interacting directly with proprietary or legacy REST APIs.
- **Infrastructure-as-Code**: Treat REST API configurations as code, enabling better collaboration, review, and tracking of changes.

## Best Practices

- **API Rate Limits**: Be mindful of rate limits imposed by the REST API you are interacting with. Configure retries accordingly to avoid overwhelming the API.
- **Security**: Avoid hardcoding API tokens directly in `.tf` files. Use environment variables or a secure Terraform variable store.
- **Error Handling**: Always test your configuration with smaller sets of data before using in production environments. Make use of timeout and retry options to avoid issues.

## Limitations

- **State Management**: The provider manages state for resources it creates, but other changes made directly to the REST API (outside of Terraform) may lead to drift. Use the `Read` operation to verify resource consistency.
- **Asynchronous APIs**: Currently, asynchronous API operations are not supported. You may need additional logic to handle APIs that return immediately but perform actions asynchronously.

## Future Enhancements

- **OAuth2 Support**: Adding support for OAuth2 to authenticate with APIs that use this standard.
- **Bulk Operations**: Support for batch creating or updating multiple resources in one call.
- **GraphQL Support**: Extending the provider to support GraphQL queries and mutations.

## Contributing

Contributions are welcome! If you would like to contribute to this provider, please feel free to open an issue or a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
