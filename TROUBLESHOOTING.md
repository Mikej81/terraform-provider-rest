# Troubleshooting Guide

## Quick Diagnosis

### **Step 1: Enable Debug Logging**

```bash
export TF_LOG=DEBUG
terraform plan
```

This shows you exactly what HTTP requests are being made and what responses are received.

### **Step 2: Check Basic Connectivity**

```terraform
# Test with a simple data source first
data "rest_data" "health_check" {
  endpoint = "/health"
  method   = "GET"
}

output "health_status" {
  value = data.rest_data.health_check.response_data
}
```

If this works, your authentication and base configuration are correct.

### **Step 3: Test Authentication**

```bash
# Test your API directly with curl
curl -H "Authorization: Bearer $YOUR_TOKEN" \
  "https://your-api.com/health"
```

Compare this with what Terraform is doing.

## Common Issues

### **Authentication Problems**

#### **Issue: 401 Unauthorized**

**Symptoms:**
- `Error: HTTP request failed: 401 Unauthorized`
- Resources fail to create or read

**Solutions:**

1. **Check your token/credentials:**
   ```terraform
   provider "rest" {
     api_url   = "https://api.example.com"
     api_token = var.api_token  # Make sure this variable is set
   }
   ```

2. **Verify the header name:**
   ```terraform
   provider "rest" {
     api_url    = "https://api.example.com"
     api_token  = var.api_token
     api_header = "X-API-Key"  # Some APIs use custom headers
   }
   ```

3. **Check token format:**
   ```terraform
   # Some APIs need "Bearer " prefix
   provider "rest" {
     api_url   = "https://api.example.com"
     api_token = "Bearer ${var.api_token}"
   }
   
   # Others use the token directly
   provider "rest" {
     api_url   = "https://api.example.com"
     api_token = var.api_token
   }
   ```

4. **Test your credentials manually:**
   ```bash
   # Test with curl
   curl -H "Authorization: Bearer $TOKEN" \
     "https://api.example.com/users"
   
   # Check the response
   ```

#### **Issue: 403 Forbidden**

**Symptoms:**
- Authentication works but specific operations fail
- Some endpoints work, others don't

**Solutions:**

1. **Check API permissions:**
   - Verify your token has the required scopes
   - Check if your user/token has permission for the specific endpoint

2. **Review API documentation:**
   - Some endpoints require admin privileges
   - Some operations need specific permissions

### **Network and SSL Issues**

#### **Issue: SSL Certificate Errors**

**Symptoms:**
- `Error: x509: certificate signed by unknown authority`
- `Error: x509: certificate has expired`

**Solutions:**

1. **For development/testing only:**
   ```terraform
   provider "rest" {
     api_url  = "https://api.example.com"
     api_token = var.api_token
     insecure = true  # Never use in production
   }
   ```

2. **For production - fix the certificate:**
   ```bash
   # Check certificate details
   openssl s_client -connect api.example.com:443 -servername api.example.com
   
   # Update system certificates
   sudo apt-get update && sudo apt-get install ca-certificates
   ```

#### **Issue: Timeout Errors**

**Symptoms:**
- `Error: context deadline exceeded`
- Resources take too long to respond

**Solutions:**

1. **Increase timeout:**
   ```terraform
   provider "rest" {
     api_url = "https://api.example.com"
     timeout = 120  # 2 minutes instead of default 30 seconds
   }
   
   # Or per resource
   resource "rest_resource" "slow_operation" {
     name     = "slow-resource"
     endpoint = "/slow-endpoint"
     timeout  = 300  # 5 minutes for this specific resource
     body     = jsonencode({data = "value"})
   }
   ```

2. **Increase retry attempts:**
   ```terraform
   provider "rest" {
     api_url        = "https://api.example.com"
     retry_attempts = 10  # Retry more times
   }
   ```

### **Data and Response Issues**

#### **Issue: Empty or Null Response Data**

**Symptoms:**
- `data.rest_data.example.response_data` is empty
- `rest_resource.example.response_data.id` is null

**Solutions:**

1. **Check the raw response:**
   ```terraform
   output "debug_response" {
     value = {
       raw_response = data.rest_data.example.response
       status_code  = data.rest_data.example.status_code
       headers      = data.rest_data.example.response_headers
     }
   }
   ```

2. **Verify JSON format:**
   ```bash
   # Test your API response format
   curl -s "https://api.example.com/endpoint" | jq '.'
   ```

3. **Check for different response structure:**
   ```terraform
   # If API returns: {"data": {"id": "123"}}
   # Access with: response_data.data.id
   
   # If API returns: [{"id": "123"}]
   # Access with: response_data[0].id
   ```

#### **Issue: Drift Detection Problems**

**Symptoms:**
- Terraform says resource changed when you didn't change anything
- Terraform doesn't detect actual changes

**Solutions:**

1. **Add fields to ignore list:**
   ```terraform
   resource "rest_resource" "user" {
     name     = "john-doe"
     endpoint = "/users"
     body     = jsonencode({name = "John Doe"})
     
     # Ignore server-added fields
     ignore_fields = [
       "last_login_at",
       "login_count",
       "updated_at"
     ]
   }
   ```

2. **Disable drift detection for dynamic resources:**
   ```terraform
   resource "rest_resource" "metrics" {
     name     = "cpu-metrics"
     endpoint = "/metrics"
     body     = jsonencode({metric = "cpu_usage"})
     
     # Disable for constantly changing data
     drift_detection = false
   }
   ```

3. **Debug drift detection:**
   ```bash
   export TF_LOG=DEBUG
   terraform plan 2>&1 | grep -i "drift\|comparing\|ignoring"
   ```

See [DRIFT_DETECTION.md](DRIFT_DETECTION.md) for detailed information.

### **Resource Management Issues**

#### **Issue: Resource Not Found on Read**

**Symptoms:**
- `Error: HTTP request failed: 404 Not Found`
- Resource shows as "will be created" when it should exist

**Solutions:**

1. **Check the resource name:**
   ```terraform
   resource "rest_resource" "user" {
     name     = "john-doe"  # This must match your API's identifier
     endpoint = "/users"
     body     = jsonencode({name = "John Doe"})
   }
   ```

2. **Verify the read endpoint:**
   ```bash
   # Test the read endpoint manually
   curl "https://api.example.com/users/john-doe"
   ```

3. **Check if your API uses different identifiers:**
   ```terraform
   # Some APIs use IDs instead of names
   resource "rest_resource" "user" {
     name     = "123"  # Use the ID from the API
     endpoint = "/users"
     body     = jsonencode({name = "John Doe"})
   }
   ```

#### **Issue: Import Problems**

**Symptoms:**
- `terraform import` fails
- Imported resource shows as changed immediately

**Solutions:**

1. **Use correct import format:**
   ```bash
   # Format: endpoint/name
   terraform import rest_resource.user "users/john-doe"
   ```

2. **Match configuration to current state:**
   ```bash
   # After import, check current state
   terraform show
   
   # Write configuration to match
   ```

3. **Check for drift after import:**
   ```bash
   terraform plan
   # Should show no changes after import
   ```

### **API-Specific Issues**

#### **Issue: Rate Limiting**

**Symptoms:**
- `Error: HTTP request failed: 429 Too Many Requests`
- Random failures during apply

**Solutions:**

1. **Increase retry attempts:**
   ```terraform
   provider "rest" {
     api_url        = "https://api.example.com"
     retry_attempts = 10
   }
   ```

2. **Add delays between resources:**
   ```terraform
   resource "rest_resource" "user1" {
     name     = "user1"
     endpoint = "/users"
     body     = jsonencode({name = "User 1"})
   }
   
   resource "time_sleep" "wait" {
     depends_on      = [rest_resource.user1]
     create_duration = "5s"
   }
   
   resource "rest_resource" "user2" {
     depends_on = [time_sleep.wait]
     name       = "user2"
     endpoint   = "/users"
     body       = jsonencode({name = "User 2"})
   }
   ```

#### **Issue: API Returns Different Data Format**

**Symptoms:**
- Expected JSON but got XML/HTML
- Response parsing fails

**Solutions:**

1. **Check Content-Type header:**
   ```terraform
   data "rest_data" "example" {
     endpoint = "/data"
     headers = {
       "Accept" = "application/json"
     }
   }
   ```

2. **Verify API endpoint:**
   ```bash
   # Check what your API actually returns
   curl -H "Accept: application/json" \
     "https://api.example.com/data"
   ```

3. **Handle non-JSON responses:**
   ```terraform
   # Use the raw response for non-JSON APIs
   output "raw_data" {
     value = data.rest_data.example.response
   }
   ```

## Debugging Workflow

### **Step-by-Step Debugging**

1. **Enable debug logging:**
   ```bash
   export TF_LOG=DEBUG
   ```

2. **Test basic connectivity:**
   ```terraform
   data "rest_data" "test" {
     endpoint = "/health"
   }
   ```

3. **Run plan and check output:**
   ```bash
   terraform plan
   ```

4. **Look for specific error patterns:**
   ```bash
   terraform plan 2>&1 | grep -i "error\|failed\|timeout"
   ```

5. **Test API directly:**
   ```bash
   curl -v -H "Authorization: Bearer $TOKEN" \
     "https://api.example.com/your-endpoint"
   ```

6. **Compare Terraform requests with manual requests:**
   ```bash
   # From debug logs, find the actual HTTP request Terraform makes
   # Compare with your manual curl request
   ```

### **Common Debug Log Patterns**

**Authentication issues:**
```
[DEBUG] HTTP Request: GET /users/123 HTTP/1.1
[DEBUG] HTTP Response: 401 Unauthorized
```

**Network issues:**
```
[DEBUG] Error: Post "https://api.example.com/users": context deadline exceeded
```

**JSON parsing issues:**
```
[DEBUG] Response body: <html><body>Error</body></html>
[DEBUG] Error parsing JSON response
```

**Drift detection:**
```
[DEBUG] Comparing field: name ("John Doe" == "John Doe")
[DEBUG] Ignoring field: created_at (default ignore list)
[DEBUG] Drift detected in field: login_count (0 != 5)
```

## Getting Help

### **Information to Include**

When asking for help, provide:

1. **Your Terraform configuration:**
   ```terraform
   # Provider configuration
   provider "rest" {
     api_url   = "https://api.example.com"
     api_token = var.api_token
   }
   
   # Resource configuration
   resource "rest_resource" "example" {
     name     = "my-resource"
     endpoint = "/resources"
     body     = jsonencode({data = "value"})
   }
   ```

2. **Error messages:**
   ```
   Error: HTTP request failed: 401 Unauthorized
   ```

3. **Debug logs (sanitized):**
   ```bash
   export TF_LOG=DEBUG
   terraform plan 2>&1 | grep -A 5 -B 5 "your-resource"
   ```

4. **Expected vs actual behavior:**
   ```
   Expected: Resource should be created successfully
   Actual: Getting 401 Unauthorized error
   ```

5. **API documentation or curl example:**
   ```bash
   # Working curl command
   curl -H "Authorization: Bearer token" \
     -X POST \
     -H "Content-Type: application/json" \
     -d '{"data": "value"}' \
     "https://api.example.com/resources"
   ```

### **Support Resources**

- **[GitHub Issues](https://github.com/Mikej81/terraform-provider-rest/issues)**: Bug reports and feature requests
- **[GitHub Discussions](https://github.com/Mikej81/terraform-provider-rest/discussions)**: Questions and community help
- **[Documentation](README.md)**: Complete usage guide
- **[Examples](examples/)**: Real-world configuration examples

### **Before Opening an Issue**

1. **Search existing issues** for similar problems
2. **Try the debugging steps** in this guide
3. **Test your API** with curl to confirm it works
4. **Provide minimal reproduction** case
5. **Include all relevant information** listed above

Remember: The more information you provide, the faster we can help you solve the problem!