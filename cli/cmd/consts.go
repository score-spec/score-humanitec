package cmd

const (
	scoreFileDefault      = "./score.yaml"
	overridesFileDefault  = "./overrides.score.yaml"
	extensionsFileDefault = "./humanitec.score.yaml"
	apiUrlDefault         = "https://api.humanitec.io"
)

var (
	scoreFile      string
	overridesFile  string
	extensionsFile string
	apiUrl         string
	apiToken       string
	orgID          string
	appID          string
	envID          string

	deploy  bool
	verbose bool
)
