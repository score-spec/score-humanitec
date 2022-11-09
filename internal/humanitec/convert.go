/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package humanitec

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	score "github.com/score-spec/score-go/types"
	"github.com/score-spec/score-humanitec/internal/humanitec/extensions"
	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
)

// resourceRefRegex extracts the resource ID from the resource reference: '${resources.RESOURCE_ID}'
var resourceRefRegex = regexp.MustCompile(`\${resources\.(.+)}`)

// resourcesMap is an internal utility type to group some helper methods.
type resourcesMap struct {
	Spec map[string]score.ResourceSpec
	Meta extensions.HumanitecResourcesSpecs
}

// mapResourceVar maps resources properties references.
// Returns an empty string if the reference can't be resolved.
func (r resourcesMap) mapVar(ref string) string {
	if ref == "$" {
		return ref
	}

	var segments = strings.SplitN(ref, ".", 3)
	if segments[0] == "resources" && len(segments) == 3 {
		var resName = segments[1]
		var propName = segments[2]
		if res, ok := r.Spec[resName]; ok {
			if _, ok := res.Properties[propName]; ok {
				var envVar string
				switch res.Type {
				case "environment":
					envVar = fmt.Sprintf("values.%s", propName)
				case "workload":
					envVar = fmt.Sprintf("modules.%s.%s", resName, propName)
				default:
					var scope = "externals"
					if meta, exists := r.Meta[resName]; exists && meta.Scope == "shared" {
						scope = "shared"
					}
					envVar = fmt.Sprintf("%s.%s.%s", scope, resName, propName)
				}
				return fmt.Sprintf("${%s}", envVar)
			} else {
				log.Printf("Warning: Can not resolve '%s'. Property '%s' is not declared for '%s'.", ref, propName, resName)
			}
		} else {
			log.Printf("Warning: Can not resolve '%s'. Resource '%s' is not declared.", ref, resName)
		}
	}

	return fmt.Sprintf("${%s}", ref)
}

// getProbeDetails extracts an httpGet probe details from the source spec.
// Returns nil if the source spec is empty.
func getProbeDetails(probe *score.ContainerProbeSpec) map[string]interface{} {
	if probe.HTTPGet.Path == "" {
		return nil
	}

	var res = map[string]interface{}{
		"type": "http",
		"path": probe.HTTPGet.Path,
		"port": probe.HTTPGet.Port,
	}

	if len(probe.HTTPGet.HTTPHeaders) > 0 {
		var hdrs = map[string]string{}
		for _, hdr := range probe.HTTPGet.HTTPHeaders {
			hdrs[hdr.Name] = hdr.Value
		}
		res["headers"] = hdrs
	}

	return res
}

// convertContainerSpec extracts a container details from the source spec.
func convertContainerSpec(name string, spec score.ContainerSpec, resources score.ResourcesSpecs, meta extensions.HumanitecResourcesSpecs) (map[string]interface{}, error) {
	var resourcesSpec = resourcesMap{
		Spec: resources,
		Meta: meta,
	}
	var containerSpec = map[string]interface{}{
		"id": name,
	}
	if spec.Image != "" {
		containerSpec["image"] = spec.Image
	}
	if len(spec.Command) > 0 {
		containerSpec["command"] = spec.Command
	}
	if len(spec.Args) > 0 {
		containerSpec["args"] = spec.Args
	}
	if len(spec.Variables) > 0 {
		var envVars = make(map[string]interface{}, len(spec.Variables))
		for key, val := range spec.Variables {
			envVars[key] = os.Expand(val, resourcesSpec.mapVar)
		}
		containerSpec["variables"] = envVars
	}
	if len(spec.Resources.Requests) > 0 || len(spec.Resources.Limits) > 0 {
		containerSpec["resources"] = map[string]interface{}{
			"requests": spec.Resources.Requests,
			"limits":   spec.Resources.Limits,
		}
	}
	if probe := getProbeDetails(&spec.LivenessProbe); len(probe) > 0 {
		containerSpec["liveness_probe"] = probe
	}
	if probe := getProbeDetails(&spec.ReadinessProbe); len(probe) > 0 {
		containerSpec["readiness_probe"] = probe
	}
	if len(spec.Files) > 0 {
		var files = map[string]interface{}{}
		for _, f := range spec.Files {
			files[f.Target] = map[string]interface{}{
				"mode":  f.Mode,
				"value": os.Expand(strings.Join(f.Content, "\n"), resourcesSpec.mapVar),
			}
		}
		containerSpec["files"] = files
	}
	if len(spec.Volumes) > 0 {
		var volumes = map[string]interface{}{}
		for _, vol := range spec.Volumes {
			var source = resourceRefRegex.ReplaceAllString(vol.Source, "externals.$1")
			volumes[vol.Target] = map[string]interface{}{
				"id":        source,
				"sub_path":  vol.Path,
				"read_only": vol.ReadOnly,
			}
		}
		containerSpec["volume_mounts"] = volumes
	}

	return containerSpec, nil
}

// ConvertSpec converts SCORE specification into Humanitec deployment delta.
func ConvertSpec(name, envID string, spec *score.WorkloadSpec, ext *extensions.HumanitecExtensionsSpec) (*humanitec.CreateDeploymentDeltaRequest, error) {
	var containers = make(map[string]interface{}, len(spec.Containers))
	for cName, cSpec := range spec.Containers {
		if container, err := convertContainerSpec(cName, cSpec, spec.Resources, ext.Resources); err == nil {
			containers[cName] = container
		} else {
			return nil, fmt.Errorf("processing container specification for '%s': %w", cName, err)
		}
	}

	var workloadSpec = map[string]interface{}{
		"containers": containers,
	}
	if len(spec.Service.Ports) > 0 {
		var ports = map[string]interface{}{}
		for pName, pSpec := range spec.Service.Ports {
			var proto = pSpec.Protocol
			if proto == "" {
				proto = "TCP" // Defaults to "TCP"
			}
			var targetPport = pSpec.TargetPort
			if targetPport == 0 {
				targetPport = pSpec.Port // Defaults to the published port
			}
			ports[pName] = map[string]interface{}{
				"protocol":       proto,
				"service_port":   pSpec.Port,
				"container_port": targetPport,
			}
		}
		workloadSpec["service"] = map[string]interface{}{
			"ports": ports,
		}
	}

	if ext != nil && len(ext.Service.Routes) > 0 {
		var rules = map[string]interface{}{}
		for proto, pRoutes := range ext.Service.Routes {
			for path, rSpec := range pRoutes {
				var from = resourceRefRegex.ReplaceAllString(rSpec.From, "externals.$1")
				var proto = strings.ToLower(proto)
				rules[from] = map[string]interface{}{
					proto: map[string]interface{}{
						path: map[string]interface{}{
							"port": rSpec.Port,
							"type": rSpec.Type,
						},
					},
				}
			}
		}
		workloadSpec["ingress"] = map[string]interface{}{
			"rules": rules,
		}
	}

	var workload = map[string]interface{}{
		"profile": "humanitec/default-module",
		"spec":    workloadSpec,
	}

	var externals = map[string]interface{}{}
	for name, res := range spec.Resources {
		if meta, exists := ext.Resources[name]; !exists || meta.Scope == "" || meta.Scope == "external" {
			if res.Type != "workload" && res.Type != "environment" {
				externals[name] = map[string]interface{}{
					"type": res.Type,
				}
			}
		}
	}
	if len(externals) > 0 {
		workload["externals"] = externals
	}

	var shared []humanitec.UpdateAction
	for name, res := range spec.Resources {
		if meta, exists := ext.Resources[name]; exists && meta.Scope == "shared" {
			if shared == nil {
				shared = make([]humanitec.UpdateAction, 0)
			}
			shared = append(shared, humanitec.UpdateAction{
				Operation: "add",
				Path:      "/" + name,
				Value: map[string]interface{}{
					"type": res.Type,
				},
			})
		}
	}

	var res = humanitec.CreateDeploymentDeltaRequest{
		Metadata: humanitec.DeltaMetadata{
			Name:  name,
			EnvID: envID,
		},
		Modules: humanitec.ModuleDeltas{
			Add: map[string]map[string]interface{}{
				spec.Metadata.Name: workload,
			},
		},
		Shared: shared,
	}

	return &res, nil
}
