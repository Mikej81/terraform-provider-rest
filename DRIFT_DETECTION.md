# Drift Detection in terraform-provider-rest

This provider includes comprehensive drift detection capabilities to handle the common scenario where APIs add server-side metadata fields that shouldn't trigger configuration drift.

## Overview

When APIs return additional metadata fields (like `created_at`, `updated_at`, `id`, etc.) that weren't in the original request, traditional Terraform providers might detect this as configuration drift. This provider intelligently handles such scenarios.

## Key Features

### 1. **Automatic Server Metadata Handling**
The provider automatically ignores common server-side metadata fields:

```hcl
# These fields are ignored by default during drift detection:
# - id, created_at, updated_at, last_modified, etag
# - version, revision, timestamp, last_updated_at  
# - created_by, updated_by, modified_by, owner_id
# - _id, _created, _updated, _modified, _version
# - createdAt, updatedAt, lastModified, lastUpdated
# - href, self, links, _links, meta, _meta
```

### 2. **Custom Ignore Fields**
Specify additional fields to ignore during drift detection:

```hcl
resource "rest_resource" "example" {
  endpoint = "/api/users"
  name     = "john-doe"
  
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
  })
  
  # Ignore custom server-side fields
  ignore_fields = [
    "last_login_time",
    "login_count", 
    "computed_score",
    "external_references"
  ]
}
```

### 3. **Configurable Drift Detection**
Enable or disable drift detection entirely:

```hcl
resource "rest_resource" "analytics" {
  endpoint = "/analytics/data"
  name     = "user-metrics"
  
  body = jsonencode({
    metric_name = "user_engagement"
    enabled     = true
  })
  
  # Disable drift detection for frequently changing data
  drift_detection = false
}
```

## Common Use Cases

### API Response Enhancement
Many APIs enhance responses with server-side metadata:

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "role": "admin"
}
```

**API Response:**
```json
{
  "id": "user-12345",
  "name": "John Doe", 
  "email": "john@example.com",
  "role": "admin",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z",
  "last_login": null,
  "account_status": "active",
  "profile_url": "/users/user-12345"
}
```

With this provider, the additional fields (`id`, `created_at`, `updated_at`, `last_login`, `account_status`, `profile_url`) are automatically ignored, preventing false drift detection.

### Dynamic Content APIs
For APIs that return dynamic content:

```hcl
resource "rest_resource" "dashboard_config" {
  endpoint = "/dashboards"
  name     = "sales-dashboard"
  
  body = jsonencode({
    title   = "Sales Dashboard"
    widgets = ["revenue", "leads", "conversion"]
  })
  
  ignore_fields = [
    "last_viewed_at",    # Updates when users view dashboard
    "view_count",        # Increments on each view
    "cache_timestamp",   # Updates when cache refreshes
    "computed_metrics"   # Server calculates dynamic metrics
  ]
}
```

## Drift Detection Logic

### 1. **Structural Comparison**
The provider performs deep comparison of JSON structures:
- Recursively compares nested objects
- Handles array comparisons
- Ignores specified fields at any nesting level

### 2. **Type-Safe Comparisons**
Handles different numeric representations:
- Compares `1` (int) and `1.0` (float) as equal
- Handles JSON unmarshaling type variations
- Preserves null vs undefined distinctions

### 3. **Logging and Debugging**
Comprehensive logging for troubleshooting:
```bash
# Enable debug logging
export TF_LOG=DEBUG

# Drift detection logs include:
# - Field-level comparison results
# - Ignored field notifications  
# - Detected drift warnings
# - State update confirmations
```

## Configuration Reference

### Schema Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `ignore_fields` | `list(string)` | Additional field names to ignore during drift detection |
| `drift_detection` | `bool` | Enable/disable drift detection (default: `true`) |

### Default Ignored Fields

The provider automatically ignores these common patterns:
- **IDs**: `id`, `_id`, `external_id`
- **Timestamps**: `created_at`, `updated_at`, `last_modified`, `timestamp`
- **Versions**: `version`, `revision`, `etag`, `_version`
- **User Info**: `created_by`, `updated_by`, `modified_by`, `owner_id`
- **Metadata**: `meta`, `_meta`, `href`, `self`, `links`, `_links`
- **CamelCase**: `createdAt`, `updatedAt`, `lastModified`, `lastUpdated`

## Best Practices

### 1. **Start with Defaults**
The default ignore list covers most common scenarios:

```hcl
resource "rest_resource" "simple" {
  endpoint = "/api/items"
  name     = "item-1"
  body     = jsonencode({ name = "Item 1" })
  # drift_detection defaults to true
  # Default ignore fields automatically applied
}
```

### 2. **Add Custom Fields Gradually**
Start with defaults, then add custom ignores as needed:

```hcl
resource "rest_resource" "custom" {
  endpoint = "/api/items"
  name     = "item-1"
  body     = jsonencode({ name = "Item 1" })
  
  # Add API-specific fields to ignore
  ignore_fields = ["custom_score", "recommendation_id"]
}
```

### 3. **Use Descriptive Names**
Make ignore field purposes clear:

```hcl
ignore_fields = [
  "server_computed_hash",    # Server calculates content hash
  "auto_generated_slug",     # Server creates URL-friendly name
  "system_assigned_priority" # Server assigns based on load
]
```

### 4. **Consider Disabling for Dynamic APIs**
For APIs that return constantly changing data:

```hcl
resource "rest_resource" "metrics" {
  endpoint = "/metrics/realtime"
  name     = "cpu-usage"
  body     = jsonencode({ metric = "cpu_percent" })
  
  # Disable drift detection for real-time data
  drift_detection = false
}
```

## Troubleshooting

### Enable Debug Logging
```bash
export TF_LOG=DEBUG
terraform plan
```

### Common Issues

1. **False Drift Detection**
   - **Symptom**: Terraform detects changes when none were made
   - **Solution**: Add fields to `ignore_fields` list

2. **Missing Drift Detection**
   - **Symptom**: Actual changes not detected
   - **Solution**: Verify `drift_detection = true` and check ignore fields

3. **Numeric Comparison Issues**
   - **Symptom**: Numbers like `1` vs `1.0` cause drift
   - **Solution**: This is handled automatically by the provider

## Examples

See `examples/drift_detection.tf` for comprehensive examples covering:
- Basic drift detection with defaults
- Custom ignore fields configuration  
- Disabled drift detection scenarios
- Advanced configurations with conditional operations
- Real-world API integration patterns

## Migration Guide

### From Basic REST Providers
If migrating from basic REST providers, enable drift detection:

```hcl
# Old configuration
resource "old_rest_resource" "example" {
  endpoint = "/api/items"
  body     = jsonencode({ name = "Item" })
}

# New configuration with drift detection
resource "rest_resource" "example" {
  endpoint = "/api/items" 
  name     = "item-name"  # Required for CRUD operations
  body     = jsonencode({ name = "Item" })
  # drift_detection = true (default)
  # Default ignore fields automatically applied
}
```

This enhancement makes the terraform-provider-rest robust for production use with real-world APIs that add server-side metadata.