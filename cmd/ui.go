package cmd

import (
	"github.com/Ramyprojs/goclip/internal/ui"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the interactive clipboard manager UI",
	Args:  cobra.NoArgs,
	RunE:  runUICommand,
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func runUICommand(cmd *cobra.Command, args []string) error {
	return ui.StartTUI()
}
