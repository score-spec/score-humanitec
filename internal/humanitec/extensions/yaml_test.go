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

service:
  routes:
    http:
      "/":
        type: prefix
        from: ${resources.dns}
        port: 80

resources:
    db:
        scope: external
    dns:
        scope: shared
`)),
			Output: HumanitecExtensionsSpec{
				ApiVersion: "humanitec.org/v1b1",
				Service: HumanitecServiceSpec{
					Routes: HumanitecServiceRoutesSpecs{
						"http": HumanitecServiceRoutePathsSpec{
							"/": HumanitecServiceRoutePathSpec{
								From: "${resources.dns}",
								Type: "prefix",
								Port: 80,
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
