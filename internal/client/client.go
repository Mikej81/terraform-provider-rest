package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/crypto/pkcs12"
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// RestClient provides a robust HTTP client for REST operations
type RestClient struct {
	baseURL    string
	httpClient HTTPClient
	headers    map[string]string
	timeout    time.Duration
	retries    int
	userAgent  string
}

// Config holds the configuration for the REST client
type Config struct {
	BaseURL string
	// Token Authentication
	Token       string
	TokenHeader string
	// Certificate Authentication
	ClientCert     string // PEM-encoded client certificate
	ClientKey      string // PEM-encoded client key
	ClientCertFile string // Path to client certificate file
	ClientKeyFile  string // Path to client key file
	// PKCS12 Authentication
	PKCS12Bundle   string // Base64-encoded PKCS12 bundle
	PKCS12File     string // Path to PKCS12 file
	PKCS12Password string // Password for PKCS12 bundle
	// General Options
	Timeout           time.Duration
	Insecure          bool
	RetryAttempts     int
	CustomHeaders     map[string]string
	UserAgent         string
	MaxIdleConns      int
	IdleConnTimeout   time.Duration
	DisableKeepAlives bool
}

// NewRestClient creates a new REST client with the provided configuration
func NewRestClient(config Config) (*RestClient, error) {
	// Validate base URL
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Ensure URL has a scheme
	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("base URL must include a scheme (http or https)")
	}

	// Set default values
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.UserAgent == "" {
		config.UserAgent = "terraform-provider-rest/1.0"
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 100
	}
	if config.IdleConnTimeout == 0 {
		config.IdleConnTimeout = 90 * time.Second
	}

	// Create HTTP transport
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.Insecure,
	}

	// Handle certificate authentication
	if err := configureTLSAuth(config, tlsConfig); err != nil {
		return nil, fmt.Errorf("failed to configure TLS authentication: %w", err)
	}

	transport.TLSClientConfig = tlsConfig

	// Create HTTP client
	httpClient := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	// Initialize headers
	headers := make(map[string]string)
	headers["User-Agent"] = config.UserAgent
	headers["Accept"] = "application/json"
	headers["Content-Type"] = "application/json"

	// Add authentication header if provided
	if config.Token != "" && config.TokenHeader != "" {
		headers[config.TokenHeader] = config.Token
	}

	// Add custom headers
	for k, v := range config.CustomHeaders {
		headers[k] = v
	}

	return &RestClient{
		baseURL:    strings.TrimRight(config.BaseURL, "/"),
		httpClient: httpClient,
		headers:    headers,
		timeout:    config.Timeout,
		retries:    config.RetryAttempts,
		userAgent:  config.UserAgent,
	}, nil
}

// RequestOptions holds options for HTTP requests
type RequestOptions struct {
	Method      string
	Endpoint    string
	Body        []byte
	Headers     map[string]string
	QueryParams map[string]string
	Timeout     time.Duration
	Retries     int
}

// Response holds the HTTP response data
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
	Request    *http.Request
}

// Do executes an HTTP request with retry logic and proper error handling
func (c *RestClient) Do(ctx context.Context, options RequestOptions) (*Response, error) {
	// Build full URL
	fullURL, err := c.buildURL(options.Endpoint, options.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Create request body
	var body io.Reader
	if options.Body != nil {
		body = bytes.NewReader(options.Body)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, options.Method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	c.setHeaders(req, options.Headers)

	// Set timeout if specified
	timeout := c.timeout
	if options.Timeout > 0 {
		timeout = options.Timeout
	}

	// Set retry attempts if specified
	retries := c.retries
	if options.Retries > 0 {
		retries = options.Retries
	}

	// Execute request with retry logic
	return c.executeWithRetry(ctx, req, retries, timeout)
}

// buildURL constructs the full URL with query parameters
func (c *RestClient) buildURL(endpoint string, queryParams map[string]string) (string, error) {
	// Clean endpoint
	endpoint = strings.TrimPrefix(endpoint, "/")

	// Build base URL
	fullURL := fmt.Sprintf("%s/%s", c.baseURL, endpoint)

	// Parse URL for query parameters
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}

	// Add query parameters
	if len(queryParams) > 0 {
		values := parsedURL.Query()
		for k, v := range queryParams {
			values.Add(k, v)
		}
		parsedURL.RawQuery = values.Encode()
	}

	return parsedURL.String(), nil
}

// setHeaders sets the request headers
func (c *RestClient) setHeaders(req *http.Request, customHeaders map[string]string) {
	// Set default headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Set custom headers (overrides defaults)
	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}
}

// executeWithRetry executes the request with exponential backoff retry logic
func (c *RestClient) executeWithRetry(ctx context.Context, req *http.Request, retries int, timeout time.Duration) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt < retries; attempt++ {
		// Create a new context with timeout for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, timeout)

		// Clone the request for retry attempts
		clonedReq := req.Clone(attemptCtx)

		// Execute the request
		resp, err := c.httpClient.Do(clonedReq)
		cancel()

		if err != nil {
			lastErr = err

			// Log the retry attempt
			tflog.Warn(ctx, fmt.Sprintf("Request attempt %d failed: %s", attempt+1, err))

			// Don't retry on context cancellation
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			// Calculate backoff delay
			if attempt < retries-1 {
				backoffDelay := c.calculateBackoff(attempt)
				tflog.Debug(ctx, fmt.Sprintf("Retrying in %v", backoffDelay))

				select {
				case <-time.After(backoffDelay):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)

			// Log the error
			tflog.Warn(ctx, fmt.Sprintf("Failed to read response body on attempt %d: %s", attempt+1, err))

			// Don't retry on body read errors for successful HTTP responses
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil, lastErr
			}

			if attempt < retries-1 {
				backoffDelay := c.calculateBackoff(attempt)
				select {
				case <-time.After(backoffDelay):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		// Check if we should retry based on status code
		if c.IsRetryableStatusCode(resp.StatusCode) && attempt < retries-1 {
			lastErr = fmt.Errorf("received retryable status code %d", resp.StatusCode)

			// Log the retry attempt for status code
			tflog.Warn(ctx, fmt.Sprintf("Retryable status code %d on attempt %d", resp.StatusCode, attempt+1))

			backoffDelay := c.calculateBackoff(attempt)
			tflog.Debug(ctx, fmt.Sprintf("Retrying in %v due to status code %d", backoffDelay, resp.StatusCode))

			select {
			case <-time.After(backoffDelay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Create response
		response := &Response{
			StatusCode: resp.StatusCode,
			Body:       body,
			Headers:    resp.Header,
			Request:    req,
		}

		// Log successful request
		tflog.Trace(ctx, "HTTP request completed", map[string]interface{}{
			"method":      req.Method,
			"url":         req.URL.String(),
			"status_code": resp.StatusCode,
			"attempt":     attempt + 1,
		})

		return response, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", retries, lastErr)
}

// calculateBackoff calculates the exponential backoff delay
func (c *RestClient) calculateBackoff(attempt int) time.Duration {
	// Base delay of 1 second with exponential backoff
	baseDelay := time.Second
	maxDelay := 30 * time.Second

	delay := baseDelay * time.Duration(1<<uint(attempt))
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// IsRetryableError determines if an error is retryable
func (c *RestClient) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types that should be retried
	switch {
	case strings.Contains(err.Error(), "connection refused"):
		return true
	case strings.Contains(err.Error(), "timeout"):
		return true
	case strings.Contains(err.Error(), "temporary failure"):
		return true
	case strings.Contains(err.Error(), "network is unreachable"):
		return true
	default:
		return false
	}
}

// IsRetryableStatusCode determines if a status code should be retried
func (c *RestClient) IsRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case 429, // Too Many Requests
		500, // Internal Server Error
		502, // Bad Gateway
		503, // Service Unavailable
		504: // Gateway Timeout
		return true
	default:
		return false
	}
}

// configureTLSAuth configures TLS authentication based on the provided config
func configureTLSAuth(config Config, tlsConfig *tls.Config) error {
	// Handle Certificate Authentication (PEM format)
	if config.ClientCert != "" && config.ClientKey != "" {
		cert, err := tls.X509KeyPair([]byte(config.ClientCert), []byte(config.ClientKey))
		if err != nil {
			return fmt.Errorf("failed to parse client certificate and key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		return nil
	}

	// Handle Certificate Authentication (file-based)
	if config.ClientCertFile != "" && config.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.ClientCertFile, config.ClientKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load client certificate files: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		return nil
	}

	// Handle PKCS12 Authentication (inline base64)
	if config.PKCS12Bundle != "" {
		pkcs12Data, err := base64.StdEncoding.DecodeString(config.PKCS12Bundle)
		if err != nil {
			return fmt.Errorf("failed to decode PKCS12 bundle: %w", err)
		}

		cert, err := parsePKCS12(pkcs12Data, config.PKCS12Password)
		if err != nil {
			return fmt.Errorf("failed to parse PKCS12 bundle: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		return nil
	}

	// Handle PKCS12 Authentication (file-based)
	if config.PKCS12File != "" {
		pkcs12Data, err := ioutil.ReadFile(config.PKCS12File)
		if err != nil {
			return fmt.Errorf("failed to read PKCS12 file: %w", err)
		}

		cert, err := parsePKCS12(pkcs12Data, config.PKCS12Password)
		if err != nil {
			return fmt.Errorf("failed to parse PKCS12 file: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		return nil
	}

	return nil // No certificate authentication configured
}

// parsePKCS12 parses a PKCS12 bundle and returns a TLS certificate
func parsePKCS12(data []byte, password string) (tls.Certificate, error) {
	privateKey, cert, err := pkcs12.Decode(data, password)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode PKCS12: %w", err)
	}

	// Create certificate chain
	certChain := [][]byte{cert.Raw}

	return tls.Certificate{
		Certificate: certChain,
		PrivateKey:  privateKey,
		Leaf:        cert,
	}, nil
}
