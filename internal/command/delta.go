/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package command

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/score-spec/score-humanitec/internal/humanitec"
	api "github.com/score-spec/score-humanitec/internal/humanitec_go/client"
	ht "github.com/score-spec/score-humanitec/internal/humanitec_go/types"
	"github.com/spf13/cobra"
)

func init() {
	deltaCmd.Flags().StringVarP(&scoreFile, "file", "f", scoreFileDefault, "Source SCORE file")
	deltaCmd.Flags().StringVar(&overridesFile, "overrides", overridesFileDefault, "Overrides file")
	deltaCmd.Flags().StringVar(&extensionsFile, "extensions", extensionsFileDefault, "Extensions file")
	deltaCmd.Flags().StringVar(&uiUrl, "ui-url", uiUrlDefault, "Humanitec UI")
	deltaCmd.Flags().StringVar(&apiUrl, "api-url", apiUrlDefault, "Humanitec API endpoint")
	deltaCmd.Flags().StringVar(&deltaID, "delta", "", "The ID of an existing delta in Humanitec into which to merge the generated delta")

	deltaCmd.Flags().StringVar(&apiToken, "token", "", "Humanitec API authentication token")
	deltaCmd.MarkFlagRequired("token")
	deltaCmd.Flags().StringVar(&orgID, "org", "", "Organization ID")
	deltaCmd.MarkFlagRequired("org")
	deltaCmd.Flags().StringVar(&appID, "app", "", "Application ID")
	deltaCmd.MarkFlagRequired("app")
	deltaCmd.Flags().StringVar(&envID, "env", "", "Environment ID")
	deltaCmd.MarkFlagRequired("env")

	deltaCmd.Flags().StringArrayVarP(&overrideParams, "property", "p", nil, "Overrides selected property value")
	deltaCmd.Flags().StringVarP(&message, "message", "m", messageDefault, "Message")

	deltaCmd.Flags().BoolVar(&deploy, "deploy", false, "Trigger a new delta deployment at the end")
	deltaCmd.Flags().BoolVar(&retry, "retry", false, "Retry deployments when a deployment is currently in progress")
	deltaCmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "DEPRECATED: Disables Score file schema validation.")
	deltaCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable diagnostic messages (written to STDERR)")

	rootCmd.AddCommand(deltaCmd)
}

var deltaCmd = &cobra.Command{
	Use:   "delta",
	Short: "Creates or updates a Humanitec deployment delta from the source SCORE file",
	Long: `This command will translate the SCORE file into a Humanitec deployment delta and submit it to the Humanitec
environment specified by the --org, --app, and --env flags. If the --delta flag is provided, the generated delta will
be merged with the specified existing delta. The --deploy flag allows the deployment of the delta to be triggered.
`,
	RunE: delta,
}

func delta(cmd *cobra.Command, args []string) error {
	if !verbose {
		log.SetOutput(io.Discard)
	}

	// Load SCORE spec and extensions
	//
	spec, ext, err := loadSpec(scoreFile, overridesFile, extensionsFile, skipValidation)
	if err != nil {
		return err
	}

	// Prepare a new deployment
	//
	log.Print("Preparing a new deployment...\n")
	delta, err := humanitec.ConvertSpec(message, envID, spec, ext)
	if err != nil {
		return fmt.Errorf("preparing new deployment: %w", err)
	}

	client, err := api.NewClient(apiUrl, apiToken, http.DefaultClient)
	if err != nil {
		return err
	}

	var res *ht.DeploymentDelta
	if deltaID == "" {
		log.Print("Creating a new deployment delta...\n")
		res, err = client.CreateDelta(cmd.Context(), orgID, appID, delta)
	} else {
		log.Printf("Updating existing delta %s in place...\n", deltaID)
		updates := []*ht.UpdateDeploymentDeltaRequest{{Modules: delta.Modules, Shared: delta.Shared}}
		res, err = client.UpdateDelta(cmd.Context(), orgID, appID, deltaID, updates)
	}
	if err != nil {
		return err
	}
	res.Metadata.Url = fmt.Sprintf("%s/orgs/%s/apps/%s/envs/%s/draft/%s", uiUrl, orgID, appID, delta.Metadata.EnvID, res.ID)

	// Output resulting deployment delta
	//
	tmp, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	os.Stdout.Write(tmp)

	// Trigger the deployment (optional)
	//
	if deploy {
		log.Printf("Starting a new deployment for delta '%s'...\n", res.ID)
		_, err := client.StartDeployment(cmd.Context(), orgID, appID, envID, retry, &ht.StartDeploymentRequest{
			DeltaID: res.ID,
			Comment: message,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
