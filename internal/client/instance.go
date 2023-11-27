package client

import (
	"github.com/crossplane/crossplane-runtime/pkg/logging"
)

func NewCodeFreshService(creds []byte, logger logging.Logger) (interface{}, error) {
	logger.Info("Received raw credentials", "creds", string(creds))
	apiKey := string(creds)

	baseURL := "https://g.codefresh.io/api"
	return NewCodeFreshAPIClient(apiKey, baseURL, logger), nil
}
