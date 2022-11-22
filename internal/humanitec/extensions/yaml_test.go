/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package extensions

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestYamlDecode(t *testing.T) {
	var tests = []struct {
		Name   string
		Source io.Reader
		Output HumanitecExtensionsSpec
		Error  error
	}{
		{
			Name:   "Should handle empty input",
			Source: bytes.NewReader([]byte{}),
			Error:  errors.New("EOF"),
		},
		{
			Name:   "Should handle invalid YAML input",
			Source: bytes.NewReader([]byte("<NOT A VALID YAML>")),
			Error:  errors.New("cannot unmarshal"),
		},
		{
			Name: "Should decode the Humanitec extensions spec",
			Source: bytes.NewReader([]byte(`
---
apiVersion: humanitec.org/v1b1

profile: "test-org/test-module"
spec:
  "labels":
    "tags.datadoghq.com/env": "${resources.env.DATADOG_ENV}"
  "ingress":
    rules:
      "${resources.dns}":
        http:
          "/":
            type: prefix
            port: 80

resources:
  db:
    scope: external
  dns:
    scope: shared
`)),
			Output: HumanitecExtensionsSpec{
				ApiVersion: "humanitec.org/v1b1",
				Profile:    "test-org/test-module",
				Spec: map[string]interface{}{
					"labels": map[string]interface{}{
						"tags.datadoghq.com/env": "${resources.env.DATADOG_ENV}",
					},
					"ingress": map[string]interface{}{
						"rules": map[string]interface{}{
							"${resources.dns}": map[string]interface{}{
								"http": map[string]interface{}{
									"/": map[string]interface{}{
										"type": "prefix",
										"port": 80,
									},
								},
							},
						},
					},
				},
				Resources: HumanitecResourcesSpecs{
					"db":  HumanitecResourceSpec{Scope: "external"},
					"dns": HumanitecResourceSpec{Scope: "shared"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var src map[string]interface{}
			var err = yaml.NewDecoder(tt.Source).Decode(&src)

			if tt.Error != nil {
				// On Error
				//
				assert.ErrorContains(t, err, tt.Error.Error())
			} else {
				// On Success
				//
				assert.NoError(t, err)

				var ext HumanitecExtensionsSpec
				assert.NoError(t, mapstructure.Decode(src, &ext))

				assert.Equal(t, tt.Output, ext)
			}
		})
	}
}
