package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <index>",
	Short: "Delete a clip from clipboard history",
	Long:  "Delete a clip from clipboard history by its displayed list index.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("parse index: %w", err)
		}

		if index <= 0 {
			return errors.New("index must be greater than 0")
		}

		store, err := openStore()
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

		if len(clips) == 0 {
			return errors.New("no clips found")
		}

		if index > len(clips) {
			return fmt.Errorf("clip #%d not found", index)
		}

		confirmed, err := confirmDelete(cmd, index)
		if err != nil {
			return err
		}

		if !confirmed {
			fmt.Fprintln(cmd.OutOrStdout(), "Deletion cancelled.")
			return nil
		}

		if err := store.DeleteClip(clips[index-1].ID); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Deleted clip #%d.\n", index)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func confirmDelete(cmd *cobra.Command, index int) (bool, error) {
	fmt.Fprintf(cmd.OutOrStdout(), "Delete clip #%d? [y/N] ", index)

	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(response))
	return answer == "y" || answer == "yes", nil
}
