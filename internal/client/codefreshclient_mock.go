package client

import (
	"context"
	"crossplane-provider-codefresh/apis/resource/v1alpha1"
)

type MockCodeFreshAPIClient struct {
	MockCheckResourceExistsResponse bool
	MockCheckResourceExistsErr      error

	MockGetResourceResponse *v1alpha1.ProjectDetails
	MockGetResourceErr      error

	MockCreateResourceResponse interface{}
	MockCreateResourceErr      error

	MockUpdateResourceResponse interface{}
	MockUpdateResourceErr      error

	MockDeleteResourceErr error
}

var _ CodeFreshAPI = &MockCodeFreshAPIClient{}

// CheckResourceExists simulates checking if a resource exists in CodeFresh.
func (m *MockCodeFreshAPIClient) CheckResourceExists(ctx context.Context, resourceType, id string) (bool, error) {
	return m.MockCheckResourceExistsResponse, m.MockCheckResourceExistsErr
}

// GetResource simulates fetching a resource from CodeFresh.
func (m *MockCodeFreshAPIClient) GetResource(ctx context.Context, resourceType, id string, response interface{}) error {
	// Check if a mock error is set
	if m.MockGetResourceErr != nil {
		return m.MockGetResourceErr
	}

	// Cast the response to the expected type
	switch resourceType {
	case "projects":
		*response.(*v1alpha1.ProjectDetails) = *m.MockGetResourceResponse
	default:
		return nil
	}

	return nil
}

// CreateResource simulates creating a resource in CodeFresh.
func (m *MockCodeFreshAPIClient) CreateResource(ctx context.Context, resourceType string, params, response interface{}) error {
	// Simulate populating the response
	return m.MockCreateResourceErr
}

// UpdateResource simulates updating a resource in CodeFresh.
func (m *MockCodeFreshAPIClient) UpdateResource(ctx context.Context, resourceType, id string, params, response interface{}) error {
	// Simulate populating the response
	return m.MockUpdateResourceErr
}

// DeleteResource simulates deleting a resource from CodeFresh.
func (m *MockCodeFreshAPIClient) DeleteResource(ctx context.Context, resourceType, id string) error {
	return m.MockDeleteResourceErr
}
