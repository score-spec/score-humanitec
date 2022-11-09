/*
Apache Score
Copyright 2022 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package types

import "time"

type StartDeploymentRequest struct {
	DeltaID string `json:"delta_id"`
	Comment string `json:"comment"`
}

type Deployment struct {
	ID    string `json:"id"`
	EnvID string `json:"env_id"`

	FromID  string `json:"from_id"`
	DeltaID string `json:"delta_id"`
	Comment string `json:"comment"`

	Status          string    `json:"status"`
	StatusChangedAt time.Time `json:"status_changed_at"`

	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}
