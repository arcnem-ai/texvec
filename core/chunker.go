package core

import (
	"regexp"
	"strings"
)

var sentenceRegexp = regexp.MustCompile(`[^.!?\n]+(?:[.!?]+|$)`)

type DocumentChunk struct {
	Index     int
	StartLine int
	EndLine   int
	Content   string
}

type textSegment struct {
	Text      string
	StartLine int
	EndLine   int
	WordCount int
}

func ChunkDocument(text string, targetWords, overlapWords int) []DocumentChunk {
	segments := splitIntoSegments(text)
	if len(segments) == 0 {
		return nil
	}

	var chunks []DocumentChunk
	start := 0

	for start < len(segments) {
		end := start
		words := 0

		for end < len(segments) && (words < targetWords || end == start) {
			words += segments[end].WordCount
			end++
		}

		piece := segments[start:end]
		chunks = append(chunks, DocumentChunk{
			Index:     len(chunks),
			StartLine: piece[0].StartLine,
			EndLine:   piece[len(piece)-1].EndLine,
			Content:   joinSegmentText(piece),
		})

		if end >= len(segments) {
			break
		}

		nextStart := end
		overlap := 0
		for nextStart > start && overlap < overlapWords {
			nextStart--
			overlap += segments[nextStart].WordCount
		}
		if nextStart <= start {
			nextStart = end
		}
		start = nextStart
	}

	return chunks
}

func splitIntoSegments(text string) []textSegment {
	lines := strings.Split(normalizeLineEndings(text), "\n")
	var segments []textSegment

	for i, line := range lines {
		lineNumber := i + 1
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		matches := sentenceRegexp.FindAllString(trimmed, -1)
		if len(matches) == 0 {
			matches = []string{trimmed}
		}

		for _, match := range matches {
			cleaned := whitespaceRegexp.ReplaceAllString(strings.TrimSpace(match), " ")
			if cleaned == "" {
				continue
			}

			segments = append(segments, textSegment{
				Text:      cleaned,
				StartLine: lineNumber,
				EndLine:   lineNumber,
				WordCount: CountWords(cleaned),
			})
		}
	}

	return segments
}

func joinSegmentText(segments []textSegment) string {
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		parts = append(parts, segment.Text)
	}

	return strings.Join(parts, " ")
}
