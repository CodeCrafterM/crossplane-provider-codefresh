package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/pkg/errors"
)

const (
	errorParsingRequest     = "error parsing request URL"
	errorMarshaling         = "error marshalling request body"
	errorCreatingRequest    = "error creating request"
	errorSendingRequest     = "error sending request to CodeFresh"
	errorClosingBodyRequest = "Failed to close response body"
)

var ErrResourceNotFound = errors.New("resource not found in CodeFresh")

type CodeFreshAPI interface {
	CheckResourceExists(ctx context.Context, resourceType, id string) (bool, error)
	GetResource(ctx context.Context, resourceType, id string, response interface{}) error
	CreateResource(ctx context.Context, resourceType string, params, response interface{}) error
	UpdateResource(ctx context.Context, resourceType, id string, params, response interface{}) error
	DeleteResource(ctx context.Context, resourceType, id string) error
	// Add other methods as needed
}

type CodeFreshAPIClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	logger     logging.Logger
}

func NewCodeFreshAPIClient(apiKey, baseURL string, logger logging.Logger) *CodeFreshAPIClient {
	return &CodeFreshAPIClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
		apiKey:     apiKey,
		logger:     logger,
	}
}

// sendRequest sends an HTTP request to the CodeFresh API and handles the response.
func (c *CodeFreshAPIClient) sendRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Build the request URL
	requestURL, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, errors.Wrap(err, errorParsingRequest)
	}

	var requestBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, errorMarshaling)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), requestBody)
	if err != nil {
		return nil, errors.Wrap(err, errorCreatingRequest)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errorSendingRequest)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.logger.Debug(errorClosingBodyRequest)
			}
		}()
		respBody, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusNotFound, http.StatusInternalServerError:
			// Specific handling for 404 and 500 errors
			return nil, errors.Errorf("CodeFresh API returned error: %s - %s", resp.Status, string(respBody))
		default:
			// General error handling
			return nil, errors.Errorf("unexpected response from CodeFresh: %s - %s", resp.Status, string(respBody))
		}
	}

	return resp, nil
}

// Generic CRUD operations for different resources in CodeFresh.

// CheckResourceExists checks if a resource exists in CodeFresh.
func (c *CodeFreshAPIClient) CheckResourceExists(ctx context.Context, resourceType, id string) (bool, error) {
	var response struct {
		ID string `json:"id"`
	}

	resp, err := c.sendRequest(ctx, "GET", "/"+resourceType+"/"+id, nil)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "500") {
			return false, nil
		}
		return false, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Debug(errorClosingBodyRequest)
		}
	}()

	return response.ID != "", nil
}

// GetResource fetches a resource from CodeFresh.
func (c *CodeFreshAPIClient) GetResource(ctx context.Context, resourceType, id string, response interface{}) error {
	resp, err := c.sendRequest(ctx, "GET", "/"+resourceType+"/"+id, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Debug(errorClosingBodyRequest)
		}
	}()

	return json.NewDecoder(resp.Body).Decode(response)
}

// CreateResource creates a new resource in CodeFresh.
func (c *CodeFreshAPIClient) CreateResource(ctx context.Context, resourceType string, params, response interface{}) error {
	resp, err := c.sendRequest(ctx, "POST", "/"+resourceType, params)
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Debug(errorClosingBodyRequest)
		}
	}()

	return json.NewDecoder(resp.Body).Decode(response)
}

// UpdateResource updates an existing resource in CodeFresh.
func (c *CodeFreshAPIClient) UpdateResource(ctx context.Context, resourceType, id string, params, response interface{}) error {
	resp, err := c.sendRequest(ctx, "PATCH", "/"+resourceType+"/"+id, params)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Debug("Error closing response body", "error", err)
		}
	}()

	// Only attempt to unmarshal if a response struct is provided
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}

	return nil
}

// DeleteResource deletes a resource from CodeFresh.
func (c *CodeFreshAPIClient) DeleteResource(ctx context.Context, resourceType, id string) error {
	resp, err := c.sendRequest(ctx, "DELETE", "/"+resourceType+"/"+id, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Debug(errorClosingBodyRequest)
		}
	}()

	// Check the response status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("unexpected response from CodeFresh on delete: %s", resp.Status)
	}

	return nil
}
