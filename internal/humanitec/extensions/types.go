/*
Apache Score
Copyright 2022 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package extensions

// HumanitecExtensionsSpec is a set of extra definitions supported by Humanitec.
//
// YAML example:
//
//	apiVersion: humanitec.org/v1b1
//	service:
//	  routes:
//	    http:
//	      "/":
//	        from: ${resources.dns}
//	        type: prefix
//	        port: 80
type HumanitecExtensionsSpec struct {
	ApiVersion string                  `mapstructure:"apiVersion"`
	Service    HumanitecServiceSpec    `mapstructure:"service"`
	Resources  HumanitecResourcesSpecs `mapstructure:"resources"`
}

// HumanitecServiceSpec is a workload service specification.
type HumanitecServiceSpec struct {
	Routes HumanitecServiceRoutesSpecs `mapstructure:"routes"`
}

// HumanitecServiceRoutesSpecs is a map of service routes specifications for each network protocol.
type HumanitecServiceRoutesSpecs map[string]HumanitecServiceRoutePathsSpec

// HumanitecServiceRoutePathsSpec is a map of service routes specifications for each path.
type HumanitecServiceRoutePathsSpec map[string]HumanitecServiceRoutePathSpec

// HumanitecServiceRoutePathSpec is a service route specification for a single path.
type HumanitecServiceRoutePathSpec struct {
	From string `mapstructure:"from"`
	Type string `mapstructure:"type"`
	Port int    `mapstructure:"port"`
}

// HumanitecResourcesSpecs is a map of workload resources specifications.
type HumanitecResourcesSpecs map[string]HumanitecResourceSpec

// HumanitecResourceSpec is a resource specification.
type HumanitecResourceSpec struct {
	Scope string `mapstructure:"scope"`
}
