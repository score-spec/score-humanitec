/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sendgrid/rest"

	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
)

var (
	retryDelay = 5 * time.Second
)

// StartDeployment starts a new Deployment.
func (api *apiClient) StartDeployment(ctx context.Context, orgID, appID, envID string, retry bool, deployment *humanitec.StartDeploymentRequest) (*humanitec.Deployment, error) {
	data, err := json.Marshal(deployment)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload into JSON: %w", err)
	}

	apiPath := fmt.Sprintf("/orgs/%s/apps/%s/envs/%s/deploys", orgID, appID, envID)
	req := rest.Request{
		Method:  http.MethodPost,
		BaseURL: api.baseUrl + apiPath,
		Headers: map[string]string{
			"Authorization":        "Bearer " + api.token,
			"Content-Type":         "application/json",
			"Accept":               "application/json",
			"Humanitec-User-Agent": api.humanitecUserAgent,
		},
		Body: data,
	}

	resp, err := api.client.SendWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("humanitec api: %s %s: %w", req.Method, req.BaseURL, err)
	}

	switch resp.StatusCode {

	case http.StatusOK, http.StatusCreated:
		{
			var res humanitec.Deployment
			if err = json.Unmarshal([]byte(resp.Body), &res); err != nil {
				return nil, fmt.Errorf("humanitec api: %s %s: parsing response: %w", req.Method, req.BaseURL, err)
			}
			return &res, nil
		}

	case http.StatusConflict:
		if retry {
			time.Sleep(retryDelay)
			return api.StartDeployment(ctx, orgID, appID, envID, retry, deployment)
		}

		fallthrough
	default:
		return nil, fmt.Errorf("humanitec api: %s %s: HTTP %d - %s", req.Method, req.BaseURL, resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}
