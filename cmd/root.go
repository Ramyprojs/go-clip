package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:           "goclip",
	Short:         "A terminal clipboard history manager",
	Long:          "goclip stores clipboard history locally and lets you search it from the terminal.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute runs the root Cobra command.
func Execute() error {
	return rootCmd.Execute()
}
