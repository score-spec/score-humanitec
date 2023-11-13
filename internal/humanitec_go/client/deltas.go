/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sendgrid/rest"

	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
)

// CreateDelta creates a new Deployment Delta for the orgID and appID.
// The Deployment Delta will be added with the provided content of modules and the 'env_id' and 'name' properties of the 'metadata' property.
func (api *apiClient) CreateDelta(ctx context.Context, orgID, appID string, delta *humanitec.CreateDeploymentDeltaRequest) (*humanitec.DeploymentDelta, error) {
	data, err := json.Marshal(delta)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload into JSON: %w", err)
	}

	apiPath := fmt.Sprintf("/orgs/%s/apps/%s/deltas", orgID, appID)
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
			var res humanitec.DeploymentDelta
			if err = json.Unmarshal([]byte(resp.Body), &res); err != nil {
				return nil, fmt.Errorf("humanitec api: %s %s: parsing response: %w", req.Method, req.BaseURL, err)
			}
			return &res, nil
		}

	default:
		return nil, resError(req, resp)
	}
}

// UpdateDelta updates an existing Deployment Delta for the orgID and appID with the given deltaID.
// The Deltas in the request will be combined and applied on top of existing Delta to produce a merged result.
func (api *apiClient) UpdateDelta(ctx context.Context, orgID string, appID string, deltaID string, deltas []*humanitec.UpdateDeploymentDeltaRequest) (*humanitec.DeploymentDelta, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(deltas)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload into JSON: %w", err)
	}

	apiPath := fmt.Sprintf("/orgs/%s/apps/%s/deltas/%s", orgID, appID, deltaID)
	req := rest.Request{
		Method:  http.MethodPatch,
		BaseURL: api.baseUrl + apiPath,
		Headers: map[string]string{
			"Authorization":        "Bearer " + api.token,
			"Content-Type":         "application/json",
			"Accept":               "application/json",
			"Humanitec-User-Agent": api.humanitecUserAgent,
		},
		Body: buf.Bytes(),
	}

	resp, err := api.client.SendWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("humanitec api: %s %s: %w", req.Method, req.BaseURL, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		{
			var res humanitec.DeploymentDelta
			if err = json.Unmarshal([]byte(resp.Body), &res); err != nil {
				return nil, fmt.Errorf("humanitec api: %s %s: parsing response: %w", req.Method, req.BaseURL, err)
			}
			return &res, nil
		}
	default:
		return nil, resError(req, resp)
	}
}

func resError(req rest.Request, resp *rest.Response) error {
	return fmt.Errorf("humanitec api: %s %s: unexpected response status %d - %s\n%s", req.Method, req.BaseURL, resp.StatusCode, http.StatusText(resp.StatusCode), resp.Body)
}
