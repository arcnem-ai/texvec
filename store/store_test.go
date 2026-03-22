package store

import (
	"database/sql"
	"math"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open("file:" + filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return db
}

func insertSearchFixture(t *testing.T, db *sql.DB, path, hash, summaryModel, embeddingModel string, summaryEmbedding []float32, chunks []ChunkInput) Document {
	t.Helper()

	document, _, err := UpsertDocument(db, path, hash)
	if err != nil {
		t.Fatalf("upsert document: %v", err)
	}

	if _, err := InsertSummaryModel(db, summaryModel); err != nil {
		t.Fatalf("insert summary model: %v", err)
	}
	if _, err := InsertEmbeddingModel(db, embeddingModel, len(summaryEmbedding)); err != nil {
		t.Fatalf("insert embedding model: %v", err)
	}

	summary, err := UpsertDocumentSummary(db, document.ID, summaryModel, "summary")
	if err != nil {
		t.Fatalf("upsert summary: %v", err)
	}

	if err := UpsertDocumentSummaryEmbedding(db, summary.ID, embeddingModel, summaryEmbedding); err != nil {
		t.Fatalf("upsert summary embedding: %v", err)
	}

	if err := ReplaceDocumentChunks(db, document.ID, embeddingModel, chunks); err != nil {
		t.Fatalf("replace chunks: %v", err)
	}

	return document
}

func TestUpsertDocumentDetectsHashChange(t *testing.T) {
	db := setupTestDB(t)

	document, changed, err := UpsertDocument(db, "/tmp/a.txt", "hash-1")
	if err != nil {
		t.Fatalf("upsert document: %v", err)
	}
	if !changed {
		t.Fatalf("expected first insert to mark changed")
	}

	same, changed, err := UpsertDocument(db, "/tmp/a.txt", "hash-1")
	if err != nil {
		t.Fatalf("upsert document: %v", err)
	}
	if changed {
		t.Fatalf("expected same hash to avoid change flag")
	}
	if same.ID != document.ID {
		t.Fatalf("expected same document id, got %s vs %s", same.ID, document.ID)
	}

	updated, changed, err := UpsertDocument(db, "/tmp/a.txt", "hash-2")
	if err != nil {
		t.Fatalf("upsert document: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed hash to set changed flag")
	}
	if updated.ContentHash != "hash-2" {
		t.Fatalf("expected updated hash, got %s", updated.ContentHash)
	}
}

func TestReplaceDocumentChunksAndHasDocumentChunks(t *testing.T) {
	db := setupTestDB(t)

	document, _, err := UpsertDocument(db, "/tmp/a.txt", "hash")
	if err != nil {
		t.Fatalf("upsert document: %v", err)
	}
	if _, err := InsertEmbeddingModel(db, "embed", 4); err != nil {
		t.Fatalf("insert embedding model: %v", err)
	}

	hasChunks, err := HasDocumentChunks(db, document.ID, "embed")
	if err != nil {
		t.Fatalf("has chunks: %v", err)
	}
	if hasChunks {
		t.Fatalf("expected no chunks before insert")
	}

	err = ReplaceDocumentChunks(db, document.ID, "embed", []ChunkInput{
		{ChunkIndex: 0, StartLine: 1, EndLine: 2, Content: "chunk one", Embedding: []float32{1, 0, 0, 0}},
		{ChunkIndex: 1, StartLine: 2, EndLine: 3, Content: "chunk two", Embedding: []float32{0, 1, 0, 0}},
	})
	if err != nil {
		t.Fatalf("replace chunks: %v", err)
	}

	hasChunks, err = HasDocumentChunks(db, document.ID, "embed")
	if err != nil {
		t.Fatalf("has chunks: %v", err)
	}
	if !hasChunks {
		t.Fatalf("expected chunks after replace")
	}

	err = ReplaceDocumentChunks(db, document.ID, "embed", []ChunkInput{
		{ChunkIndex: 0, StartLine: 10, EndLine: 10, Content: "replacement", Embedding: []float32{0, 0, 1, 0}},
	})
	if err != nil {
		t.Fatalf("replace chunks: %v", err)
	}

	rows, err := db.Query(`SELECT chunk_index, start_line, end_line FROM document_embedding_chunks WHERE document_id = ? AND embedding_model_id = ? ORDER BY chunk_index`, document.ID, "embed")
	if err != nil {
		t.Fatalf("query chunks: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var chunkIndex, startLine, endLine int
		if err := rows.Scan(&chunkIndex, &startLine, &endLine); err != nil {
			t.Fatalf("scan chunk: %v", err)
		}
		count++
		if chunkIndex != 0 || startLine != 10 || endLine != 10 {
			t.Fatalf("unexpected chunk row: %d %d %d", chunkIndex, startLine, endLine)
		}
	}
	if count != 1 {
		t.Fatalf("expected replaced chunk set, got %d rows", count)
	}
}

func TestSearchAndMerge(t *testing.T) {
	db := setupTestDB(t)

	insertSearchFixture(
		t,
		db,
		"/tmp/a.txt",
		"hash-a",
		"summary",
		"embed",
		[]float32{1, 0, 0, 0},
		[]ChunkInput{
			{ChunkIndex: 0, StartLine: 1, EndLine: 2, Content: "alpha", Embedding: []float32{1, 0, 0, 0}},
		},
	)
	insertSearchFixture(
		t,
		db,
		"/tmp/b.txt",
		"hash-b",
		"summary",
		"embed",
		[]float32{0.8, 0.2, 0, 0},
		[]ChunkInput{
			{ChunkIndex: 0, StartLine: 1, EndLine: 1, Content: "beta", Embedding: []float32{0.9, 0.1, 0, 0}},
		},
	)

	summaryRows, err := SearchSummaryEmbeddings(db, "summary", "embed", []float32{1, 0, 0, 0}, 10, "")
	if err != nil {
		t.Fatalf("search summary embeddings: %v", err)
	}
	chunkRows, err := SearchDocumentChunks(db, "embed", []float32{1, 0, 0, 0}, 10, "")
	if err != nil {
		t.Fatalf("search document chunks: %v", err)
	}

	merged := MergeSearchRows(10, summaryRows, chunkRows)
	if len(merged) != 2 {
		t.Fatalf("expected 2 merged rows, got %d", len(merged))
	}
	if merged[0].Path != "/tmp/a.txt" {
		t.Fatalf("expected best match to be a.txt, got %s", merged[0].Path)
	}
	if merged[1].Distance < merged[0].Distance {
		t.Fatalf("expected distances to be sorted ascending")
	}
}

func TestGetDocumentSummaryEmbedding(t *testing.T) {
	db := setupTestDB(t)

	document, _, err := UpsertDocument(db, "/tmp/a.txt", "hash")
	if err != nil {
		t.Fatalf("upsert document: %v", err)
	}
	if _, err := InsertSummaryModel(db, "summary"); err != nil {
		t.Fatalf("insert summary model: %v", err)
	}
	if _, err := InsertEmbeddingModel(db, "embed", 4); err != nil {
		t.Fatalf("insert embedding model: %v", err)
	}

	summary, err := UpsertDocumentSummary(db, document.ID, "summary", "hello")
	if err != nil {
		t.Fatalf("upsert summary: %v", err)
	}
	original := []float32{0.1, 0.2, 0.3, 0.4}
	if err := UpsertDocumentSummaryEmbedding(db, summary.ID, "embed", original); err != nil {
		t.Fatalf("upsert summary embedding: %v", err)
	}

	got, err := GetDocumentSummaryEmbedding(db, summary.ID, "embed")
	if err != nil {
		t.Fatalf("get summary embedding: %v", err)
	}
	if len(got) != len(original) {
		t.Fatalf("expected %d floats, got %d", len(original), len(got))
	}
	for i := range original {
		if math.Abs(float64(original[i]-got[i])) > 1e-6 {
			t.Fatalf("expected %f at %d, got %f", original[i], i, got[i])
		}
	}
}

func TestListDocuments(t *testing.T) {
	db := setupTestDB(t)

	insertSearchFixture(
		t,
		db,
		"/tmp/a.txt",
		"hash-a",
		"summary",
		"embed-a",
		[]float32{1, 0, 0, 0},
		[]ChunkInput{{ChunkIndex: 0, StartLine: 1, EndLine: 1, Content: "one", Embedding: []float32{1, 0, 0, 0}}},
	)
	insertSearchFixture(
		t,
		db,
		"/tmp/a.txt",
		"hash-a",
		"summary",
		"embed-b",
		[]float32{1, 0, 0, 0},
		[]ChunkInput{{ChunkIndex: 0, StartLine: 1, EndLine: 1, Content: "one", Embedding: []float32{1, 0, 0, 0}}},
	)
	insertSearchFixture(
		t,
		db,
		"/tmp/b.txt",
		"hash-b",
		"summary",
		"embed-a",
		[]float32{1, 0, 0, 0},
		[]ChunkInput{{ChunkIndex: 0, StartLine: 1, EndLine: 1, Content: "two", Embedding: []float32{1, 0, 0, 0}}},
	)

	records, err := ListDocuments(db, "", 0)
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(records))
	}
	if len(records[0].Models) != 2 {
		t.Fatalf("expected first document to have 2 models, got %d", len(records[0].Models))
	}
}
