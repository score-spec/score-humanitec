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
	"os"
	"path/filepath"
	"strings"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/score-spec/score-humanitec/internal/humanitec"
	"github.com/score-spec/score-humanitec/internal/humanitec/extensions"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"

	yaml "gopkg.in/yaml.v3"

	loader "github.com/score-spec/score-go/loader"
	schema "github.com/score-spec/score-go/schema"
	score "github.com/score-spec/score-go/types"
)

func init() {
	runCmd.Flags().StringVarP(&scoreFile, "file", "f", scoreFileDefault, "Source SCORE file")
	runCmd.Flags().StringVar(&overridesFile, "overrides", overridesFileDefault, "Overrides file")
	runCmd.Flags().StringVar(&extensionsFile, "extensions", extensionsFileDefault, "Extensions file")
	runCmd.Flags().StringVar(&workloadSourceURL, "workload-source-url", "", "URL of file that is managing the humanitec workload")
	runCmd.Flags().StringVar(&envID, "env", "", "Environment ID")
	runCmd.MarkFlagRequired("env")

	runCmd.Flags().StringArrayVarP(&overrideParams, "property", "p", nil, "Overrides selected property value")
	runCmd.Flags().StringVarP(&message, "message", "m", messageDefault, "Message")

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
	baseDir := filepath.Dir(scoreFile)
	spec, ext, err := loadSpec(scoreFile, overridesFile, extensionsFile, skipValidation)
	if err != nil {
		return err
	}

	// Prepare a new deployment
	//
	log.Print("Preparing a new deployment...\n")
	delta, err := humanitec.ConvertSpec(message, envID, baseDir, workloadSourceURL, spec, ext)
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

	// Apply overrides from file (optional)
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

	// Apply overrides from command line (optional)
	//
	for _, pstr := range overrideParams {
		log.Print("Applying SCORE properties overrides...\n")

		jsonBytes, err := json.Marshal(srcMap)
		if err != nil {
			return nil, nil, fmt.Errorf("marshalling score spec: %w", err)
		}

		pmap := strings.SplitN(pstr, "=", 2)
		if len(pmap) <= 1 {
			var path = pmap[0]
			log.Printf("removing '%s'", path)
			if jsonBytes, err = sjson.DeleteBytes(jsonBytes, path); err != nil {
				return nil, nil, fmt.Errorf("removing '%s': %w", path, err)
			}
		} else {
			var path = pmap[0]
			var val interface{}
			if err := yaml.Unmarshal([]byte(pmap[1]), &val); err != nil {
				val = pmap[1]
			}

			log.Printf("overriding '%s' = '%s'", path, val)
			if jsonBytes, err = sjson.SetBytes(jsonBytes, path, val); err != nil {
				return nil, nil, fmt.Errorf("overriding '%s': %w", path, err)
			}
		}

		if err = json.Unmarshal(jsonBytes, &srcMap); err != nil {
			return nil, nil, fmt.Errorf("unmarshalling score spec: %w", err)
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
		if err := schema.Validate(srcMap); err != nil {
			return nil, nil, fmt.Errorf("validating workload spec: %w", err)
		}
	}

	// Convert SCORE spec
	//

	var spec score.WorkloadSpec
	log.Print("Applying SCORE spec...\n")
	if err = mapstructure.Decode(srcMap, &spec); err != nil {
		return nil, nil, fmt.Errorf("applying workload spec: %w", err)
	}

	var ext extensions.HumanitecExtensionsSpec
	if err = mapstructure.Decode(extMap, &ext); err != nil {
		return nil, nil, fmt.Errorf("applying extensions spec: %w", err)
	}

	return &spec, &ext, nil
}
