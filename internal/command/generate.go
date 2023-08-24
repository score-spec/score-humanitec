package command

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

// Embed the score.tmpl file
//
//go:embed score.tmpl
var scoreTemplate string

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a score.yaml template",
	Long:  `Generates a default score.yaml template in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use the values from the flags
		data := struct {
			Metadata    string
			Environment string
			Resource    string
		}{
			Metadata:    appMetadata,
			Environment: appEnvironment,
			Resource:    appResourceName,
		}

		// Parse the embedded template
		tmpl, err := template.New("score").Parse(scoreTemplate)
		if err != nil {
			fmt.Println("Error parsing template:", err)
			return
		}

		// Create or overwrite the score.yaml file
		file, err := os.Create("score.yaml")
		if err != nil {
			fmt.Println("Error creating score.yaml:", err)
			return
		}
		defer file.Close()

		// Execute the template and write to the file
		if err := tmpl.Execute(file, data); err != nil {
			fmt.Println("Error executing template:", err)
			return
		}

		fmt.Println("score.yaml generated successfully!")
	},
}

var (
	appMetadata     string
	appEnvironment  string
	appResourceName string
)

func init() {
	// Assuming rootCmd is defined elsewhere in your project
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&appMetadata, "metadata", "m", "hello-world", "The metadata description of your Workload.")
	generateCmd.Flags().StringVarP(&appResourceName, "resource-name", "r", "application", "Name of the application")
	generateCmd.Flags().StringVarP(&appEnvironment, "environment", "e", "development", "Environment of the application")
}
