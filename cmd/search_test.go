package cmd

import (
	"strings"
	"testing"

	"github.com/arcnem-ai/texvec/store"
)

func TestFormatSearchResultIncludesChunkPreviewWhenChunkWins(t *testing.T) {
	result := store.SearchResult{
		Path:     "/tmp/a.txt",
		Distance: 0.1234,
		Chunks: []store.ChunkMatch{
			{StartLine: 3, EndLine: 5, Distance: 0.1452, Content: "alpha beta gamma"},
			{StartLine: 9, EndLine: 9, Distance: 0.1666, Content: "delta"},
		},
	}

	formatted := formatSearchResult(result)
	if !strings.Contains(formatted, "0.1234  /tmp/a.txt") {
		t.Fatalf("expected primary result line, got %q", formatted)
	}
	if !strings.Contains(formatted, "0.1452  lines 3-5: alpha beta gamma") {
		t.Fatalf("expected first chunk preview, got %q", formatted)
	}
	if !strings.Contains(formatted, "0.1666  line 9: delta") {
		t.Fatalf("expected second chunk preview, got %q", formatted)
	}
}

func TestFormatSearchResultWithoutChunks(t *testing.T) {
	result := store.SearchResult{
		Path:     "/tmp/a.txt",
		Distance: 0.1000,
	}
	formatted := formatSearchResult(result)
	if formatted != "0.1000  /tmp/a.txt" {
		t.Fatalf("expected only primary result line, got %q", formatted)
	}
}
