/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package humanitec

import (
	"testing"

	score "github.com/score-spec/score-go/types"
	"github.com/score-spec/score-humanitec/internal/humanitec/extensions"
	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
	"github.com/stretchr/testify/assert"
)

func TestScoreConvert(t *testing.T) {
	const (
		envID = "test"
		name  = "Test delta"
	)

	var tests = []struct {
		Name       string
		Source     *score.WorkloadSpec
		Extensions *extensions.HumanitecExtensionsSpec
		Output     *humanitec.CreateDeploymentDeltaRequest
		Error      error
	}{
		{
			Name: "Should convert SCORE to deployment delta",
			Source: &score.WorkloadSpec{
				Metadata: score.WorkloadMeta{
					Name: "backend",
				},
				Service: score.ServiceSpec{
					Ports: score.ServicePortsSpecs{
						"www": score.ServicePortSpec{
							Port:       80,
							TargetPort: 8080,
						},
						"admin": score.ServicePortSpec{
							Port: 8080,
						},
					},
				},
				Containers: score.ContainersSpecs{
					"backend": score.ContainerSpec{
						Image: "busybox",
						Command: []string{
							"/bin/sh",
						},
						Args: []string{
							"-c",
							"while true; do printenv; echo ...sleeping 10 sec...; sleep 10; done",
						},
						Variables: map[string]string{
							"CONNECTION_STRING": "test connection string",
						},
						Resources: score.ContainerResourcesRequirementsSpec{
							Limits: map[string]interface{}{
								"memory": "128Mi",
								"cpu":    "500m",
							},
							Requests: map[string]interface{}{
								"memory": "64Mi",
								"cpu":    "250m",
							},
						},
						LivenessProbe: score.ContainerProbeSpec{
							HTTPGet: score.HTTPGetActionSpec{
								Path: "/alive",
								Port: 8080,
							},
						},
						ReadinessProbe: score.ContainerProbeSpec{
							HTTPGet: score.HTTPGetActionSpec{
								Path: "/health",
								Port: 8080,
								HTTPHeaders: []score.HTTPHeaderSpec{
									{Name: "Custom-Header", Value: "Ops!"},
								},
							},
						},
					},
				},
			},
			Extensions: &extensions.HumanitecExtensionsSpec{},
			Output: &humanitec.CreateDeploymentDeltaRequest{
				Metadata: humanitec.DeltaMetadata{EnvID: envID, Name: name},
				Modules: humanitec.ModuleDeltas{
					Add: map[string]map[string]interface{}{
						"backend": {
							"profile": "humanitec/default-module",
							"spec": map[string]interface{}{
								"containers": map[string]interface{}{
									"backend": map[string]interface{}{
										"id":    "backend",
										"image": "busybox",
										"command": []string{
											"/bin/sh",
										},
										"args": []string{
											"-c",
											"while true; do printenv; echo ...sleeping 10 sec...; sleep 10; done",
										},
										"variables": map[string]interface{}{
											"CONNECTION_STRING": "test connection string",
										},
										"resources": map[string]interface{}{
											"limits": map[string]interface{}{
												"memory": "128Mi",
												"cpu":    "500m",
											},
											"requests": map[string]interface{}{
												"memory": "64Mi",
												"cpu":    "250m",
											},
										},
										"liveness_probe": map[string]interface{}{
											"type": "http",
											"path": "/alive",
											"port": 8080,
										},
										"readiness_probe": map[string]interface{}{
											"type": "http",
											"path": "/health",
											"port": 8080,
											"headers": map[string]string{
												"Custom-Header": "Ops!",
											},
										},
									},
								},
								"service": map[string]interface{}{
									"ports": map[string]interface{}{
										"www": map[string]interface{}{
											"protocol":       "TCP",
											"service_port":   80,
											"container_port": 8080,
										},
										"admin": map[string]interface{}{
											"protocol":       "TCP",
											"service_port":   8080,
											"container_port": 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "Should convert all resources references",
			Source: &score.WorkloadSpec{
				Metadata: score.WorkloadMeta{
					Name: "test",
				},
				Containers: score.ContainersSpecs{
					"backend": score.ContainerSpec{
						Variables: map[string]string{
							"DEBUG":             "${resources.env.DEBUG}",
							"LOGS_LEVEL":        "${pod.debug.level}",
							"ORDERS_SERVICE":    "http://${resources.orders.service.name}:${resources.orders.service.port}/api",
							"CONNECTION_STRING": "postgresql://${resources.db.host}:${resources.db.port}/${resources.db.name}",
							"DOMAIN_NAME":       "${resources.dns.domain}",
						},
						Files: []score.FileMountSpec{
							{
								Target: "/etc/backend/config.yaml",
								Mode:   "666",
								Content: []string{
									"---",
									"DEBUG: ${resources.env.DEBUG}",
								},
							},
						},
						Volumes: []score.VolumeMountSpec{
							{
								Source:   "${resources.data}",
								Path:     "sub/path",
								Target:   "/mnt/data",
								ReadOnly: true,
							},
						},
					},
				},
				Resources: map[string]score.ResourceSpec{
					"env": {
						Type: "environment",
						Properties: map[string]score.ResourcePropertySpec{
							"DEBUG":       {Default: false, Required: false},
							"DATADOG_ENV": {},
						},
					},
					"dns": {
						Type: "dns",
						Properties: map[string]score.ResourcePropertySpec{
							"domain": {},
						},
					},
					"data": {
						Type: "volume",
					},
					"db": {
						Type: "postgres",
						Properties: map[string]score.ResourcePropertySpec{
							"host":      {Default: "localhost", Required: true},
							"port":      {Default: 5432, Required: false},
							"name":      {Required: true},
							"user_name": {Required: true, Secret: true},
							"password":  {Required: true, Secret: true},
						},
					},
					"orders": {
						Type: "workload",
						Properties: map[string]score.ResourcePropertySpec{
							"service.name": {Required: false},
							"service.port": {},
						},
					},
				},
			},
			Extensions: &extensions.HumanitecExtensionsSpec{
				Profile: "test-org/test-module",
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
				Resources: extensions.HumanitecResourcesSpecs{
					"dns": extensions.HumanitecResourceSpec{
						Scope: "shared",
					},
				},
			},
			Output: &humanitec.CreateDeploymentDeltaRequest{
				Metadata: humanitec.DeltaMetadata{EnvID: envID, Name: name},
				Modules: humanitec.ModuleDeltas{
					Add: map[string]map[string]interface{}{
						"test": {
							"profile": "test-org/test-module",
							"spec": map[string]interface{}{
								"containers": map[string]interface{}{
									"backend": map[string]interface{}{
										"id": "backend",
										"variables": map[string]interface{}{
											"DEBUG":             "${values.DEBUG}",
											"LOGS_LEVEL":        "${pod.debug.level}",
											"ORDERS_SERVICE":    "http://${modules.orders.service.name}:${modules.orders.service.port}/api",
											"CONNECTION_STRING": "postgresql://${externals.db.host}:${externals.db.port}/${externals.db.name}",
											"DOMAIN_NAME":       "${shared.dns.domain}",
										},
										"files": map[string]interface{}{
											"/etc/backend/config.yaml": map[string]interface{}{
												"mode":  "666",
												"value": "---\nDEBUG: ${values.DEBUG}",
											},
										},
										"volume_mounts": map[string]interface{}{
											"/mnt/data": map[string]interface{}{
												"id":        "externals.data",
												"sub_path":  "sub/path",
												"read_only": true,
											},
										},
									},
								},
								"ingress": map[string]interface{}{
									"rules": map[string]interface{}{
										"shared.dns": map[string]interface{}{
											"http": map[string]interface{}{
												"/": map[string]interface{}{
													"type": "prefix",
													"port": 80,
												},
											},
										},
									},
								},
								"labels": map[string]interface{}{
									"tags.datadoghq.com/env": "${values.DATADOG_ENV}",
								},
							},
							"externals": map[string]interface{}{
								"data": map[string]interface{}{
									"type": "volume",
								},
								"db": map[string]interface{}{
									"type": "postgres",
								},
							},
						},
					},
				},
				Shared: []humanitec.UpdateAction{
					{
						Operation: "add",
						Path:      "/dns",
						Value: map[string]interface{}{
							"type": "dns",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			res, err := ConvertSpec(name, envID, tt.Source, tt.Extensions)

			if tt.Error != nil {
				// On Error
				//
				assert.ErrorContains(t, err, tt.Error.Error())
			} else {
				// On Success
				//
				assert.NoError(t, err)
				assert.Equal(t, tt.Output, res)
			}
		})
	}
}
