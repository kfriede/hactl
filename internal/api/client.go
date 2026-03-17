package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	retryBaseDelay = 1 * time.Second
	userAgent      = "hactl/0.1.0"
)

// Client is a generic HTTP API client with retry logic, verbose/debug logging.
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	verbose    bool
	debug      bool
	errWriter  io.Writer
}

// ClientConfig configures a Client.
type ClientConfig struct {
	Token     string        // Bearer token for authentication
	BaseURL   string        // Home Assistant base URL
	Timeout   time.Duration // HTTP request timeout
	Verbose   bool          // Log request method/URL and status codes to stderr
	Debug     bool          // Log full request/response bodies to stderr
	ErrWriter io.Writer     // Writer for verbose/debug output (typically os.Stderr)
}

// NewClient creates a new API client.
func NewClient(cfg ClientConfig) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL:   baseURL,
		token:     cfg.Token,
		verbose:   cfg.Verbose,
		debug:     cfg.Debug,
		errWriter: cfg.ErrWriter,
	}
}

// APIError represents an error returned by the Home Assistant API.
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s (%d): %s", e.Code, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

// Do performs an HTTP request with retry logic.
func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	var lastErr error
	for attempt := range maxRetries {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		// Home Assistant uses Bearer token authentication
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}

		if c.verbose {
			_, _ = fmt.Fprintf(c.errWriter, "[%s] %s\n", method, url)
		}

		start := time.Now()
		resp, err := c.httpClient.Do(req)
		elapsed := time.Since(start)

		if c.verbose {
			if err != nil {
				_, _ = fmt.Fprintf(c.errWriter, "  error: %v (%.1fs)\n", err, elapsed.Seconds())
			} else {
				_, _ = fmt.Fprintf(c.errWriter, "  %d (%.1fs)\n", resp.StatusCode, elapsed.Seconds())
			}
		}

		if err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay(attempt))
			}
			continue
		}

		// Retry on 429 and 5xx
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d from %s %s", resp.StatusCode, method, path)
			if attempt < maxRetries-1 {
				delay := retryDelay(attempt)
				if c.verbose {
					_, _ = fmt.Fprintf(c.errWriter, "  retrying in %v...\n", delay)
				}
				time.Sleep(delay)
			}
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// Get performs a GET request and returns the response body.
func (c *Client) Get(path string) ([]byte, error) {
	resp, err := c.Do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if c.debug {
		_, _ = fmt.Fprintf(c.errWriter, "  response body: %s\n", truncate(string(data), 2000))
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, data)
	}

	return data, nil
}

// GetJSON performs a GET request and unmarshals the response.
func (c *Client) GetJSON(path string, target any) error {
	data, err := c.Get(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// GetRaw performs a GET request and returns the response as a string.
// Useful for endpoints that return plaintext (e.g., /api/error_log).
func (c *Client) GetRaw(path string) (string, error) {
	data, err := c.Get(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetToFile performs a GET request and writes the response body to a file.
// Useful for binary endpoints (e.g., camera snapshots).
func (c *Client) GetToFile(path, filePath string) error {
	resp, err := c.Do("GET", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return parseAPIError(resp.StatusCode, data)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body any) ([]byte, error) {
	return c.mutate("POST", path, body)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(path string, body any) ([]byte, error) {
	return c.mutate("PUT", path, body)
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(path string, body any) ([]byte, error) {
	return c.mutate("PATCH", path, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	resp, err := c.Do("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return parseAPIError(resp.StatusCode, data)
	}

	return nil
}

func (c *Client) mutate(method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		if c.debug {
			_, _ = fmt.Fprintf(c.errWriter, "  request body: %s\n", truncate(string(data), 2000))
		}
		bodyReader = strings.NewReader(string(data))
	}

	resp, err := c.Do(method, path, bodyReader)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if c.debug {
		_, _ = fmt.Fprintf(c.errWriter, "  response body: %s\n", truncate(string(respData), 2000))
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, respData)
	}

	return respData, nil
}

func parseAPIError(statusCode int, data []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}
	if err := json.Unmarshal(data, apiErr); err != nil {
		apiErr.Message = string(data)
	}
	if apiErr.Code == "" {
		apiErr.Code = http.StatusText(statusCode)
	}
	return apiErr
}

func retryDelay(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * retryBaseDelay
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
