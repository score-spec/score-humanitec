package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
	"github.com/score-spec/score-humanitec/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCreateDelta(t *testing.T) {
	const (
		orgID    = "test_org"
		appID    = "test-app"
		apiToken = "qwe...rty"
	)

	var tests = []struct {
		Name           string
		ApiUrl         string
		Data           *humanitec.CreateDeploymentDeltaRequest
		StatusCode     int
		Response       []byte
		ExpectedResult *humanitec.DeploymentDelta
		ExpectedError  error
	}{
		// Success Path
		//
		{
			Name: "Should return new Deployment Delta",
			Data: &humanitec.CreateDeploymentDeltaRequest{
				Metadata: humanitec.DeltaMetadata{EnvID: "test", Name: "Test draft"},
				Modules: humanitec.ModuleDeltas{
					Add: map[string]map[string]interface{}{
						"module-01": {"image": "busybox"},
					},
				},
			},
			StatusCode: http.StatusOK,
			Response: []byte(`{
				"id": "qwe...rty",
				"metadata": { "env_id": "test", "name": "Test draft" },
				"modules": { 
					"add": { "module-01": { "image": "busybox" } }
				}
			}`),
			ExpectedResult: &humanitec.DeploymentDelta{
				ID:       "qwe...rty",
				Metadata: humanitec.DeltaMetadata{EnvID: "test", Name: "Test draft"},
				Modules: humanitec.ModuleDeltas{
					Add: map[string]map[string]interface{}{
						"module-01": {"image": "busybox"},
					},
				},
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
			StatusCode:    http.StatusInternalServerError,
			ExpectedError: errors.New("HTTP 500"),
		},
		{
			Name:          "Should handle response parsing errors",
			StatusCode:    http.StatusOK,
			Response:      []byte(`{NOT A VALID JSON}`),
			ExpectedError: errors.New("parsing response"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fakeServer := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						switch r.URL.Path {
						case fmt.Sprintf("/orgs/%s/apps/%s/deltas", orgID, appID):
							if r.Method != http.MethodPost {
								w.WriteHeader(http.StatusMethodNotAllowed)
								return
							}
							assert.Equal(t, []string{"Bearer " + apiToken}, r.Header["Authorization"])
							assert.Equal(t, []string{"application/json"}, r.Header["Accept"])
							assert.Equal(t, []string{"application/json"}, r.Header["Content-Type"])

							if tt.Data != nil {
								var body humanitec.CreateDeploymentDeltaRequest
								var dec = json.NewDecoder(r.Body)
								assert.NoError(t, dec.Decode(&body))
								assert.Equal(t, tt.Data, &body)
							}

							w.WriteHeader(tt.StatusCode)
							if len(tt.Response) > 0 {
								w.Header().Set("Content-Type", "application/json")
								w.Write(tt.Response)
							}
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
			res, err := client.CreateDelta(testutil.TestContext(), orgID, appID, tt.Data)

			if tt.ExpectedError != nil {
				// On Error
				assert.ErrorContains(t, err, tt.ExpectedError.Error())
			} else {
				// On Success
				assert.NoError(t, err)
				assert.Equal(t, tt.ExpectedResult, res)
			}
		})
	}
}
