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
	"strings"

	mergo "github.com/imdario/mergo"
	score "github.com/score-spec/score-go/types"
	extensions "github.com/score-spec/score-humanitec/internal/humanitec/extensions"
	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
)

const (
	AnnotationLabelResourceId = "score.humanitec.io/resId"
)

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
func convertContainerSpec(name string, spec *score.ContainerSpec, context *templatesContext) (map[string]interface{}, error) {
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
			envVars[key] = context.Substitute(val)
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
				"value": context.Substitute(strings.Join(f.Content, "\n")),
			}
		}
		containerSpec["files"] = files
	}
	if len(spec.Volumes) > 0 {
		var volumes = map[string]interface{}{}
		for _, vol := range spec.Volumes {
			volumes[vol.Target] = map[string]interface{}{
				"id":        context.Substitute(vol.Source),
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
	context, err := buildContext(spec.Metadata, spec.Resources, ext.Resources)
	if err != nil {
		return nil, fmt.Errorf("preparing context: %w", err)
	}

	var containers = make(map[string]interface{}, len(spec.Containers))
	for cName, cSpec := range spec.Containers {
		if container, err := convertContainerSpec(cName, &cSpec, &context); err == nil {
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

	if ext != nil && len(ext.Spec) > 0 {
		var features = context.SubstituteAll(ext.Spec)
		if err := mergo.Merge(&workloadSpec, features); err != nil {
			return nil, fmt.Errorf("applying workload profile features: %w", err)
		}
	}

	var profile = DefaultWorkloadProfile
	if ext != nil && ext.Profile != "" {
		profile = ext.Profile
	}

	var workload = map[string]interface{}{
		"profile": profile,
		"spec":    workloadSpec,
	}

	var externals = make(map[string]interface{})
	var shared = make([]humanitec.UpdateAction, 0)
	for name, res := range spec.Resources {
		switch res.Type {

		case "service", "environment":
			continue

		default:
			resId, hasAnnotation := res.Metadata.Annotations[AnnotationLabelResourceId]
			if resId == "" {
				resId = fmt.Sprintf("externals.%s", name)
			}

			// DEPRECATED: Should use resource annotations instead
			if meta, hasMeta := ext.Resources[name]; hasMeta {
				log.Printf("Warning: Extensions for resources has been deprecated. Use Score resource annotations instead. Extensions are stil configured for '%s'.\n", name)
				if !hasAnnotation && (meta.Scope == "" || meta.Scope == "externals") {
					resId = fmt.Sprintf("externals.%s", name)
				} else if !hasAnnotation && meta.Scope == "shared" {
					resId = fmt.Sprintf("shared.%s", name)
				}
			}
			// END (DEPRECATED)

			if strings.HasPrefix(resId, "externals.") {
				var resName = strings.Replace(resId, "externals.", "", 1)
				var extRes = map[string]interface{}{
					"type": res.Type,
				}
				if len(res.Params) > 0 {
					extRes["params"] = res.Params
				}
				externals[resName] = extRes
			} else if strings.HasPrefix(resId, "shared.") {
				var resName = strings.Replace(resId, "shared.", "", 1)
				var sharedRes = map[string]interface{}{
					"type": res.Type,
				}
				if len(res.Params) > 0 {
					sharedRes["params"] = res.Params
				}
				shared = append(shared, humanitec.UpdateAction{
					Operation: "add",
					Path:      "/" + resName,
					Value:     sharedRes,
				})
			} else {
				log.Printf("Warning: Invalid resource id value '%s'. Not supported.\n", resId)
			}
		}
	}
	if len(externals) > 0 {
		workload["externals"] = externals
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
	}
	if len(shared) > 0 {
		res.Shared = shared
	}

	return &res, nil
}
