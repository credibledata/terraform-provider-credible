package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the HTTP client for the Credible Admin API.
type Client struct {
	BaseURL      string
	AuthHeader   string // Full Authorization header value (e.g. "ApiKey xxx" or "Bearer xxx")
	Organization string
	HTTPClient   *http.Client
}

// NewClient creates a new Credible API client.
// authHeader should be the full Authorization header value.
func NewClient(baseURL, authHeader, organization string) *Client {
	return &Client{
		BaseURL:      strings.TrimRight(baseURL, "/"),
		AuthHeader:   authHeader,
		Organization: organization,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// doRequest performs an HTTP request and returns the response body.
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, int, error) {
	url := fmt.Sprintf("%s/api/v0%s", c.BaseURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", c.AuthHeader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// doJSON performs an HTTP request with JSON body, checks for errors, and unmarshals the response.
func (c *Client) doJSON(method, path string, reqBody interface{}, result interface{}) error {
	respBody, statusCode, err := c.doRequest(method, path, reqBody)
	if err != nil {
		return err
	}

	if statusCode < 200 || statusCode >= 300 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("API error (HTTP %d): %s", statusCode, apiErr.Message)
		}
		return fmt.Errorf("API error (HTTP %d): %s", statusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

// doRequestRaw performs an HTTP request with a pre-built *http.Request.
func (c *Client) doRequestRaw(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Authorization", c.AuthHeader)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// IsNotFound checks if status code indicates a 404.
func IsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "HTTP 404")
}

// APIError represents an error response from the API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
