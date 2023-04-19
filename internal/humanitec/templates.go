/*
Apache Score
Copyright 2022 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package humanitec

import (
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/mapstructure"

	score "github.com/score-spec/score-go/types"
	extensions "github.com/score-spec/score-humanitec/internal/humanitec/extensions"
)

// templatesContext ia an utility type that provides a context for '${...}' templates substitution
type templatesContext map[string]string

// buildContext initializes a new templatesContext instance
func buildContext(metadata score.WorkloadMeta, resources score.ResourcesSpecs, ext extensions.HumanitecResourcesSpecs) (templatesContext, error) {
	var ctx = make(map[string]string)

	var metadataMap = make(map[string]interface{})
	if decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &metadataMap,
	}); err != nil {
		return nil, err
	} else {
		decoder.Decode(metadata)
		for key, val := range metadataMap {
			var ref = fmt.Sprintf("metadata.%s", key)
			if _, exists := ctx[ref]; exists {
				return nil, fmt.Errorf("ambiguous property reference '%s'", ref)
			}
			ctx[ref] = fmt.Sprintf("%v", val)
		}
	}

	for resName, res := range resources {
		var source string
		switch res.Type {
		case "environment":
			source = "values"
		case "service":
			source = fmt.Sprintf("modules.%s", resName)
		default:
			if res.Type == "workload" {
				log.Println("Warning: 'workload' is a reserved resource type. Its usage may lead to compatibility issues with future releases of this application.")
			}
			resId, hasAnnotation := res.Metadata.Annotations[AnnotationLabelResourceId]
			// DEPRECATED: Should use resource annotations instead
			if resExt, hasMeta := ext[resName]; hasMeta && !hasAnnotation {
				if resExt.Scope == "" || resExt.Scope == "external" {
					resId = fmt.Sprintf("externals.%s", resName)
				} else if resExt.Scope == "shared" {
					resId = fmt.Sprintf("shared.%s", resName)
				}
			}
			// END (DEPRECATED)

			if resId != "" {
				source = resId
			} else {
				source = fmt.Sprintf("externals.%s", resName)
			}
		}
		ctx[fmt.Sprintf("resources.%s", resName)] = source

		for propName := range res.Properties {
			var ref = fmt.Sprintf("resources.%s.%s", resName, propName)
			if _, exists := ctx[ref]; exists {
				return nil, fmt.Errorf("ambiguous property reference '%s'", ref)
			}
			var sourceProp string
			switch res.Type {
			case "service":
				sourceProp = fmt.Sprintf("service.%s", propName)
			default:
				sourceProp = propName
			}
			ctx[ref] = fmt.Sprintf("${%s.%s}", source, sourceProp)
		}
	}

	return ctx, nil
}

// SubstituteAll replaces all matching '${...}' templates in map keys and string values recursively.
func (context templatesContext) SubstituteAll(src map[string]interface{}) map[string]interface{} {
	var dst = make(map[string]interface{}, 0)

	for key, val := range src {
		key = context.Substitute(key)
		switch v := val.(type) {
		case string:
			val = context.Substitute(v)
		case map[string]interface{}:
			val = context.SubstituteAll(v)
		}
		dst[key] = val
	}

	return dst
}

// Substitute replaces all matching '${...}' templates in a source string
func (context templatesContext) Substitute(src string) string {
	return os.Expand(src, context.mapVar)
}

// MapVar replaces objects and properties references with corresponding values
// Returns an empty string if the reference can't be resolved
func (context templatesContext) mapVar(ref string) string {
	if ref == "" {
		return ""
	}

	// NOTE: os.Expand(..) would invoke a callback function with "$" as an argument for escaped sequences.
	//       "$${abc}" is treated as "$$" pattern and "{abc}" static text.
	//       The first segment (pattern) would trigger a callback function call.
	//       By returning "$" value we would ensure that escaped sequences would remain in the source text.
	//       For example "$${abc}" would result in "${abc}" after os.Expand(..) call.
	if ref == "$" {
		return ref
	}

	if res, ok := context[ref]; ok {
		return res
	}

	log.Printf("Warning: Can not resolve '%s'. Resource or property is not declared.", ref)
	return fmt.Sprintf("${%s}", ref)
}
