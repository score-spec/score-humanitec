/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package client

import (
	"context"

	humanitec "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
)

// Client describes Humanitec API client functionality.
type Client interface {

	// Resources
	//
	ListResourceTypes(ctx context.Context, orgID string) ([]humanitec.ResourceType, error)

	// Deployment Deltas
	//
	CreateDelta(ctx context.Context, orgID, appID string, delta *humanitec.CreateDeploymentDeltaRequest) (*humanitec.DeploymentDelta, error)
	UpdateDelta(ctx context.Context, orgID string, appID string, deltaID string, deltas []*humanitec.UpdateDeploymentDeltaRequest) (*humanitec.DeploymentDelta, error)

	// Deployments
	//
	StartDeployment(ctx context.Context, orgID, appID, envID string, retry bool, deployment *humanitec.StartDeploymentRequest) (*humanitec.Deployment, error)
}
