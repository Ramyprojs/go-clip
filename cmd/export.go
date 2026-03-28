package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ramyprojs/goclip/internal/clip"
	"github.com/Ramyprojs/goclip/internal/db"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export clipboard history to a file",
	Long:  "Export full clipboard history to a file in JSON or plain text format.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(exportOutput) == "" {
			return errors.New("output path is required")
		}

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

		data, err := buildExportData(clips, exportFormat)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(exportOutput), 0o755); err != nil {
			return fmt.Errorf("create export directory: %w", err)
		}

		if err := os.WriteFile(exportOutput, data, 0o644); err != nil {
			return fmt.Errorf("write export file: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Exported %d clips to %s\n", len(clips), exportOutput)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "Export format: json or txt")
	exportCmd.Flags().StringVar(&exportOutput, "output", "", "Path to the export file")
	rootCmd.AddCommand(exportCmd)
}

func buildExportData(clips []clip.Clip, format string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return buildJSONExport(clips)
	case "txt":
		return buildTextExport(clips), nil
	default:
		return nil, fmt.Errorf("unsupported export format %q", format)
	}
}

func buildJSONExport(clips []clip.Clip) ([]byte, error) {
	data, err := json.MarshalIndent(clips, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal clips as json: %w", err)
	}

	return append(data, '\n'), nil
}

func buildTextExport(clips []clip.Clip) []byte {
	var builder strings.Builder
	for i, entry := range clips {
		builder.WriteString(fmt.Sprintf("%d. %s [%s]\n", i+1, entry.CopiedAt.Format("2006-01-02 15:04:05"), entry.Source))
		builder.WriteString(entry.Content)
		if !strings.HasSuffix(entry.Content, "\n") {
			builder.WriteByte('\n')
		}

		if i < len(clips)-1 {
			builder.WriteString("\n---\n\n")
		}
	}

	return []byte(builder.String())
}
