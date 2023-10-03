/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package command

import (
	"github.com/spf13/cobra"
)

func init() {
	draftCmd.Flags().StringVarP(&scoreFile, "file", "f", scoreFileDefault, "Source SCORE file")
	draftCmd.Flags().StringVar(&overridesFile, "overrides", overridesFileDefault, "Overrides file")
	draftCmd.Flags().StringVar(&extensionsFile, "extensions", extensionsFileDefault, "Extensions file")
	draftCmd.Flags().StringVar(&workloadSourceURL, "workload-source-url", "", "URL of file that is managing the humanitec workload")
	draftCmd.Flags().StringVar(&uiUrl, "ui-url", uiUrlDefault, "Humanitec API endpoint")
	draftCmd.Flags().StringVar(&apiUrl, "api-url", apiUrlDefault, "Humanitec API endpoint")
	draftCmd.Flags().StringVar(&apiToken, "token", "", "Humanitec API authentication token")
	draftCmd.MarkFlagRequired("token")
	draftCmd.Flags().StringVar(&orgID, "org", "", "Organization ID")
	draftCmd.MarkFlagRequired("org")
	draftCmd.Flags().StringVar(&appID, "app", "", "Application ID")
	draftCmd.MarkFlagRequired("app")
	draftCmd.Flags().StringVar(&envID, "env", "", "Environment ID")
	draftCmd.MarkFlagRequired("env")

	draftCmd.Flags().StringArrayVarP(&overrideParams, "property", "p", nil, "Overrides selected property value")

	draftCmd.Flags().BoolVar(&deploy, "deploy", false, "Trigger a new draft deployment at the end")
	draftCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable diagnostic messages (written to STDERR)")

	rootCmd.AddCommand(draftCmd)
}

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "DEPRECATED - use 'delta' instead - creates Humanitec deployment draft from the source SCORE file",
	RunE:  draft,
}

func draft(cmd *cobra.Command, args []string) error {
	return delta(cmd, args)
}
