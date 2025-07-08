# PKCS12 Certificate Authentication Example
# This example demonstrates PKCS12 certificate bundle authentication

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "0.1.0"
    }
  }
}

# Provider configuration with PKCS12 certificate authentication
provider "rest" {
  api_url         = "https://enterprise-api.example.com"
  pkcs12_file     = var.pkcs12_cert_path
  pkcs12_password = var.pkcs12_password
  
  timeout        = 45
  retry_attempts = 5
}

# Example: Enterprise resource management
resource "rest_resource" "enterprise_policy" {
  name     = "security-policy-2024"
  endpoint = "/api/v2/policies"
  method   = "POST"
  
  headers = {
    "Content-Type"     = "application/json"
    "X-Policy-Version" = "2024.1"
  }
  
  query_params = {
    "validate" = "true"
    "notify"   = "administrators"
  }
  
  body = jsonencode({
    name = "Security Policy 2024"
    type = "security"
    rules = [
      {
        id     = "rule-001"
        action = "allow"
        conditions = {
          ip_ranges = ["10.0.0.0/8", "192.168.0.0/16"]
          protocols = ["https", "ssh"]
        }
      },
      {
        id     = "rule-002"
        action = "deny"
        conditions = {
          ports = [23, 21, 80]
        }
      }
    ]
    compliance = {
      soc2     = true
      iso27001 = true
      pci_dss  = false
    }
  })
  
  # Custom update body for policy modifications
  update_body = jsonencode({
    name = "Security Policy 2024"
    type = "security"
    version = "2024.1.1"
    rules = [
      {
        id     = "rule-001"
        action = "allow"
        conditions = {
          ip_ranges = ["10.0.0.0/8", "192.168.0.0/16"]
          protocols = ["https", "ssh"]
        }
      },
      {
        id     = "rule-002"
        action = "deny"
        conditions = {
          ports = [23, 21, 80]
        }
      }
    ]
    compliance = {
      soc2     = true
      iso27001 = true
      pci_dss  = true # Updated compliance requirement
    }
  })
}

# Example: Query enterprise compliance status
data "rest_data" "compliance_report" {
  endpoint = "/api/v2/compliance/report"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
    "Accept"       = "application/json"
  }
  
  body = jsonencode({
    report_type = "summary"
    standards   = ["soc2", "iso27001", "pci_dss"]
    date_range = {
      start = "2024-01-01"
      end   = "2024-12-31"
    }
  })
}

# Outputs
output "policy_id" {
  description = "The ID of the created security policy"
  value       = rest_resource.enterprise_policy.response_data.id
}

output "policy_version" {
  description = "Current policy version"
  value       = rest_resource.enterprise_policy.response_data.version
}

output "compliance_score" {
  description = "Overall compliance score"
  value       = data.rest_data.compliance_report.response_data.overall_score
}

output "compliance_status" {
  description = "Compliance status by standard"
  value = {
    soc2     = data.rest_data.compliance_report.response_data.soc2_status
    iso27001 = data.rest_data.compliance_report.response_data.iso27001_status
    pci_dss  = data.rest_data.compliance_report.response_data.pci_dss_status
  }
}

# Variables
variable "pkcs12_cert_path" {
  description = "Path to the PKCS12 certificate file"
  type        = string
}

variable "pkcs12_password" {
  description = "Password for the PKCS12 certificate"
  type        = string
  sensitive   = true
}