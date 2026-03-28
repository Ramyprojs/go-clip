package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Ramyprojs/goclip/internal/db"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete all clips from clipboard history",
	Long:  "Delete every stored clip from clipboard history after confirmation.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := db.OpenDB("")
		if err != nil {
			return err
		}
		defer func() {
			_ = store.CloseDB()
		}()

		clips, err := store.GetAllClips()
		if err != nil {
			return err
		}

		confirmed, err := confirmClear(cmd, len(clips))
		if err != nil {
			return err
		}

		if !confirmed {
			fmt.Fprintln(cmd.OutOrStdout(), "Clear cancelled.")
			return nil
		}

		if err := store.ClearAll(); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Cleared clipboard history.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}

func confirmClear(cmd *cobra.Command, count int) (bool, error) {
	fmt.Fprintf(cmd.OutOrStdout(), "This will delete all %d clips. Are you sure? [y/N] ", count)

	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(response))
	return answer == "y" || answer == "yes", nil
}
