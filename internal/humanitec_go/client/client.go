package client

import (
	"net/http"

	"github.com/sendgrid/rest"
)

type apiClient struct {
	baseUrl string
	token   string

	client *rest.Client
}

// NewClient constructs new Humanitec API client.
func NewClient(url, token string, httpClient *http.Client) (Client, error) {
	return &apiClient{
		baseUrl: url,
		token:   token,

		client: &rest.Client{
			HTTPClient: httpClient,
		},
	}, nil
}
