package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"rest": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}


// Simple acceptance test for the data source
func TestAccRestDataSource_Basic(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			fmt.Fprintln(w, `{"message": "Hello, World!", "status": "success"}`)
		} else {
			w.WriteHeader(404)
			fmt.Fprintln(w, `{"error": "Not Found"}`)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccRestDataSourceConfig(server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.rest_data.test", "method", "GET"),
					resource.TestCheckResourceAttr("data.rest_data.test", "endpoint", "/test"),
					resource.TestCheckResourceAttr("data.rest_data.test", "status_code", "200"),
					resource.TestCheckResourceAttrSet("data.rest_data.test", "response"),
					resource.TestCheckResourceAttrSet("data.rest_data.test", "id"),
				),
			},
		},
	})
}

// Simple acceptance test for the resource
func TestAccRestResource_Basic(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	var resourceState string

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch {
		case r.URL.Path == "/items" && r.Method == "POST":
			w.WriteHeader(201)
			fmt.Fprintln(w, `{"id": "test-item-1", "name": "test", "status": "created"}`)
			
		case r.URL.Path == "/items/test" && r.Method == "GET":
			w.WriteHeader(200)
			fmt.Fprintln(w, `{"id": "test-item-1", "name": "test", "status": "active"}`)
			
		case r.URL.Path == "/items/test" && r.Method == "PUT":
			w.WriteHeader(200)
			fmt.Fprintln(w, `{"id": "test-item-1", "name": "test", "status": "updated"}`)
			
		case r.URL.Path == "/items/test" && r.Method == "DELETE":
			w.WriteHeader(204)
			
		default:
			w.WriteHeader(404)
			fmt.Fprintln(w, `{"error": "Not Found"}`)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRestResourceConfig(server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rest_resource.test", "name", "test"),
					resource.TestCheckResourceAttr("rest_resource.test", "endpoint", "/items"),
					resource.TestCheckResourceAttr("rest_resource.test", "method", "POST"),
					resource.TestCheckResourceAttr("rest_resource.test", "status_code", "201"),
					resource.TestCheckResourceAttrSet("rest_resource.test", "response"),
					resource.TestCheckResourceAttrSet("rest_resource.test", "id"),
					resource.TestCheckResourceAttr("rest_resource.test", "response_data.id", "test-item-1"),
					resource.TestCheckResourceAttr("rest_resource.test", "response_data.name", "test"),
					resource.TestCheckResourceAttr("rest_resource.test", "response_data.status", "created"),
					testAccCheckRestResourceExists("rest_resource.test", &resourceState),
				),
			},
			// ImportState testing
			{
				ResourceName:            "rest_resource.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "items/test",
				ImportStateVerifyIgnore: []string{"body", "destroy_body", "created_at", "last_updated", "response", "response_data", "response_headers", "status_code"},
			},
			// Update and Read testing
			{
				Config: testAccRestResourceConfigUpdated(server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rest_resource.test", "name", "test"),
					resource.TestCheckResourceAttr("rest_resource.test", "endpoint", "/items"),
					resource.TestCheckResourceAttrSet("rest_resource.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccRestDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "rest" {
  api_url    = "%s"
  api_token  = "test-token"
  api_header = "Authorization"
}

data "rest_data" "test" {
  endpoint = "/test"
  method   = "GET"
}
`, baseURL)
}

func testAccRestResourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "rest" {
  api_url    = "%s"
  api_token  = "test-token"
  api_header = "Authorization"
}

resource "rest_resource" "test" {
  endpoint = "/items"
  name     = "test"
  method   = "POST"
  body     = jsonencode({
    name = "test"
    type = "example"
  })
}
`, baseURL)
}

func testAccRestResourceConfigUpdated(baseURL string) string {
	return fmt.Sprintf(`
provider "rest" {
  api_url    = "%s"
  api_token  = "test-token"
  api_header = "Authorization"
}

resource "rest_resource" "test" {
  endpoint = "/items"
  name     = "test"
  method   = "POST"
  body     = jsonencode({
    name = "test"
    type = "example-updated"
  })
  update_body = jsonencode({
    name = "test"
    type = "example-updated"
    status = "active"
  })
}
`, baseURL)
}

func testAccCheckRestResourceExists(resourceName string, resourceState *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID not set")
		}

		*resourceState = rs.Primary.ID
		return nil
	}
}

// Test with custom headers and query parameters
func TestAccRestResource_WithHeaders(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Check for custom headers
		if r.Header.Get("X-Custom-Header") != "test-value" {
			w.WriteHeader(400)
			fmt.Fprintln(w, `{"error": "Missing custom header"}`)
			return
		}
		
		// Check for query parameters
		if r.URL.Query().Get("filter") != "active" {
			w.WriteHeader(400)
			fmt.Fprintln(w, `{"error": "Missing query parameter"}`)
			return
		}
		
		switch {
		case r.URL.Path == "/api/users" && r.Method == "POST":
			w.WriteHeader(201)
			fmt.Fprintln(w, `{"id": "user-123", "name": "John Doe", "status": "active"}`)
		case r.URL.Path == "/api/users/test-user" && r.Method == "GET":
			w.WriteHeader(200)
			fmt.Fprintln(w, `{"id": "user-123", "name": "John Doe", "status": "active"}`)
		case r.URL.Path == "/api/users/test-user" && r.Method == "PUT":
			w.WriteHeader(200)
			fmt.Fprintln(w, `{"id": "user-123", "name": "John Doe", "status": "updated"}`)
		case r.URL.Path == "/api/users/test-user" && r.Method == "DELETE":
			w.WriteHeader(204)
		default:
			w.WriteHeader(404)
			fmt.Fprintln(w, `{"error": "Not Found"}`)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRestResourceConfigWithHeaders(server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rest_resource.test", "name", "test-user"),
					resource.TestCheckResourceAttr("rest_resource.test", "endpoint", "/api/users"),
					resource.TestCheckResourceAttr("rest_resource.test", "method", "POST"),
					resource.TestCheckResourceAttr("rest_resource.test", "headers.X-Custom-Header", "test-value"),
					resource.TestCheckResourceAttr("rest_resource.test", "query_params.filter", "active"),
					resource.TestCheckResourceAttr("rest_resource.test", "status_code", "201"),
				),
			},
		},
	})
}

// Test with different HTTP methods
func TestAccRestResource_HTTPMethods(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch {
		case r.URL.Path == "/api/items" && r.Method == "PATCH":
			w.WriteHeader(200)
			fmt.Fprintln(w, `{"id": "item-456", "name": "Updated Item", "status": "modified"}`)
		case r.URL.Path == "/api/items/patch-test" && r.Method == "GET":
			w.WriteHeader(200)
			fmt.Fprintln(w, `{"id": "item-456", "name": "Updated Item", "status": "modified"}`)
		case r.URL.Path == "/api/items/patch-test" && r.Method == "DELETE":
			w.WriteHeader(204)
		default:
			w.WriteHeader(404)
			fmt.Fprintln(w, `{"error": "Not Found"}`)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRestResourceConfigPatch(server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rest_resource.test", "name", "patch-test"),
					resource.TestCheckResourceAttr("rest_resource.test", "endpoint", "/api/items"),
					resource.TestCheckResourceAttr("rest_resource.test", "method", "PATCH"),
					resource.TestCheckResourceAttr("rest_resource.test", "status_code", "200"),
				),
			},
		},
	})
}

// Test error handling and retries
func TestAccRestResource_ErrorHandling(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	retryCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path == "/api/flaky" && r.Method == "POST" {
			retryCount++
			if retryCount <= 2 {
				// Simulate temporary server error
				w.WriteHeader(503)
				fmt.Fprintln(w, `{"error": "Service temporarily unavailable"}`)
				return
			}
			// Success after retries
			w.WriteHeader(201)
			fmt.Fprintln(w, `{"id": "flaky-123", "name": "Eventually Successful", "status": "created"}`)
		} else if r.URL.Path == "/api/flaky/retry-test" && r.Method == "GET" {
			w.WriteHeader(200)
			fmt.Fprintln(w, `{"id": "flaky-123", "name": "Eventually Successful", "status": "active"}`)
		} else if r.URL.Path == "/api/flaky/retry-test" && r.Method == "DELETE" {
			w.WriteHeader(204)
		} else {
			w.WriteHeader(404)
			fmt.Fprintln(w, `{"error": "Not Found"}`)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRestResourceConfigRetry(server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rest_resource.test", "name", "retry-test"),
					resource.TestCheckResourceAttr("rest_resource.test", "endpoint", "/api/flaky"),
					resource.TestCheckResourceAttr("rest_resource.test", "method", "POST"),
					resource.TestCheckResourceAttr("rest_resource.test", "status_code", "201"),
				),
			},
		},
	})
}

func testAccRestResourceConfigWithHeaders(baseURL string) string {
	return fmt.Sprintf(`
provider "rest" {
  api_url       = "%s"
  api_token     = "test-token"
  api_header    = "Authorization"
  retry_attempts = 3
}

resource "rest_resource" "test" {
  endpoint = "/api/users"
  name     = "test-user"
  method   = "POST"
  headers = {
    "X-Custom-Header" = "test-value"
    "Content-Type"    = "application/json"
  }
  query_params = {
    "filter" = "active"
    "limit"  = "10"
  }
  body = jsonencode({
    name = "John Doe"
    email = "john@example.com"
    role = "user"
  })
}
`, baseURL)
}

func testAccRestResourceConfigPatch(baseURL string) string {
	return fmt.Sprintf(`
provider "rest" {
  api_url    = "%s"
  api_token  = "test-token"
  api_header = "Authorization"
}

resource "rest_resource" "test" {
  endpoint = "/api/items"
  name     = "patch-test"
  method   = "PATCH"
  body     = jsonencode({
    name = "Updated Item"
    status = "modified"
  })
}
`, baseURL)
}

func testAccRestResourceConfigRetry(baseURL string) string {
	return fmt.Sprintf(`
provider "rest" {
  api_url        = "%s"
  api_token      = "test-token"
  api_header     = "Authorization"
  retry_attempts = 5
  timeout        = 30
}

resource "rest_resource" "test" {
  endpoint = "/api/flaky"
  name     = "retry-test"
  method   = "POST"
  body     = jsonencode({
    name = "Eventually Successful"
    type = "test"
  })
}
`, baseURL)
}