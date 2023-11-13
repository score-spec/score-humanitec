/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
	"github.com/score-spec/score-humanitec/internal/testutil"
)

func TestStartDeployment(t *testing.T) {
	const (
		orgID    = "test_org"
		appID    = "test-app"
		envID    = "test-env"
		apiToken = "qwe...rty"
	)

	var tests = []struct {
		Name           string
		ApiUrl         string
		Data           *humanitec.StartDeploymentRequest
		Retry          bool
		StatusCode     []int
		Response       []byte
		ExpectedResult *humanitec.Deployment
		ExpectedError  error
	}{
		// Success Path
		//
		{
			Name: "Should return new Deployment",
			Data: &humanitec.StartDeploymentRequest{
				DeltaID: "test-delta",
				Comment: "Test deployment",
			},
			StatusCode: []int{http.StatusCreated},
			Response: []byte(`{
				"id": "qwe...rty",
				"env_id": "test-env",
				"from_id": "qwe...rty",
				"delta_id": "test-delta",
				"comment": "Test deployment",
				"status": "in progress",
				"status_changed_at": "2020-05-22T14:59:01Z",
				"created_at": "2020-05-22T14:58:07Z",
				"created_by": "a.user@example.com"
			}`),
			ExpectedResult: &humanitec.Deployment{
				ID:              "qwe...rty",
				EnvID:           "test-env",
				FromID:          "qwe...rty",
				DeltaID:         "test-delta",
				Comment:         "Test deployment",
				Status:          "in progress",
				StatusChangedAt: time.Time{},
				CreatedBy:       "a.user@example.com",
				CreatedAt:       time.Time{},
			},
		},

		// Errors Handling
		//
		{
			Name:          "Should handle request errors",
			ApiUrl:        "bad URL",
			ExpectedError: errors.New("unsupported protocol scheme"),
		},
		{
			Name:          "Should handle API errors",
			StatusCode:    []int{http.StatusInternalServerError},
			Response:      []byte(`error details`),
			ExpectedError: errors.New("unexpected response status 500 - Internal Server Error\nerror details"),
		},
		{
			Name:          "Should handle response parsing errors",
			StatusCode:    []int{http.StatusOK},
			Response:      []byte(`{NOT A VALID JSON}`),
			ExpectedError: errors.New("parsing response"),
		},
		{
			Name: "Should return conflict errors without retry",
			Data: &humanitec.StartDeploymentRequest{
				DeltaID: "test-delta",
				Comment: "Test deployment",
			},
			Retry:         false,
			StatusCode:    []int{http.StatusConflict},
			Response:      []byte(`conflict details`),
			ExpectedError: errors.New("unexpected response status 409 - Conflict\nconflict details"),
		},
		{
			Name: "Should retry conflict errors with retry",
			Data: &humanitec.StartDeploymentRequest{
				DeltaID: "test-delta",
				Comment: "Test deployment",
			},
			Retry:      true,
			StatusCode: []int{http.StatusConflict, http.StatusCreated},
			Response: []byte(`{
				"id": "qwe...rty",
				"env_id": "test-env",
				"from_id": "qwe...rty",
				"delta_id": "test-delta",
				"comment": "Test deployment",
				"status": "in progress",
				"status_changed_at": "2020-05-22T14:59:01Z",
				"created_at": "2020-05-22T14:58:07Z",
				"created_by": "a.user@example.com"
			}`),
			ExpectedResult: &humanitec.Deployment{
				ID:              "qwe...rty",
				EnvID:           "test-env",
				FromID:          "qwe...rty",
				DeltaID:         "test-delta",
				Comment:         "Test deployment",
				Status:          "in progress",
				StatusChangedAt: time.Time{},
				CreatedBy:       "a.user@example.com",
				CreatedAt:       time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			request := 0
			fakeServer := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						switch r.URL.Path {
						case fmt.Sprintf("/orgs/%s/apps/%s/envs/%s/deploys", orgID, appID, envID):
							if r.Method != http.MethodPost {
								w.WriteHeader(http.StatusMethodNotAllowed)
								return
							}
							assert.Equal(t, []string{"Bearer " + apiToken}, r.Header["Authorization"])
							assert.Equal(t, []string{"application/json"}, r.Header["Accept"])
							assert.Equal(t, []string{"application/json"}, r.Header["Content-Type"])
							assert.Equal(t, []string{"app score-humanitec/0.0.0; sdk score-humanitec/0.0.0"}, r.Header["Humanitec-User-Agent"])

							if tt.Data != nil {
								var body humanitec.StartDeploymentRequest
								var dec = json.NewDecoder(r.Body)
								assert.NoError(t, dec.Decode(&body))
								assert.Equal(t, tt.Data, &body)
							}

							w.WriteHeader(tt.StatusCode[request])
							if len(tt.Response) > 0 {
								w.Header().Set("Content-Type", "application/json")
								w.Write(tt.Response)
							}

							request++
							return
						}

						w.WriteHeader(http.StatusNotFound)
					},
				),
			)
			defer fakeServer.Close()

			if tt.ApiUrl == "" {
				tt.ApiUrl = fakeServer.URL
			}

			client, err := NewClient(tt.ApiUrl, apiToken, fakeServer.Client())
			assert.NoError(t, err)
			res, err := client.StartDeployment(testutil.TestContext(), orgID, appID, envID, tt.Retry, tt.Data)

			if tt.ExpectedError != nil {
				// On Error
				assert.ErrorContains(t, err, tt.ExpectedError.Error())
			} else {
				// On Success
				assert.NoError(t, err)
				tt.ExpectedResult.StatusChangedAt = res.StatusChangedAt
				tt.ExpectedResult.CreatedAt = res.CreatedAt
				assert.Equal(t, tt.ExpectedResult, res)
			}
		})
	}
}
