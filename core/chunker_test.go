package core

import "testing"

func TestChunkDocumentTracksLines(t *testing.T) {
	text := "Alpha sentence. Beta sentence.\n\nGamma sentence.\nDelta sentence."

	chunks := ChunkDocument(text, 4, 1)
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	if chunks[0].StartLine != 1 {
		t.Fatalf("expected first chunk to start at line 1, got %d", chunks[0].StartLine)
	}

	last := chunks[len(chunks)-1]
	if last.EndLine != 4 {
		t.Fatalf("expected last chunk to end at line 4, got %d", last.EndLine)
	}
}

func TestChunkDocumentKeepsChunkIndex(t *testing.T) {
	text := "One two three four five six seven eight nine ten."

	chunks := ChunkDocument(text, 3, 1)
	for i, chunk := range chunks {
		if chunk.Index != i {
			t.Fatalf("expected chunk index %d, got %d", i, chunk.Index)
		}
	}
}

func TestNormalizeText(t *testing.T) {
	input := "  hello   world \r\n\r\n\r\n second\tline "
	got := NormalizeText(input)
	want := "hello world\n\nsecond line"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestShouldSummarizeQuery(t *testing.T) {
	short := "short query"
	if ShouldSummarizeQuery(short) {
		t.Fatalf("short query should not be summarized")
	}

	long := ""
	for range LongQueryWordLimit + 1 {
		long += "word "
	}
	if !ShouldSummarizeQuery(long) {
		t.Fatalf("long query should be summarized")
	}
}
