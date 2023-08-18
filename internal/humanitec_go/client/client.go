/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package client

import (
	"fmt"
	"net/http"

	"github.com/score-spec/score-humanitec/internal/version"
	"github.com/sendgrid/rest"
)

var (
	ScoreUserAgent = fmt.Sprintf("score-humanitec/%s", version.Version)
)

type apiClient struct {
	baseUrl            string
	token              string
	humanitecUserAgent string

	client *rest.Client
}

// NewClient constructs new Humanitec API client.
func NewClient(url, token string, httpClient *http.Client) (Client, error) {
	return &apiClient{
		baseUrl:            url,
		token:              token,
		humanitecUserAgent: fmt.Sprintf("app %s; sdk %s", ScoreUserAgent, ScoreUserAgent),

		client: &rest.Client{
			HTTPClient: httpClient,
		},
	}, nil
}
