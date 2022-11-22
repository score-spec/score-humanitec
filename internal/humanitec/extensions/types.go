/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package extensions

// HumanitecExtensionsSpec is a set of extra definitions supported by Humanitec.
//
// YAML example:
//
//	apiVersion: humanitec.org/v1b1
//	profile: "humanitec/default-module"
//	spec:
//	  "labels":
//	    "tags.datadoghq.com/env": "${resources.env.DATADOG_ENV}"
//	  "ingress":
//	    rules:
//	      "${resources.dns}":
//	        http:
//	          "/":
//	            type: prefix
//	            port: 80
//	resources:
//	  db:
//	    scope: external
//	  dns:
//	    scope: shared
type HumanitecExtensionsSpec struct {
	ApiVersion string                  `mapstructure:"apiVersion"`
	Profile    string                  `mapstructure:"profile"`
	Spec       map[string]interface{}  `mapstructure:"spec"`
	Resources  HumanitecResourcesSpecs `mapstructure:"resources"`
}

// HumanitecResourcesSpecs is a map of workload resources specifications.
type HumanitecResourcesSpecs map[string]HumanitecResourceSpec

// HumanitecResourceSpec is a resource specification.
type HumanitecResourceSpec struct {
	Scope string `mapstructure:"scope"`
}
