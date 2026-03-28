package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Ramyprojs/goclip/internal/clip"
	"github.com/spf13/cobra"
)

const defaultListLimit = 20

var (
	listLimit int
	listJSON  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show clipboard history",
	Long:  "Show clipboard history in the terminal or as JSON.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("list does not accept positional arguments")
		}

		if listLimit < 0 {
			return errors.New("limit cannot be negative")
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

		clips = limitClips(clips, listLimit)
		if listJSON {
			return writeClipsJSON(cmd, clips)
		}

		if len(clips) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No clips found.")
			return nil
		}

		for i, entry := range clips {
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"%d. %s  %s\n",
				i+1,
				entry.CopiedAt.Format("2006-01-02 15:04:05"),
				previewClip(entry.Content, configuredPreviewLength()),
			)
		}

		return nil
	},
}

func init() {
	listCmd.Flags().IntVar(&listLimit, "limit", defaultListLimit, "Maximum number of clips to display (0 shows all)")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Print clips as JSON")
	rootCmd.AddCommand(listCmd)
}

func limitClips(clips []clip.Clip, limit int) []clip.Clip {
	if limit == 0 || len(clips) <= limit {
		return clips
	}

	return clips[:limit]
}

func writeClipsJSON(cmd *cobra.Command, clips []clip.Clip) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(clips); err != nil {
		return fmt.Errorf("encode clips as json: %w", err)
	}

	return nil
}

func previewClip(content string, limit int) string {
	compact := strings.Join(strings.Fields(content), " ")
	if compact == "" || limit <= 0 {
		return compact
	}

	if utf8.RuneCountInString(compact) <= limit {
		return compact
	}

	runes := []rune(compact)
	if limit <= 3 {
		return string(runes[:limit])
	}

	return string(runes[:limit-3]) + "..."
}
