/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package types

type CreateDeploymentDeltaRequest struct {
	Metadata DeltaMetadata  `json:"metadata,omitempty"`
	Modules  ModuleDeltas   `json:"modules,omitempty"`
	Shared   []UpdateAction `json:"shared,omitempty"`
}

type UpdateDeploymentDeltaRequest struct {
	Modules ModuleDeltas   `json:"modules,omitempty"`
	Shared  []UpdateAction `json:"shared,omitempty"`
}

type DeploymentDelta struct {
	ID string `json:"id"`

	Metadata DeltaMetadata  `json:"metadata,omitempty"`
	Modules  ModuleDeltas   `json:"modules,omitempty"`
	Shared   []UpdateAction `json:"shared,omitempty"`
}

type DeltaMetadata struct {
	EnvID string `json:"env_id,omitempty"`
	Name  string `json:"name,omitempty"`
	Url   string `json:"url,omitempty"`

	Contributers []string `json:"contributers,omitempty"`

	Shared   bool `json:"shared,omitempty"`
	Archived bool `json:"archived,omitempty"`

	CreatedBy  string `json:"created_by,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	ModifiedAt string `json:"last_modified_at,omitempty"`
}

type ModuleDeltas struct {
	Add    map[string]map[string]interface{} `json:"add,omitempty"`
	Remove []string                          `json:"remove,omitempty"`
	Update map[string][]UpdateAction         `json:"update,omitempty"`
}

type UpdateAction struct {
	Path      string      `json:"path"`
	Operation string      `json:"op"`
	From      string      `json:"from,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}
