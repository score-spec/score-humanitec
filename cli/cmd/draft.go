package cmd

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
	draftCmd.Flags().StringVarP(&scoreFile, "file", "f", scoreFileDefault, "Source SCORE file")
	draftCmd.Flags().StringVar(&overridesFile, "overrides", overridesFileDefault, "Overrides file")
	draftCmd.Flags().StringVar(&extensionsFile, "extensions", extensionsFileDefault, "Extensions file")
	draftCmd.Flags().StringVar(&apiUrl, "url", apiUrlDefault, "Humanitec API endpoint")
	draftCmd.Flags().StringVar(&apiToken, "token", "", "Humanitec API authentication token")
	draftCmd.MarkFlagRequired("token")
	draftCmd.Flags().StringVar(&orgID, "org", "", "Organization ID")
	draftCmd.MarkFlagRequired("org")
	draftCmd.Flags().StringVar(&appID, "app", "", "Application ID")
	draftCmd.MarkFlagRequired("app")
	draftCmd.Flags().StringVar(&envID, "env", "", "Environment ID")
	draftCmd.MarkFlagRequired("env")

	draftCmd.Flags().BoolVar(&deploy, "deploy", false, "Trigger a new draft deployment at the end")
	draftCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable diagnostic messages (written to STDERR)")

	rootCmd.AddCommand(draftCmd)
}

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Creates Humanitec deployment draft from the source SCORE file",
	RunE:  draft,
}

func draft(cmd *cobra.Command, args []string) error {
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
	res.Metadata.Url = fmt.Sprintf("%s/orgs/%s/apps/%s/envs/%s/draft/%s", apiUrl, orgID, appID, delta.Metadata.EnvID, res.ID)

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
