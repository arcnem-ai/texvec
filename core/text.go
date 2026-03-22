package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	DefaultChunkTargetWords  = 220
	DefaultChunkOverlapWords = 40
	LongQueryWordLimit       = 220
)

var blankLineRegexp = regexp.MustCompile(`\n{3,}`)
var whitespaceRegexp = regexp.MustCompile(`[ \t]+`)

type Document struct {
	Path           string
	RawText        string
	NormalizedText string
	ContentHash    string
}

func LoadDocument(path string) (Document, error) {
	if !SupportsDocumentPath(path) {
		return Document{}, fmt.Errorf("%s is not a supported document type", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}

	rawText := normalizeLineEndings(string(content))
	normalized := NormalizeText(rawText)
	if normalized == "" {
		return Document{}, fmt.Errorf("%s is empty after normalization", path)
	}

	hash := sha256.Sum256([]byte(rawText))

	return Document{
		Path:           path,
		RawText:        rawText,
		NormalizedText: normalized,
		ContentHash:    hex.EncodeToString(hash[:]),
	}, nil
}

func SupportsDocumentPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt", ".md", ".markdown":
		return true
	default:
		return false
	}
}

func NormalizeText(text string) string {
	lines := strings.Split(normalizeLineEndings(text), "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := whitespaceRegexp.ReplaceAllString(strings.TrimSpace(line), " ")
		cleaned = append(cleaned, trimmed)
	}

	result := strings.Join(cleaned, "\n")
	result = blankLineRegexp.ReplaceAllString(result, "\n\n")
	result = strings.TrimSpace(result)
	return result
}

func CountWords(text string) int {
	return len(strings.Fields(text))
}

func ShouldSummarizeQuery(text string) bool {
	return CountWords(text) > LongQueryWordLimit
}

func normalizeLineEndings(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}
