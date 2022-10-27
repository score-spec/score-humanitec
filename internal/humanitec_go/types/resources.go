/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package types

type ResourceType struct {
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Category      string                 `json:"category,omitempty"`
	InputsSchema  map[string]interface{} `json:"inputs_schema,omitempty"`
	OutputsSchema map[string]interface{} `json:"outputs_schema,omitempty"`
}
