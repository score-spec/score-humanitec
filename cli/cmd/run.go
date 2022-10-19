package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	loader "github.com/score-spec/score-go/loader"
	score "github.com/score-spec/score-go/types"
	"github.com/score-spec/score-humanitec/internal/humanitec"
	"github.com/score-spec/score-humanitec/internal/humanitec/extensions"
	api "github.com/score-spec/score-humanitec/internal/humanitec_go/client"
	"github.com/score-spec/score-humanitec/internal/humanitec_go/types"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
)

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

func init() {
	runCmd.Flags().StringVarP(&scoreFile, "file", "f", scoreFileDefault, "Source SCORE file")
	runCmd.Flags().StringVar(&overridesFile, "overrides", overridesFileDefault, "Overrides SCORE file")
	runCmd.Flags().StringVar(&extensionsFile, "extensions", overridesFileDefault, "Extensions SCORE file")
	runCmd.Flags().StringVar(&apiUrl, "url", apiUrlDefault, "Humanitec API endpoint")
	runCmd.Flags().StringVar(&apiToken, "token", "", "Humanitec API authentication token")
	runCmd.MarkFlagRequired("token")
	runCmd.Flags().StringVar(&orgID, "org", "", "Organization ID")
	runCmd.MarkFlagRequired("org")
	runCmd.Flags().StringVar(&appID, "app", "", "Application ID")
	runCmd.MarkFlagRequired("app")
	runCmd.Flags().StringVar(&envID, "env", "", "Environment ID")
	runCmd.MarkFlagRequired("env")

	runCmd.Flags().BoolVar(&deploy, "deploy", false, "Trigger a new draft deployment at the end")
	runCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable diagnostic messages (written to STDERR)")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Translate the SCORE file to Humanitec deployment",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	if !verbose {
		log.SetOutput(io.Discard)
	}

	// Open source file
	//
	log.Printf("Reading '%s'...\n", scoreFile)
	var err error
	var src *os.File
	if src, err = os.Open(scoreFile); err != nil {
		return err
	}

	// Parse SCORE spec
	//
	log.Print("Parsing SCORE spec...\n")
	var srcMap map[string]interface{}
	if err := loader.ParseYAML(&srcMap, src); err != nil {
		return err
	}

	// Apply overrides (optional)
	//
	if overridesFile != "" {
		log.Printf("Checking '%s'...\n", overridesFile)
		if ovr, err := os.Open(overridesFile); err != nil {
			log.Print("Applying SCORE overrides...\n")
			var ovrMap map[string]interface{}
			if err := loader.ParseYAML(&ovrMap, ovr); err != nil {
				return err
			}
			if err := mergo.MergeWithOverwrite(&srcMap, ovrMap); err != nil {
				return fmt.Errorf("applying overrides fom '%s': %w", overridesFile, err)
			}
		} else if !os.IsNotExist(err) || overridesFile != overridesFileDefault {
			return err
		}
	}

	// Load extensions (optional)
	//
	var extMap = make(map[string]interface{})
	if extensionsFile != "" {
		log.Printf("Checking '%s'...\n", extensionsFile)
		if extFile, err := os.Open(extensionsFile); err == nil {
			defer extFile.Close()

			log.Print("Loading SCORE extensions...\n")
			if err := yaml.NewDecoder(extFile).Decode(extMap); err != nil {
				return fmt.Errorf("parsing extensions file '%s': %w", extensionsFile, err)
			}
		} else if !os.IsNotExist(err) || extensionsFile != extensionsFileDefault {
			return err
		}
	}

	// Validate SCORE spec
	//
	var spec score.WorkloadSpec
	log.Print("Validating SCORE spec...\n")
	if err = mapstructure.Decode(srcMap, &spec); err != nil {
		return fmt.Errorf("validating workload spec: %w", err)
	}

	var extSpec extensions.HumanitecExtensionsSpec
	if err = mapstructure.Decode(extMap, &extSpec); err != nil {
		return fmt.Errorf("validating extensions spec: %w", err)
	}

	// Prepare a new deployment
	//
	log.Print("Preparing a new deployment...\n")
	delta, err := humanitec.ConvertSpec("Auto-generated (SCORE)", envID, &spec, &extSpec)
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
		_, err := client.StartDeployment(context.Background(), orgID, appID, envID, &types.StartDeploymentRequest{
			DeltaID: res.ID,
			Comment: "Auto-deployment (SCORE)",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
