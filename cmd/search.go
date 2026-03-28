package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Ramyprojs/goclip/internal/db"
	clipsearch "github.com/Ramyprojs/goclip/internal/search"
	"github.com/spf13/cobra"
)

const (
	searchPreviewLimit = 60
	ansiBold           = "\033[1m"
	ansiReset          = "\033[0m"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search clipboard history",
	Long:  "Search clipboard history using fuzzy matching.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.TrimSpace(strings.Join(args, " "))
		if query == "" {
			return errors.New("query cannot be empty")
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

		matches := clipsearch.FuzzySearch(query, clips)
		if len(matches) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No clips found matching '%s'\n", query)
			return nil
		}

		for i, entry := range matches {
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"%d. %s  %s\n",
				i+1,
				entry.CopiedAt.Format("2006-01-02 15:04:05"),
				highlightSearchPreview(entry.Content, query, searchPreviewLimit),
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func highlightSearchPreview(content, query string, limit int) string {
	compact := strings.Join(strings.Fields(content), " ")
	if compact == "" {
		return ""
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery == "" {
		return previewClip(compact, limit)
	}

	contentRunes := []rune(compact)
	lowerContentRunes := []rune(strings.ToLower(compact))
	queryRunes := []rune(normalizedQuery)

	if start := findSubstringMatch(lowerContentRunes, queryRunes); start >= 0 {
		positions := make([]int, len(queryRunes))
		for i := range queryRunes {
			positions[i] = start + i
		}

		return buildHighlightedPreview(contentRunes, positions, limit)
	}

	if positions := findSubsequenceMatch(lowerContentRunes, queryRunes); len(positions) > 0 {
		return buildHighlightedPreview(contentRunes, positions, limit)
	}

	return previewClip(compact, limit)
}

func findSubstringMatch(content, query []rune) int {
	if len(query) == 0 || len(query) > len(content) {
		return -1
	}

	for i := 0; i <= len(content)-len(query); i++ {
		matched := true
		for j := range query {
			if content[i+j] != query[j] {
				matched = false
				break
			}
		}

		if matched {
			return i
		}
	}

	return -1
}

func findSubsequenceMatch(content, query []rune) []int {
	if len(query) == 0 {
		return nil
	}

	positions := make([]int, 0, len(query))
	queryIndex := 0
	for i, r := range content {
		if queryIndex >= len(query) || r != query[queryIndex] {
			continue
		}

		positions = append(positions, i)
		queryIndex++
		if queryIndex == len(query) {
			return positions
		}
	}

	return nil
}

func buildHighlightedPreview(content []rune, positions []int, limit int) string {
	if len(content) == 0 {
		return ""
	}

	if limit <= 0 {
		limit = len(content)
	}

	first := positions[0]
	last := positions[len(positions)-1]
	windowStart, windowEnd := previewWindow(len(content), first, last, limit)

	relative := make([]int, 0, len(positions))
	for _, position := range positions {
		if position >= windowStart && position < windowEnd {
			relative = append(relative, position-windowStart)
		}
	}

	snippet := applyBold(content[windowStart:windowEnd], relative)
	if windowStart > 0 {
		snippet = "..." + snippet
	}

	if windowEnd < len(content) {
		snippet += "..."
	}

	return snippet
}

func previewWindow(totalLength, firstMatch, lastMatch, limit int) (int, int) {
	if limit >= totalLength {
		return 0, totalLength
	}

	matchSpan := lastMatch - firstMatch + 1
	if matchSpan >= limit {
		return firstMatch, firstMatch + limit
	}

	padding := (limit - matchSpan) / 2
	start := firstMatch - padding
	if start < 0 {
		start = 0
	}

	end := start + limit
	if end > totalLength {
		end = totalLength
		start = end - limit
	}

	return start, end
}

func applyBold(content []rune, positions []int) string {
	if len(content) == 0 {
		return ""
	}

	marks := make([]bool, len(content))
	for _, position := range positions {
		if position >= 0 && position < len(content) {
			marks[position] = true
		}
	}

	var builder strings.Builder
	inBold := false
	for i, r := range content {
		if marks[i] && !inBold {
			builder.WriteString(ansiBold)
			inBold = true
		}

		if !marks[i] && inBold {
			builder.WriteString(ansiReset)
			inBold = false
		}

		builder.WriteRune(r)
	}

	if inBold {
		builder.WriteString(ansiReset)
	}

	return builder.String()
}
