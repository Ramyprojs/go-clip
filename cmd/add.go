package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Ramyprojs/goclip/internal/clip"
	"github.com/Ramyprojs/goclip/internal/db"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [text]",
	Short: "Save a clip to clipboard history",
	Long:  "Save a clip to clipboard history from a command argument or piped stdin input.",
	Example: strings.Join([]string{
		`goclip add "some text"`,
		`echo "hello" | goclip add`,
	}, "\n"),
	RunE: func(cmd *cobra.Command, args []string) error {
		content, source, err := resolveAddInput(args)
		if err != nil {
			return err
		}

		store, err := db.OpenDB("")
		if err != nil {
			return err
		}
		defer func() {
			_ = store.CloseDB()
		}()

		entry := clip.Clip{
			Content: content,
			Source:  source,
		}

		if err := store.SaveClip(entry); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ Saved to clipboard history")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func resolveAddInput(args []string) (string, string, error) {
	if len(args) > 0 {
		content := strings.Join(args, " ")
		if strings.TrimSpace(content) == "" {
			return "", "", errors.New("clip content cannot be empty")
		}

		return content, "terminal", nil
	}

	piped, err := hasPipedInput()
	if err != nil {
		return "", "", err
	}

	if !piped {
		return "", "", errors.New("provide clip content as an argument or via stdin")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", "", fmt.Errorf("read stdin: %w", err)
	}

	content := strings.TrimRight(string(data), "\r\n")
	if strings.TrimSpace(content) == "" {
		return "", "", errors.New("clip content cannot be empty")
	}

	return content, "stdin", nil
}

func hasPipedInput() (bool, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("inspect stdin: %w", err)
	}

	return info.Mode()&os.ModeCharDevice == 0, nil
}
