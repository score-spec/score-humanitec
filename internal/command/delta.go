/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package command

import (
	"context"
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
	deltaCmd.Flags().StringVar(&apiToken, "token", "", "Humanitec API authentication token")
	deltaCmd.MarkFlagRequired("token")
	deltaCmd.Flags().StringVar(&orgID, "org", "", "Organization ID")
	deltaCmd.MarkFlagRequired("org")
	deltaCmd.Flags().StringVar(&appID, "app", "", "Application ID")
	deltaCmd.MarkFlagRequired("app")
	deltaCmd.Flags().StringVar(&envID, "env", "", "Environment ID")
	deltaCmd.MarkFlagRequired("env")

	deltaCmd.Flags().BoolVar(&deploy, "deploy", false, "Trigger a new delta deployment at the end")
	deltaCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable diagnostic messages (written to STDERR)")

	rootCmd.AddCommand(deltaCmd)
}

var deltaCmd = &cobra.Command{
	Use:   "delta",
	Short: "Creates Humanitec deployment delta from the source SCORE file",
	RunE:  delta,
}

func delta(cmd *cobra.Command, args []string) error {
	if !verbose {
		log.SetOutput(io.Discard)
	}

	// Load SCORE spec and extensions
	//
	spec, ext, err := loadSpec(scoreFile, overridesFile, extensionsFile)
	if err != nil {
		return err
	}

	// Prepare a new deployment
	//
	log.Print("Preparing a new deployment...\n")
	delta, err := humanitec.ConvertSpec("Auto-generated (SCORE)", envID, spec, ext)
	if err != nil {
		return fmt.Errorf("preparing new deployment: %w", err)
	}

	// Create a new deployment delta
	//
	log.Print("Creating a new deployment delta...\n")
	client, err := api.NewClient(apiUrl, apiToken, http.DefaultClient)
	if err != nil {
		return err
	}
	res, err := client.CreateDelta(context.Background(), orgID, appID, delta)
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
		_, err := client.StartDeployment(context.Background(), orgID, appID, envID, &ht.StartDeploymentRequest{
			DeltaID: res.ID,
			Comment: "Auto-deployment (SCORE)",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
