# Real-World Example: Monitoring and Alerting Setup
# This example demonstrates setting up monitoring, metrics, and alerting through REST APIs

terraform {
  required_providers {
    rest = {
      source  = "local/rest"
      version = "1.0.8"
    }
  }
}

provider "rest" {
  api_url   = var.monitoring_api_url
  api_token = var.monitoring_token
  
  timeout        = 45
  retry_attempts = 3
}

# Create monitoring workspace
resource "rest_resource" "monitoring_workspace" {
  name     = var.workspace_name
  endpoint = "/api/v1/workspaces"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name        = var.workspace_name
    description = var.workspace_description
    settings = {
      retention_days     = var.retention_days
      sampling_rate     = var.sampling_rate
      enable_profiling  = var.enable_profiling
      data_compression  = true
    }
    tags = var.workspace_tags
  })
}

# Create data sources for monitoring
resource "rest_resource" "prometheus_datasource" {
  name     = "${var.workspace_name}-prometheus"
  endpoint = "/api/v1/workspaces/${rest_resource.monitoring_workspace.response_data.id}/datasources"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = "Prometheus"
    type = "prometheus"
    url  = var.prometheus_url
    access = "proxy"
    basic_auth = var.prometheus_basic_auth
    basic_auth_user = var.prometheus_username
    basic_auth_password = var.prometheus_password
    json_data = {
      httpMethod        = "POST"
      queryTimeout      = "60s"
      timeInterval      = "30s"
      prometheusType    = "Prometheus"
      prometheusVersion = "2.40.0"
    }
    is_default = true
  })
}

# Create application dashboard
resource "rest_resource" "application_dashboard" {
  name     = "${var.application_name}-dashboard"
  endpoint = "/api/v1/workspaces/${rest_resource.monitoring_workspace.response_data.id}/dashboards"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    dashboard = {
      id      = null
      title   = "${var.application_name} - Application Metrics"
      tags    = ["application", "monitoring", var.application_name]
      timezone = "browser"
      panels = [
        {
          id    = 1
          title = "Request Rate"
          type  = "stat"
          targets = [
            {
              expr         = "rate(http_requests_total{job=\"${var.application_name}\"}[5m])"
              datasource   = rest_resource.prometheus_datasource.response_data.name
              refId        = "A"
            }
          ]
          gridPos = { h = 8, w = 12, x = 0, y = 0 }
          options = {
            reduceOptions = {
              values = false
              calcs  = ["lastNotNull"]
              fields = ""
            }
            orientation = "auto"
            textMode    = "auto"
            colorMode   = "value"
          }
        },
        {
          id    = 2
          title = "Response Time (95th percentile)"
          type  = "stat"
          targets = [
            {
              expr         = "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"${var.application_name}\"}[5m]))"
              datasource   = rest_resource.prometheus_datasource.response_data.name
              refId        = "A"
            }
          ]
          gridPos = { h = 8, w = 12, x = 12, y = 0 }
          options = {
            reduceOptions = {
              values = false
              calcs  = ["lastNotNull"]
              fields = ""
            }
            orientation = "auto"
            textMode    = "auto"
            colorMode   = "value"
          }
        },
        {
          id    = 3
          title = "Error Rate"
          type  = "timeseries"
          targets = [
            {
              expr         = "rate(http_requests_total{job=\"${var.application_name}\",status=~\"5..\"}[5m])"
              datasource   = rest_resource.prometheus_datasource.response_data.name
              refId        = "A"
              legendFormat = "5xx errors"
            },
            {
              expr         = "rate(http_requests_total{job=\"${var.application_name}\",status=~\"4..\"}[5m])"
              datasource   = rest_resource.prometheus_datasource.response_data.name
              refId        = "B"
              legendFormat = "4xx errors"
            }
          ]
          gridPos = { h = 8, w = 24, x = 0, y = 8 }
          options = {
            legend = {
              displayMode = "table"
              placement   = "right"
            }
          }
        },
        {
          id    = 4
          title = "Memory Usage"
          type  = "timeseries"
          targets = [
            {
              expr         = "process_resident_memory_bytes{job=\"${var.application_name}\"}"
              datasource   = rest_resource.prometheus_datasource.response_data.name
              refId        = "A"
              legendFormat = "Memory Usage"
            }
          ]
          gridPos = { h = 8, w = 12, x = 0, y = 16 }
        },
        {
          id    = 5
          title = "CPU Usage"
          type  = "timeseries"
          targets = [
            {
              expr         = "rate(process_cpu_seconds_total{job=\"${var.application_name}\"}[5m]) * 100"
              datasource   = rest_resource.prometheus_datasource.response_data.name
              refId        = "A"
              legendFormat = "CPU Usage %"
            }
          ]
          gridPos = { h = 8, w = 12, x = 12, y = 16 }
        }
      ]
      time = {
        from = "now-1h"
        to   = "now"
      }
      refresh = "30s"
    }
    overwrite = false
  })
}

# Create alert rules
resource "rest_resource" "high_error_rate_alert" {
  name     = "${var.application_name}-high-error-rate"
  endpoint = "/api/v1/workspaces/${rest_resource.monitoring_workspace.response_data.id}/alerts"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = "High Error Rate - ${var.application_name}"
    condition = {
      query = {
        queryType = ""
        refId     = "A"
        model = {
          expr         = "rate(http_requests_total{job=\"${var.application_name}\",status=~\"5..\"}[5m]) > ${var.error_rate_threshold}"
          intervalMs   = 1000
          maxDataPoints = 43200
          datasource = {
            type = "prometheus"
            uid  = rest_resource.prometheus_datasource.response_data.uid
          }
        }
      }
      reducer = {
        type = "last"
        params = []
      }
      evaluator = {
        params = [var.error_rate_threshold]
        type   = "gt"
      }
    }
    execution_error_state = "alerting"
    no_data_state        = "no_data"
    for                  = "5m"
    frequency            = "1m"
    handler              = 1
    message              = "Error rate for ${var.application_name} is above ${var.error_rate_threshold}% for more than 5 minutes"
    name                 = "High Error Rate - ${var.application_name}"
    state                = "active"
    notifications = var.alert_notification_ids
  })
}

# Create high latency alert
resource "rest_resource" "high_latency_alert" {
  name     = "${var.application_name}-high-latency"
  endpoint = "/api/v1/workspaces/${rest_resource.monitoring_workspace.response_data.id}/alerts"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = "High Latency - ${var.application_name}"
    condition = {
      query = {
        queryType = ""
        refId     = "A"
        model = {
          expr         = "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"${var.application_name}\"}[5m])) > ${var.latency_threshold}"
          intervalMs   = 1000
          maxDataPoints = 43200
          datasource = {
            type = "prometheus"
            uid  = rest_resource.prometheus_datasource.response_data.uid
          }
        }
      }
      reducer = {
        type = "last"
        params = []
      }
      evaluator = {
        params = [var.latency_threshold]
        type   = "gt"
      }
    }
    execution_error_state = "alerting"
    no_data_state        = "no_data"
    for                  = "5m"
    frequency            = "1m"
    handler              = 1
    message              = "95th percentile latency for ${var.application_name} is above ${var.latency_threshold}s for more than 5 minutes"
    name                 = "High Latency - ${var.application_name}"
    state                = "active"
    notifications = var.alert_notification_ids
  })
}

# Create notification channels
resource "rest_resource" "slack_notification" {
  count = var.enable_slack_notifications ? 1 : 0
  
  name     = "${var.application_name}-slack"
  endpoint = "/api/v1/workspaces/${rest_resource.monitoring_workspace.response_data.id}/notifications"
  create_method = "POST"
  read_method   = "GET"
  update_method = "PUT"
  delete_method = "DELETE"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    name = "Slack - ${var.application_name}"
    type = "slack"
    settings = {
      url       = var.slack_webhook_url
      channel   = var.slack_channel
      username  = "Terraform Monitor"
      iconEmoji = ":warning:"
      title     = "{{ .CommonLabels.alertname }}"
      text      = "{{ .CommonAnnotations.message }}"
    }
  })
}

# Query current alert status
data "rest_data" "alert_status" {
  endpoint = "/api/v1/workspaces/${rest_resource.monitoring_workspace.response_data.id}/alerts/status"
  method   = "GET"
  
  query_params = {
    "state" = "active"
    "application" = var.application_name
  }
}

# Get dashboard metrics
data "rest_data" "dashboard_metrics" {
  endpoint = "/api/v1/workspaces/${rest_resource.monitoring_workspace.response_data.id}/metrics"
  method   = "POST"
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    queries = [
      {
        expr = "rate(http_requests_total{job=\"${var.application_name}\"}[5m])"
        start = "now-1h"
        end   = "now"
        step  = "1m"
      },
      {
        expr = "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"${var.application_name}\"}[5m]))"
        start = "now-1h"
        end   = "now"
        step  = "1m"
      }
    ]
  })
}

# Outputs
output "monitoring_workspace" {
  description = "Monitoring workspace information"
  value = {
    id          = rest_resource.monitoring_workspace.response_data.id
    name        = rest_resource.monitoring_workspace.response_data.name
    url         = rest_resource.monitoring_workspace.response_data.url
    created_at  = rest_resource.monitoring_workspace.created_at
  }
}

output "dashboard_info" {
  description = "Dashboard information"
  value = {
    id    = rest_resource.application_dashboard.response_data.id
    uid   = rest_resource.application_dashboard.response_data.uid
    url   = rest_resource.application_dashboard.response_data.url
    title = rest_resource.application_dashboard.response_data.title
  }
}

output "alerts" {
  description = "Created alerts"
  value = {
    high_error_rate = {
      id      = rest_resource.high_error_rate_alert.response_data.id
      name    = rest_resource.high_error_rate_alert.response_data.name
      state   = rest_resource.high_error_rate_alert.response_data.state
    }
    high_latency = {
      id      = rest_resource.high_latency_alert.response_data.id
      name    = rest_resource.high_latency_alert.response_data.name
      state   = rest_resource.high_latency_alert.response_data.state
    }
  }
}

output "current_metrics" {
  description = "Current application metrics"
  value = {
    request_rate    = data.rest_data.dashboard_metrics.response_data.data[0].values
    latency_p95     = data.rest_data.dashboard_metrics.response_data.data[1].values
  }
}

output "active_alerts" {
  description = "Currently active alerts"
  value = data.rest_data.alert_status.response_data
}

# Variables
variable "monitoring_api_url" {
  description = "Monitoring system API URL"
  type        = string
}

variable "monitoring_token" {
  description = "Monitoring system API token"
  type        = string
  sensitive   = true
}

variable "workspace_name" {
  description = "Name of the monitoring workspace"
  type        = string
}

variable "workspace_description" {
  description = "Description of the monitoring workspace"
  type        = string
  default     = "Terraform managed monitoring workspace"
}

variable "workspace_tags" {
  description = "Tags for the workspace"
  type        = list(string)
  default     = ["terraform", "monitoring"]
}

variable "retention_days" {
  description = "Data retention in days"
  type        = number
  default     = 30
}

variable "sampling_rate" {
  description = "Metrics sampling rate"
  type        = number
  default     = 1.0
}

variable "enable_profiling" {
  description = "Enable performance profiling"
  type        = bool
  default     = false
}

variable "application_name" {
  description = "Name of the application being monitored"
  type        = string
}

variable "prometheus_url" {
  description = "Prometheus server URL"
  type        = string
}

variable "prometheus_basic_auth" {
  description = "Enable basic auth for Prometheus"
  type        = bool
  default     = false
}

variable "prometheus_username" {
  description = "Prometheus basic auth username"
  type        = string
  default     = ""
}

variable "prometheus_password" {
  description = "Prometheus basic auth password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "error_rate_threshold" {
  description = "Error rate threshold for alerts (percentage)"
  type        = number
  default     = 5.0
}

variable "latency_threshold" {
  description = "Latency threshold for alerts (seconds)"
  type        = number
  default     = 2.0
}

variable "alert_notification_ids" {
  description = "List of notification channel IDs for alerts"
  type        = list(string)
  default     = []
}

variable "enable_slack_notifications" {
  description = "Enable Slack notifications"
  type        = bool
  default     = false
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for notifications"
  type        = string
  default     = ""
  sensitive   = true
}

variable "slack_channel" {
  description = "Slack channel for notifications"
  type        = string
  default     = "#alerts"
}