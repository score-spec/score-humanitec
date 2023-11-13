/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
	"github.com/score-spec/score-humanitec/internal/testutil"
)

func TestListResourceTypes(t *testing.T) {
	const (
		orgID    = "test_org"
		apiToken = "qwe...rty"
	)

	var tests = []struct {
		Name           string
		ApiUrl         string
		StatusCode     int
		Response       []byte
		ExpectedResult []humanitec.ResourceType
		ExpectedError  error
	}{
		// Success Path
		//
		{
			Name:           "Should return an empty list",
			StatusCode:     http.StatusOK,
			Response:       []byte(`[]`),
			ExpectedResult: make([]humanitec.ResourceType, 0),
		},
		{
			Name:       "Should return list of resource types",
			StatusCode: http.StatusOK,
			Response: []byte(`[{
				"type": "postgres",
				"name": "PostgreSQL",
				"category": "storage",
				"inputs_schema": { "properties": {}, "type": "object" },
				"outputs_schema": { "values": {}, "secrets": {} }
			}, {
				"type": "dns",
				"name": "DNS Record",
				"inputs_schema": {}
			}]`),
			ExpectedResult: []humanitec.ResourceType{
				{
					Type:     "postgres",
					Name:     "PostgreSQL",
					Category: "storage",
					InputsSchema: map[string]interface{}{
						"properties": map[string]interface{}{},
						"type":       "object",
					},
					OutputsSchema: map[string]interface{}{
						"values":  map[string]interface{}{},
						"secrets": map[string]interface{}{},
					},
				}, {
					Type:          "dns",
					Name:          "DNS Record",
					InputsSchema:  map[string]interface{}{},
					OutputsSchema: nil,
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
			ExpectedError: errors.New("unexpected response status 500 - Internal Server Error\nerror details"),
			Response:      []byte(`error details`),
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
						case fmt.Sprintf("/orgs/%s/resources/types", orgID):
							if r.Method != http.MethodGet {
								w.WriteHeader(http.StatusMethodNotAllowed)
								return
							}
							assert.Equal(t, []string{"Bearer " + apiToken}, r.Header["Authorization"])
							assert.Equal(t, []string{"application/json"}, r.Header["Accept"])
							assert.Equal(t, []string{"app score-humanitec/0.0.0; sdk score-humanitec/0.0.0"}, r.Header["Humanitec-User-Agent"])

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
			res, err := client.ListResourceTypes(testutil.TestContext(), orgID)

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
