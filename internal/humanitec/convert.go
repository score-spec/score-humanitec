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
	"path/filepath"
	"strings"

	mergo "github.com/imdario/mergo"
	score "github.com/score-spec/score-go/types"
	extensions "github.com/score-spec/score-humanitec/internal/humanitec/extensions"
	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
)

const (
	AnnotationLabelResourceId = "score.humanitec.io/resId"
)

// parseResourceId extracts resource ID details from a resource reference string.
// Supported reference string formants:
//
//	{resId}
//	{externals|shared}.{resId}
//	modules.{workloadId}.{externals|shared}.{resId}
func parseResourceId(ref string) (workload, scope, resId string, err error) {
	var segments = strings.SplitN(ref, ".", 4)
	switch len(segments) {
	case 4:
		if segments[0] != "modules" {
			err = fmt.Errorf("invalid resource reference '%s': not supported", ref)
			return
		}
		workload = segments[1]
		scope = segments[2]
		resId = segments[3]
	case 3:
		if segments[0] != "modules" {
			err = fmt.Errorf("invalid resource reference '%s': not supported", ref)
			return
		}
	case 2:
		workload = ""
		scope = segments[0]
		resId = segments[1]
	case 1:
		workload = ""
		scope = ""
		resId = segments[0]
	default:
		workload = ""
		scope = ""
		resId = ""
	}
	return
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

// readFile reads a text file into memory
func readFile(baseDir, path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading '%s': %w", path, err)
	}

	return string(raw), nil
}

// mergeFileContent joins inline file contents with '\n' character (DEPRECATED)
func mergeFileContent(content interface{}, target string) (string, error) {
	switch val := content.(type) {
	case string:
		return val, nil
	case []interface{}:
		// TODO: Deprecated functionality
		log.Printf("Warning: The content for the '%s' file is provided in a deprecated format. Strings will be joined with '\\n' character.", target)
		var sb strings.Builder
		for _, str := range val {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprint(str))
		}
		return sb.String(), nil
		// END (TODO)
	default:
		return "", fmt.Errorf("can not use '%T' as a content for '%s': not supported", content, target)
	}
}

// convertFileMountSpec extracts a mount file details from the source spec.
func convertFileMountSpec(f *score.FileMountSpec, context *templatesContext, baseDir string) (string, map[string]interface{}, error) {
	var err error
	var content string

	if f.Source != "" {
		content, err = readFile(baseDir, f.Source)
	} else {
		content, err = mergeFileContent(f.Content, f.Target)
	}
	if err != nil {
		return "", nil, err
	}

	if f.NoExpand {
		content = context.Escape(content)
	} else {
		content = context.Substitute(content)
	}

	return f.Target,
		map[string]interface{}{
			"mode":  f.Mode,
			"value": content,
		},
		nil
}

// convertContainerSpec extracts a container details from the source spec.
func convertContainerSpec(name string, spec *score.ContainerSpec, context *templatesContext, baseDir string) (map[string]interface{}, error) {
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
			if target, mount, err := convertFileMountSpec(&f, context, baseDir); err == nil {
				files[target] = mount
			} else {
				return nil, err
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
func ConvertSpec(name, envID, baseDir, workloadSourceURL string, spec *score.WorkloadSpec, ext *extensions.HumanitecExtensionsSpec) (*humanitec.CreateDeploymentDeltaRequest, error) {
	ctx, err := buildContext(spec.Metadata, spec.Resources, ext.Resources)
	if err != nil {
		return nil, fmt.Errorf("preparing context: %w", err)
	}
	annotations := map[string]interface{}{
		managedByAnnotation: managedBy,
	}
	if workloadSourceURL != "" {
		annotations[workloadSourceAnnotation] = workloadSourceURL
	}

	var containers = make(map[string]interface{}, len(spec.Containers))
	for cName, cSpec := range spec.Containers {
		if container, err := convertContainerSpec(cName, &cSpec, ctx, baseDir); err == nil {
			containers[cName] = container
		} else {
			return nil, fmt.Errorf("processing container specification for '%s': %w", cName, err)
		}
	}

	var workloadSpec = map[string]interface{}{
		"annotations": annotations,
		"containers":  containers,
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
		var features = ctx.SubstituteAll(ext.Spec)
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
				log.Printf("Warning: Extensions for resources has been deprecated. Use '%s' resource annotation instead. Extensions are still configured for '%s'.\n", AnnotationLabelResourceId, name)
				if !hasAnnotation && (meta.Scope == "" || meta.Scope == "externals") {
					resId = fmt.Sprintf("externals.%s", name)
				} else if !hasAnnotation && meta.Scope == "shared" {
					resId = fmt.Sprintf("shared.%s", name)
				}
			}
			// END (DEPRECATED)
			var class = "default"
			if res.Class != "" {
				class = res.Class
			}
			if mod, scope, resName, err := parseResourceId(resId); err != nil {
				log.Printf("Warning: %v.\n", err)
			} else if mod == "" || mod == spec.Metadata.Name {
				if scope == "externals" {
					var extRes = map[string]interface{}{
						"type":  res.Type,
						"class": class,
					}
					if len(res.Params) > 0 {
						extRes["params"] = ctx.SubstituteAll(res.Params)
					}
					externals[resName] = extRes
				} else if scope == "shared" {
					var resName = strings.Replace(resId, "shared.", "", 1)
					var sharedRes = map[string]interface{}{
						"type":  res.Type,
						"class": class,
					}
					if len(res.Params) > 0 {
						sharedRes["params"] = ctx.SubstituteAll(res.Params)
					}
					shared = append(shared, humanitec.UpdateAction{
						Operation: "add",
						Path:      "/" + resName,
						Value:     sharedRes,
					})
				} else {
					log.Printf("Warning: invalid resource reference '%s': not supported.\n", resId)
				}
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
