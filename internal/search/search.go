package search

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/Ramyprojs/goclip/internal/clip"
)

type scoredClip struct {
	clip  clip.Clip
	score int
}

// FuzzySearch returns clips that match the query, ranked by match quality and position.
func FuzzySearch(query string, clips []clip.Clip) []clip.Clip {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery == "" {
		return append([]clip.Clip(nil), clips...)
	}

	matches := make([]scoredClip, 0, len(clips))
	for _, entry := range clips {
		score := scoreContent(normalizedQuery, strings.ToLower(entry.Content))
		if score < 0 {
			continue
		}

		matches = append(matches, scoredClip{
			clip:  entry,
			score: score,
		})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			if matches[i].clip.CopiedAt.Equal(matches[j].clip.CopiedAt) {
				return matches[i].clip.ID > matches[j].clip.ID
			}

			return matches[i].clip.CopiedAt.After(matches[j].clip.CopiedAt)
		}

		return matches[i].score > matches[j].score
	})

	results := make([]clip.Clip, len(matches))
	for i, match := range matches {
		results[i] = match.clip
	}

	return results
}

func scoreContent(query, content string) int {
	if idx := strings.Index(content, query); idx >= 0 {
		lengthPenalty := utf8.RuneCountInString(content) - utf8.RuneCountInString(query)
		return 1000 - (idx * 10) - lengthPenalty
	}

	return scoreSubsequence(query, content)
}

func scoreSubsequence(query, content string) int {
	queryRunes := []rune(query)
	contentRunes := []rune(content)
	if len(queryRunes) == 0 || len(contentRunes) == 0 {
		return -1
	}

	queryIndex := 0
	start := -1
	previous := -1
	gaps := 0

	for i, r := range contentRunes {
		if queryIndex >= len(queryRunes) || r != queryRunes[queryIndex] {
			continue
		}

		if start == -1 {
			start = i
		}

		if previous >= 0 {
			gaps += i - previous - 1
		}

		previous = i
		queryIndex++
		if queryIndex == len(queryRunes) {
			break
		}
	}

	if queryIndex != len(queryRunes) {
		return -1
	}

	spanPenalty := (previous - start + 1) - len(queryRunes)
	return 500 - (start * 5) - (gaps * 3) - spanPenalty
}
