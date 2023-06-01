/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	loader "github.com/score-spec/score-go/loader"
	schema "github.com/score-spec/score-go/schema"
	score "github.com/score-spec/score-go/types"
	"github.com/score-spec/score-humanitec/internal/humanitec"
	"github.com/score-spec/score-humanitec/internal/humanitec/extensions"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
	yaml "gopkg.in/yaml.v3"
)

func init() {
	runCmd.Flags().StringVarP(&scoreFile, "file", "f", scoreFileDefault, "Source SCORE file")
	runCmd.Flags().StringVar(&overridesFile, "overrides", overridesFileDefault, "Overrides file")
	runCmd.Flags().StringVar(&extensionsFile, "extensions", extensionsFileDefault, "Extensions file")
	runCmd.Flags().StringVar(&envID, "env", "", "Environment ID")
	runCmd.MarkFlagRequired("env")

	runCmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "DEPRECATED: Disables Score file schema validation.")
	runCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable diagnostic messages (written to STDERR)")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Translate the SCORE file to Humanitec deployment delta",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
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
	delta, err := humanitec.ConvertSpec("Auto-generated (SCORE)", envID, spec, ext)
	if err != nil {
		return fmt.Errorf("preparing new deployment: %w", err)
	}

	// Output resulting deployment delta
	//
	tmp, err := json.MarshalIndent(&delta, "", "  ")
	if err != nil {
		return err
	}
	os.Stdout.Write(tmp)

	return nil
}

func loadSpec(scoreFile, overridesFile, extensionsFile string, skipValidation bool) (*score.WorkloadSpec, *extensions.HumanitecExtensionsSpec, error) {
	// Open source file
	//
	log.Printf("Reading '%s'...\n", scoreFile)
	var err error
	var src *os.File
	if src, err = os.Open(scoreFile); err != nil {
		return nil, nil, err
	}

	// Parse SCORE spec
	//
	log.Print("Parsing SCORE spec...\n")
	var srcMap map[string]interface{}
	if err := loader.ParseYAML(&srcMap, src); err != nil {
		return nil, nil, err
	}

	// Apply overrides (optional)
	//
	if overridesFile != "" {
		log.Printf("Checking '%s'...\n", overridesFile)
		if ovr, err := os.Open(overridesFile); err == nil {
			defer ovr.Close()

			log.Print("Applying SCORE overrides...\n")
			var ovrMap map[string]interface{}
			if err := loader.ParseYAML(&ovrMap, ovr); err != nil {
				return nil, nil, err
			}
			if err := mergo.MergeWithOverwrite(&srcMap, ovrMap); err != nil {
				return nil, nil, fmt.Errorf("applying overrides fom '%s': %w", overridesFile, err)
			}
		} else if !os.IsNotExist(err) || overridesFile != overridesFileDefault {
			return nil, nil, err
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
				return nil, nil, fmt.Errorf("parsing extensions file '%s': %w", extensionsFile, err)
			}
		} else if !os.IsNotExist(err) || extensionsFile != extensionsFileDefault {
			return nil, nil, err
		}
	}

	// Validate SCORE spec
	//
	if !skipValidation {
		log.Print("Validating SCORE spec...\n")
		if res, err := schema.Validate(gojsonschema.NewGoLoader(srcMap)); err != nil {
			return nil, nil, fmt.Errorf("validating workload spec: %w", err)
		} else if !res.Valid() {
			for _, valErr := range res.Errors() {
				log.Println(valErr.String())
				if err == nil {
					err = errors.New(valErr.String())
				}
			}
			return nil, nil, fmt.Errorf("validating workload spec: %w", err)
		}
	}

	// Convert SCORE spec
	//

	var spec score.WorkloadSpec
	log.Print("Validating SCORE spec...\n")
	if err = mapstructure.Decode(srcMap, &spec); err != nil {
		return nil, nil, fmt.Errorf("validating workload spec: %w", err)
	}

	var ext extensions.HumanitecExtensionsSpec
	if err = mapstructure.Decode(extMap, &ext); err != nil {
		return nil, nil, fmt.Errorf("validating extensions spec: %w", err)
	}

	return &spec, &ext, nil
}
