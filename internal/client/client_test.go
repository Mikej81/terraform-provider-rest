package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRestClient(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid configuration",
			config: Config{
				BaseURL:     "https://api.example.com",
				Token:       "test-token",
				TokenHeader: "Authorization",
			},
			expectErr: false,
		},
		{
			name: "missing base URL",
			config: Config{
				Token:       "test-token",
				TokenHeader: "Authorization",
			},
			expectErr: true,
			errMsg:    "base URL is required",
		},
		{
			name: "invalid base URL",
			config: Config{
				BaseURL:     "not-a-url",
				Token:       "test-token",
				TokenHeader: "Authorization",
			},
			expectErr: true,
			errMsg:    "base URL must include a scheme",
		},
		{
			name: "with custom settings",
			config: Config{
				BaseURL:           "https://api.example.com",
				Token:             "test-token",
				TokenHeader:       "X-API-Key",
				Timeout:           10 * time.Second,
				RetryAttempts:     5,
				MaxIdleConns:      50,
				IdleConnTimeout:   60 * time.Second,
				DisableKeepAlives: true,
				Insecure:          true,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewRestClient(tt.config)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			if client == nil {
				t.Error("Expected client to be created")
				return
			}

			// Test that defaults are set
			if client.timeout == 0 {
				t.Error("Expected timeout to be set")
			}
			if client.retries == 0 {
				t.Error("Expected retries to be set")
			}
			if client.userAgent == "" {
				t.Error("Expected user agent to be set")
			}
		})
	}
}

func TestRestClient_Do(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		body           []byte
		headers        map[string]string
		queryParams    map[string]string
		serverResponse int
		serverBody     string
		expectErr      bool
	}{
		{
			name:           "successful GET request",
			method:         "GET",
			endpoint:       "/users",
			serverResponse: 200,
			serverBody:     `{"users": []}`,
			expectErr:      false,
		},
		{
			name:           "successful POST request with body",
			method:         "POST",
			endpoint:       "/users",
			body:           []byte(`{"name": "test"}`),
			serverResponse: 201,
			serverBody:     `{"id": "123", "name": "test"}`,
			expectErr:      false,
		},
		{
			name:           "request with custom headers",
			method:         "GET",
			endpoint:       "/users",
			headers:        map[string]string{"X-Custom": "value"},
			serverResponse: 200,
			serverBody:     `{"users": []}`,
			expectErr:      false,
		},
		{
			name:           "request with query parameters",
			method:         "GET",
			endpoint:       "/users",
			queryParams:    map[string]string{"limit": "10", "offset": "0"},
			serverResponse: 200,
			serverBody:     `{"users": []}`,
			expectErr:      false,
		},
		{
			name:           "all HTTP methods",
			method:         "PUT",
			endpoint:       "/users/123",
			body:           []byte(`{"name": "updated"}`),
			serverResponse: 200,
			serverBody:     `{"id": "123", "name": "updated"}`,
			expectErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method
				if r.Method != tt.method {
					t.Errorf("Expected method %s, got %s", tt.method, r.Method)
				}

				// Verify endpoint
				expectedPath := tt.endpoint
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify query parameters
				if tt.queryParams != nil {
					query := r.URL.Query()
					for key, expectedValue := range tt.queryParams {
						if actualValue := query.Get(key); actualValue != expectedValue {
							t.Errorf("Expected query param %s=%s, got %s", key, expectedValue, actualValue)
						}
					}
				}

				// Verify custom headers
				if tt.headers != nil {
					for key, expectedValue := range tt.headers {
						if actualValue := r.Header.Get(key); actualValue != expectedValue {
							t.Errorf("Expected header %s=%s, got %s", key, expectedValue, actualValue)
						}
					}
				}

				// Verify authorization header is set
				if auth := r.Header.Get("Authorization"); auth == "" {
					t.Error("Expected Authorization header to be set")
				}

				w.WriteHeader(tt.serverResponse)
				w.Write([]byte(tt.serverBody))
			}))
			defer server.Close()

			// Create client
			client, err := NewRestClient(Config{
				BaseURL:     server.URL,
				Token:       "test-token",
				TokenHeader: "Authorization",
			})
			if err != nil {
				t.Fatalf("Failed to create client: %s", err)
			}

			// Make request
			options := RequestOptions{
				Method:      tt.method,
				Endpoint:    tt.endpoint,
				Body:        tt.body,
				Headers:     tt.headers,
				QueryParams: tt.queryParams,
			}

			ctx := context.Background()
			response, err := client.Do(ctx, options)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			if response.StatusCode != tt.serverResponse {
				t.Errorf("Expected status code %d, got %d", tt.serverResponse, response.StatusCode)
			}

			if string(response.Body) != tt.serverBody {
				t.Errorf("Expected body %s, got %s", tt.serverBody, string(response.Body))
			}
		})
	}
}

func TestRestClient_Retry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// Fail first two attempts with retryable status code
			w.WriteHeader(500)
			w.Write([]byte("Server Error"))
		} else {
			// Succeed on third attempt
			w.WriteHeader(200)
			w.Write([]byte("Success"))
		}
	}))
	defer server.Close()

	client, err := NewRestClient(Config{
		BaseURL:       server.URL,
		Token:         "test-token",
		TokenHeader:   "Authorization",
		RetryAttempts: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %s", err)
	}

	ctx := context.Background()
	response, err := client.Do(ctx, RequestOptions{
		Method:   "GET",
		Endpoint: "/test",
		Retries:  3, // Override client default to ensure retries happen
	})

	if err != nil {
		t.Errorf("Unexpected error after retries: %s", err)
		return
	}

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	if string(response.Body) != "Success" {
		t.Errorf("Expected body 'Success', got %s", string(response.Body))
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestRestClient_BuildURL(t *testing.T) {
	client, _ := NewRestClient(Config{
		BaseURL:     "https://api.example.com",
		Token:       "test",
		TokenHeader: "Authorization",
	})

	tests := []struct {
		name        string
		endpoint    string
		queryParams map[string]string
		expected    string
	}{
		{
			name:     "simple endpoint",
			endpoint: "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "endpoint with leading slash removed",
			endpoint: "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:        "endpoint with query parameters",
			endpoint:    "/users",
			queryParams: map[string]string{"limit": "10", "offset": "0"},
			expected:    "https://api.example.com/users?limit=10&offset=0",
		},
		{
			name:        "endpoint with multiple query parameters",
			endpoint:    "/users",
			queryParams: map[string]string{"limit": "10", "sort": "name", "order": "asc"},
			expected:    "https://api.example.com/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.buildURL(tt.endpoint, tt.queryParams)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			// For query parameters, we need to check they're present rather than exact match
			// since map iteration order is not guaranteed
			if tt.queryParams != nil {
				if !strings.Contains(result, "https://api.example.com/users?") {
					t.Errorf("Expected URL to contain base with query params, got %s", result)
				}
				for key, value := range tt.queryParams {
					expectedParam := key + "=" + value
					if !strings.Contains(result, expectedParam) {
						t.Errorf("Expected URL to contain %s, got %s", expectedParam, result)
					}
				}
			} else {
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestRestClient_IsRetryableError(t *testing.T) {
	client, _ := NewRestClient(Config{
		BaseURL:     "https://api.example.com",
		Token:       "test",
		TokenHeader: "Authorization",
	})

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRestClient_IsRetryableStatusCode(t *testing.T) {
	client, _ := NewRestClient(Config{
		BaseURL:     "https://api.example.com",
		Token:       "test",
		TokenHeader: "Authorization",
	})

	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{name: "200 OK", statusCode: 200, expected: false},
		{name: "400 Bad Request", statusCode: 400, expected: false},
		{name: "429 Too Many Requests", statusCode: 429, expected: true},
		{name: "500 Internal Server Error", statusCode: 500, expected: true},
		{name: "502 Bad Gateway", statusCode: 502, expected: true},
		{name: "503 Service Unavailable", statusCode: 503, expected: true},
		{name: "504 Gateway Timeout", statusCode: 504, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.IsRetryableStatusCode(tt.statusCode)
			if result != tt.expected {
				t.Errorf("Expected %v for status code %d, got %v", tt.expected, tt.statusCode, result)
			}
		})
	}
}

func TestRestClient_Timeout(t *testing.T) {
	// Create a server that takes longer than our timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(200)
	}))
	defer server.Close()

	client, err := NewRestClient(Config{
		BaseURL:     server.URL,
		Token:       "test-token",
		TokenHeader: "Authorization",
		Timeout:     100 * time.Millisecond, // Very short timeout
	})
	if err != nil {
		t.Fatalf("Failed to create client: %s", err)
	}

	ctx := context.Background()
	_, err = client.Do(ctx, RequestOptions{
		Method:   "GET",
		Endpoint: "/test",
	})

	if err == nil {
		t.Error("Expected timeout error but got none")
	}
}
