/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package command

const (
	scoreFileDefault      = "./score.yaml"
	overridesFileDefault  = "./overrides.score.yaml"
	extensionsFileDefault = "./humanitec.score.yaml"
	apiUrlDefault         = "https://api.humanitec.io"
	uiUrlDefault          = "https://app.humanitec.io"
	messageDefault        = "Auto-deployment (SCORE)"
)

var (
	scoreFile      string
	overridesFile  string
	extensionsFile string
	uiUrl          string
	apiUrl         string
	apiToken       string
	orgID          string
	appID          string
	envID          string

	overrideParams []string
	message        string

	// deltaID is a cli flag receiver used to support "score-humanitec delta --use foo"
	deltaID        string
	deploy         bool
	retry          bool
	skipValidation bool
	verbose        bool
)
