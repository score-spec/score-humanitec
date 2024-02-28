/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package humanitec

import (
	"errors"
	"testing"

	score "github.com/score-spec/score-go/types"
	"github.com/stretchr/testify/assert"

	"github.com/score-spec/score-humanitec/internal/humanitec/extensions"
	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
)

func TestParseResourceId(t *testing.T) {
	var tests = []struct {
		Name               string
		ResourceReference  string
		ExpectedModuleId   string
		ExpectedScope      string
		ExpectedResourceId string
		ExpectedError      error
	}{
		// Success path
		//
		{
			Name:               "Should accept empty string",
			ResourceReference:  "",
			ExpectedResourceId: "",
			ExpectedError:      nil,
		},
		{
			Name:               "Should accept resource ID only",
			ResourceReference:  "test-res-id",
			ExpectedResourceId: "test-res-id",
			ExpectedError:      nil,
		},
		{
			Name:               "Should accept external resource reference",
			ResourceReference:  "externals.test-res-id",
			ExpectedScope:      "externals",
			ExpectedResourceId: "test-res-id",
			ExpectedError:      nil,
		},
		{
			Name:               "Should accept shared resource reference",
			ResourceReference:  "shared.test-res-id",
			ExpectedScope:      "shared",
			ExpectedResourceId: "test-res-id",
			ExpectedError:      nil,
		},
		{
			Name:               "Should accept foreighn module resource reference",
			ResourceReference:  "modules.test-module.externals.test-res-id",
			ExpectedModuleId:   "test-module",
			ExpectedScope:      "externals",
			ExpectedResourceId: "test-res-id",
			ExpectedError:      nil,
		},

		// Errors handling
		//
		{
			Name:              "Should reject incomplete resource reference",
			ResourceReference: "test-module.externals.test-res-id",
			ExpectedError:     errors.New("not supported"),
		},
		{
			Name:              "Should reject non-module resource reference",
			ResourceReference: "something.test-something.externals.test-res-id",
			ExpectedError:     errors.New("not supported"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			mod, scope, resId, err := parseResourceId(tt.ResourceReference)

			if tt.ExpectedError != nil {
				// On Error
				//
				assert.ErrorContains(t, err, tt.ExpectedError.Error())
			} else {
				// On Success
				//
				assert.NoError(t, err)
				assert.Equal(t, tt.ExpectedModuleId, mod)
				assert.Equal(t, tt.ExpectedScope, scope)
				assert.Equal(t, tt.ExpectedResourceId, resId)
			}
		})
	}
}

func TestScoreConvert(t *testing.T) {
	const (
		envID = "test"
		name  = "Test delta"
	)

	var tests = []struct {
		Name              string
		Source            *score.Workload
		Extensions        *extensions.HumanitecExtensionsSpec
		Output            *humanitec.CreateDeploymentDeltaRequest
		WorkloadSourceURL string
		Error             error
	}{
		{
			Name: "Should convert SCORE to deployment delta",
			Source: &score.Workload{
				Metadata: score.WorkloadMetadata{
					"name": "backend",
				},
				Service: &score.WorkloadService{
					Ports: score.WorkloadServicePorts{
						"www": score.ServicePort{
							Port:       80,
							TargetPort: Ref(8080),
						},
						"admin": score.ServicePort{
							Port: 8080,
						},
					},
				},
				Containers: score.WorkloadContainers{
					"backend": score.Container{
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
						Resources: &score.ContainerResources{
							Limits: &score.ResourcesLimits{
								Memory: Ref("128Mi"),
								Cpu:    Ref("500m"),
							},
							Requests: &score.ResourcesLimits{
								Memory: Ref("64Mi"),
								Cpu:    Ref("250m"),
							},
						},
						LivenessProbe: &score.ContainerProbe{
							HttpGet: score.HttpProbe{
								Path: "/alive",
								Port: 8080,
							},
						},
						ReadinessProbe: &score.ContainerProbe{
							HttpGet: score.HttpProbe{
								Path: "/health",
								Port: 8080,
								HttpHeaders: []score.HttpProbeHttpHeadersElem{
									{Name: Ref("Custom-Header"), Value: Ref("Ops!")},
								},
							},
						},
					},
				},
			},
			Extensions:        &extensions.HumanitecExtensionsSpec{},
			WorkloadSourceURL: "",
			Output: &humanitec.CreateDeploymentDeltaRequest{
				Metadata: humanitec.DeltaMetadata{EnvID: envID, Name: name},
				Modules: humanitec.ModuleDeltas{
					Add: map[string]map[string]interface{}{
						"backend": {
							"profile": "humanitec/default-module",
							"spec": map[string]interface{}{
								"annotations": map[string]interface{}{
									"humanitec.io/managed-by": "score-humanitec",
								},
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
			Source: &score.Workload{
				Metadata: score.WorkloadMetadata{
					"name": "test",
				},
				Containers: score.WorkloadContainers{
					"backend": score.Container{
						Variables: map[string]string{
							"DEBUG":             "${resources.env.DEBUG}",
							"LOGS_LEVEL":        "${pod.debug.level}",
							"ORDERS_SERVICE":    "http://${resources.orders.name}:${resources.orders.port}/api",
							"CONNECTION_STRING": "postgresql://${resources.db.host}:${resources.db.port}/${resources.db.name}",
							"DOMAIN_NAME":       "${resources.dns.domain}",
							"EXTERNAL_RESOURCE": "${resources.external-resource.name}",
							"SENSITIVE_BUCKET":  "${resources.sensitive-bucket.name}",
						},
						Files: []score.ContainerFilesElem{
							{
								Target: "/etc/backend/config.yaml",
								Mode:   Ref("666"),
								Content: Ref(`---
DEBUG: ${resources.env.DEBUG}
`),
							},
							{
								Target:   "/etc/backend/config.yml",
								Mode:     Ref("666"),
								Content:  Ref("DEBUG: ${resources.env.DEBUG}"),
								NoExpand: Ref(true),
							},
							{
								Target:   "/etc/backend/config.txt",
								Mode:     Ref("666"),
								Source:   Ref("testdata/config.txt"),
								NoExpand: Ref(true),
							},
						},
						Volumes: []score.ContainerVolumesElem{
							{
								Source:   "${resources.data}",
								Path:     Ref("sub/path"),
								Target:   "/mnt/data",
								ReadOnly: Ref(true),
							},
						},
					},
				},
				Resources: map[string]score.Resource{
					"env": {
						Metadata: score.ResourceMetadata{
							"annotations": map[string]interface{}{
								AnnotationLabelResourceId: "externals.should-ignore-this-one",
							},
						},
						Type: "environment",
					},
					"dns": {
						Type: "dns",
						Params: map[string]interface{}{
							"test": "value",
						},
					},
					"data": {
						Type: "volume",
					},
					"db": {
						Metadata: score.ResourceMetadata{
							"annotations": map[string]interface{}{
								AnnotationLabelResourceId: "externals.annotations-db-id",
							},
						},
						Type:  "postgres",
						Class: Ref("large"),
						Params: map[string]interface{}{
							"extensions": map[string]interface{}{
								"uuid-ossp": map[string]interface{}{
									"schema":  "uuid_schema",
									"version": "1.1",
								},
							},
						},
					},
					"orders": {
						Type: "service",
					},
					"external-resource": {
						Metadata: score.ResourceMetadata{
							"annotations": map[string]interface{}{
								AnnotationLabelResourceId: "modules.test-module.externals.test-resource",
							},
						},
						Type: "some-type",
					},
					"route": {
						Type: "route",
						Params: map[string]interface{}{
							"host": "${resources.dns.host}",
						},
					},
					"sensitive-bucket": {
						Metadata: score.ResourceMetadata{
							"annotations": map[string]interface{}{
								AnnotationLabelResourceId: "shared.sensitive-bucket",
							},
						},
						Type:  "bucket",
						Class: Ref("sensitive"),
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
					"db": extensions.HumanitecResourceSpec{
						Scope: "shared",
					},
					"dns": extensions.HumanitecResourceSpec{
						Scope: "shared",
					},
				},
			},
			WorkloadSourceURL: "https://test.com",
			Output: &humanitec.CreateDeploymentDeltaRequest{
				Metadata: humanitec.DeltaMetadata{EnvID: envID, Name: name},
				Modules: humanitec.ModuleDeltas{
					Add: map[string]map[string]interface{}{
						"test": {
							"profile": "test-org/test-module",
							"spec": map[string]interface{}{
								"annotations": map[string]interface{}{
									"humanitec.io/managed-by":      "score-humanitec",
									"humanitec.io/workload-source": "https://test.com",
								},
								"containers": map[string]interface{}{
									"backend": map[string]interface{}{
										"id": "backend",
										"variables": map[string]interface{}{
											"DEBUG":             "${values.DEBUG}",
											"LOGS_LEVEL":        "${pod.debug.level}",
											"ORDERS_SERVICE":    "http://${modules.orders.service.name}:${modules.orders.service.port}/api",
											"CONNECTION_STRING": "postgresql://${externals.annotations-db-id.host}:${externals.annotations-db-id.port}/${externals.annotations-db-id.name}",
											"DOMAIN_NAME":       "${shared.dns.domain}",
											"EXTERNAL_RESOURCE": "${modules.test-module.externals.test-resource.name}",
											"SENSITIVE_BUCKET":  "${shared.sensitive-bucket-class-sensitive.name}",
										},
										"files": map[string]interface{}{
											"/etc/backend/config.yaml": map[string]interface{}{
												"mode":  "666",
												"value": "---\nDEBUG: ${values.DEBUG}\n",
											},
											"/etc/backend/config.yml": map[string]interface{}{
												"mode":  "666",
												"value": "DEBUG: $\\{resources.env.DEBUG}",
											},
											"/etc/backend/config.txt": map[string]interface{}{
												"mode":  "666",
												"value": "Mounted\nFile\nContent\n$\\{resources.env.DEBUG}",
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
									"type":  "volume",
									"class": "default",
								},
								"annotations-db-id": map[string]interface{}{
									"type":  "postgres",
									"class": "large",
									"params": map[string]interface{}{
										"extensions": map[string]interface{}{
											"uuid-ossp": map[string]interface{}{
												"schema":  "uuid_schema",
												"version": "1.1",
											},
										},
									},
								},
								"route": map[string]interface{}{
									"type":  "route",
									"class": "default",
									"params": map[string]interface{}{
										"host": "${shared.dns.host}",
									},
								},
							},
						},
					},
				},
				Shared: []humanitec.UpdateAction{
					{
						Operation: "add",
						Path:      "/sensitive-bucket-class-sensitive",
						Value: map[string]interface{}{
							"type":  "bucket",
							"class": "sensitive",
							"id":    "shared.sensitive-bucket",
						},
					},
					{
						Operation: "add",
						Path:      "/dns",
						Value: map[string]interface{}{
							"type":  "dns",
							"class": "default",
							"params": map[string]interface{}{
								"test": "value",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			res, err := ConvertSpec(name, envID, "", tt.WorkloadSourceURL, tt.Source, tt.Extensions)

			if tt.Error != nil {
				// On Error
				//
				assert.ErrorContains(t, err, tt.Error.Error())
			} else {
				// On Success
				//
				assert.NoError(t, err)
				expectedShared := tt.Output.Shared
				tt.Output.Shared = nil
				actualShared := res.Shared
				res.Shared = nil
				assert.Equal(t, tt.Output, res)
				assert.ElementsMatch(t, expectedShared, actualShared)
			}
		})
	}
}
