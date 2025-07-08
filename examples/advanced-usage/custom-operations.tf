# Custom HTTP Operations Example
# This example demonstrates advanced HTTP operations with custom headers, query parameters, and different methods

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "1.0.8"
    }
  }
}

provider "rest" {
  api_url   = "https://api.example.com"
  api_token = var.api_token
  
  timeout        = 60
  retry_attempts = 5
}

# Example: PATCH operation for partial updates
resource "rest_resource" "user_profile" {
  name     = "user-profile-${var.user_id}"
  endpoint = "/api/v1/users/${var.user_id}/profile"
  
  # Configure methods for each operation
  create_method = "PATCH"  # Use PATCH for partial updates
  read_method   = "GET"
  update_method = "PATCH"  # Use PATCH for partial updates
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
    "X-Update-Type" = "partial"
    "If-Match"     = var.etag  # Optimistic concurrency control
  }
  
  query_params = {
    "notify"    = "true"
    "validate"  = "true"
    "version"   = "v1"
  }
  
  body = jsonencode({
    profile = {
      bio           = "Updated bio information"
      location      = "San Francisco, CA"
      website       = "https://example.com"
      social_links = {
        twitter  = "@example"
        linkedin = "linkedin.com/in/example"
      }
    }
    preferences = {
      notifications = {
        email = true
        push  = false
        sms   = false
      }
      privacy = {
        profile_visibility = "public"
        activity_sharing   = "friends"
      }
    }
  })
  
  # Different body for updates (adding version info)
  update_body = jsonencode({
    profile = {
      bio           = "Updated bio information"
      location      = "San Francisco, CA"
      website       = "https://example.com"
      social_links = {
        twitter  = "@example"
        linkedin = "linkedin.com/in/example"
      }
    }
    preferences = {
      notifications = {
        email = true
        push  = false
        sms   = false
      }
      privacy = {
        profile_visibility = "public"
        activity_sharing   = "friends"
      }
    }
    metadata = {
      last_updated_by = "terraform"
      update_reason   = "profile_sync"
    }
  })
}

# Example: HEAD request to check resource existence
data "rest_data" "resource_exists" {
  endpoint = "/api/v1/users/${var.user_id}"
  method   = "HEAD"
  
  headers = {
    "Accept" = "application/json"
  }
}

# Example: Complex POST with file upload metadata
resource "rest_resource" "document_upload" {
  name     = "document-${var.user_id}"
  endpoint = "/api/v1/documents"
  
  # Configure methods for each operation
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type"   = "application/json"
    "X-Upload-Type"  = "metadata"
    "X-User-ID"      = var.user_id
  }
  
  query_params = {
    "async"      = "true"
    "webhook"    = var.webhook_url
    "encryption" = "aes256"
  }
  
  body = jsonencode({
    document = {
      name        = "Important Document.pdf"
      size        = 2048576
      mime_type   = "application/pdf"
      checksum    = "sha256:abcd1234..."
      upload_url  = "https://upload.example.com/documents/12345"
    }
    metadata = {
      category    = "financial"
      tags        = ["invoice", "2024", "quarterly"]
      retention   = "7_years"
      access_level = "restricted"
    }
    processing = {
      ocr_enabled     = true
      thumbnail_sizes = ["small", "medium", "large"]
      watermark       = var.watermark_enabled
    }
  })
  
  timeout        = 120  # Longer timeout for upload operations
  retry_attempts = 3
}

# Example: OPTIONS request to discover API capabilities
data "rest_data" "api_capabilities" {
  endpoint = "/api/v1/users"
  method   = "OPTIONS"
  
  headers = {
    "Origin" = "https://webapp.example.com"
  }
}

# Example: Complex query with POST (for complex search criteria)
data "rest_data" "advanced_search" {
  endpoint = "/api/v1/search"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
    "Accept"       = "application/json"
  }
  
  query_params = {
    "page"     = "1"
    "per_page" = "50"
    "sort"     = "relevance"
  }
  
  body = jsonencode({
    query = {
      bool = {
        must = [
          {
            match = {
              title = "terraform"
            }
          },
          {
            range = {
              created_at = {
                gte = "2024-01-01"
                lte = "2024-12-31"
              }
            }
          }
        ]
        filter = [
          {
            term = {
              status = "published"
            }
          },
          {
            terms = {
              tags = ["infrastructure", "automation"]
            }
          }
        ]
      }
    }
    aggregations = {
      categories = {
        terms = {
          field = "category.keyword"
          size  = 10
        }
      }
      monthly_counts = {
        date_histogram = {
          field    = "created_at"
          interval = "month"
        }
      }
    }
  })
}

# Outputs
output "profile_update_status" {
  description = "Profile update results"
  value = {
    profile_id   = rest_resource.user_profile.response_data.id
    version      = rest_resource.user_profile.response_data.version
    updated_at   = rest_resource.user_profile.response_data.updated_at
    etag         = rest_resource.user_profile.response_headers["ETag"]
  }
}

output "resource_check" {
  description = "Resource existence check results"
  value = {
    exists        = data.rest_data.resource_exists.status_code == 200
    last_modified = data.rest_data.resource_exists.response_headers["Last-Modified"]
    content_length = data.rest_data.resource_exists.response_headers["Content-Length"]
  }
}

output "document_upload_info" {
  description = "Document upload information"
  value = {
    document_id  = rest_resource.document_upload.response_data.id
    upload_url   = rest_resource.document_upload.response_data.upload_url
    expires_at   = rest_resource.document_upload.response_data.expires_at
    tracking_id  = rest_resource.document_upload.response_data.tracking_id
  }
}

output "api_capabilities" {
  description = "Available API capabilities"
  value = {
    allowed_methods = split(",", data.rest_data.api_capabilities.response_headers["Allow"])
    cors_headers    = data.rest_data.api_capabilities.response_headers["Access-Control-Allow-Headers"]
    max_age         = data.rest_data.api_capabilities.response_headers["Access-Control-Max-Age"]
  }
}

output "search_results" {
  description = "Advanced search results"
  value = {
    total_hits    = data.rest_data.advanced_search.response_data.hits.total.value
    max_score     = data.rest_data.advanced_search.response_data.hits.max_score
    categories    = data.rest_data.advanced_search.response_data.aggregations.categories.buckets
    monthly_data  = data.rest_data.advanced_search.response_data.aggregations.monthly_counts.buckets
  }
}

# Variables
variable "api_token" {
  description = "API token for authentication"
  type        = string
  sensitive   = true
}

variable "user_id" {
  description = "User ID for operations"
  type        = string
}

variable "etag" {
  description = "ETag for optimistic concurrency control"
  type        = string
  default     = "*"
}

variable "webhook_url" {
  description = "Webhook URL for async notifications"
  type        = string
}

variable "watermark_enabled" {
  description = "Enable watermarking for documents"
  type        = bool
  default     = true
}