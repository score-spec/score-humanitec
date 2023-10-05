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
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"

	score "github.com/score-spec/score-go/types"
	extensions "github.com/score-spec/score-humanitec/internal/humanitec/extensions"
)

var (
	placeholderRegEx = regexp.MustCompile(`\$(\$|{([a-zA-Z0-9.\-_\[\]"'#]+)})`)
)

// templatesContext ia an utility type that provides a context for '${...}' templates substitution
type templatesContext struct {
	meta       map[string]interface{}
	resources  score.ResourcesSpecs
	extensions extensions.HumanitecResourcesSpecs
}

// buildContext initializes a new templatesContext instance
func buildContext(metadata score.WorkloadMeta, resources score.ResourcesSpecs, ext extensions.HumanitecResourcesSpecs) (*templatesContext, error) {
	var metadataMap = make(map[string]interface{})
	if decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &metadataMap,
	}); err != nil {
		return nil, err
	} else {
		decoder.Decode(metadata)
	}

	return &templatesContext{
		meta:       metadataMap,
		resources:  resources,
		extensions: ext,
	}, nil
}

// SubstituteAll replaces all matching '${...}' templates in map keys and string values recursively.
func (ctx *templatesContext) SubstituteAll(src map[string]interface{}) map[string]interface{} {
	var dst = make(map[string]interface{}, 0)

	for key, val := range src {
		key = ctx.Substitute(key)
		switch v := val.(type) {
		case string:
			val = ctx.Substitute(v)
		case map[string]interface{}:
			val = ctx.SubstituteAll(v)
		}
		dst[key] = val
	}

	return dst
}

// Substitute replaces all matching '${...}' templates in a source string
func (ctx *templatesContext) Substitute(src string) string {
	return placeholderRegEx.ReplaceAllStringFunc(src, func(str string) string {
		// WORKAROUND: ReplaceAllStringFunc(..) does not provide match details
		//             https://github.com/golang/go/issues/5690
		var matches = placeholderRegEx.FindStringSubmatch(str)

		// SANITY CHECK
		if len(matches) != 3 {
			log.Printf("Error: could not find a proper match in previously captured string fragment")
			return src
		}

		// EDGE CASE: Captures "$$" sequences and empty templates "${}"
		if matches[2] == "" {
			return matches[1]
		}

		return ctx.mapVar(matches[2])
	})
}

// Escape replaces all matching '${...}' templates in a source string with '$\{...}'
func (ctx *templatesContext) Escape(src string) string {
	return placeholderRegEx.ReplaceAllStringFunc(src, func(str string) string {
		// WORKAROUND: ReplaceAllStringFunc(..) does not provide match details
		//             https://github.com/golang/go/issues/5690
		var matches = placeholderRegEx.FindStringSubmatch(str)

		// SANITY CHECK
		if len(matches) != 3 {
			log.Printf("Error: could not find a proper match in previously captured string fragment")
			return src
		}

		if strings.HasPrefix(matches[1], "{") {
			return fmt.Sprintf("$\\%s", matches[1])
		}

		return matches[0]
	})
}

// MapVar replaces objects and properties references with corresponding values
// Returns an empty string if the reference can't be resolved
func (ctx *templatesContext) mapVar(ref string) string {
	if ref == "" || ref == "$" {
		return ref
	}

	var segments = strings.SplitN(ref, ".", 2)
	switch segments[0] {
	case "metadata":
		if len(segments) == 2 {
			if val, exists := ctx.meta[segments[1]]; exists {
				return fmt.Sprintf("%v", val)
			}
		}

	case "resources":
		if len(segments) == 2 {
			segments = strings.SplitN(segments[1], ".", 2)
			var resName = segments[0]
			if res, exists := ctx.resources[resName]; exists {
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
					if resExt, hasMeta := ctx.extensions[resName]; hasMeta && !hasAnnotation {
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

				if len(segments) == 1 {
					return source
				} else {
					var propName = segments[1]
					var sourceProp string
					switch res.Type {
					case "service":
						sourceProp = fmt.Sprintf("service.%s", propName)
					default:
						sourceProp = propName
					}
					return fmt.Sprintf("${%s.%s}", source, sourceProp)
				}
			}
		}
	}

	log.Printf("Warning: Can not resolve '%s'. Resource or property is not declared.", ref)
	return fmt.Sprintf("${%s}", ref)
}
