package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/arcnem-ai/texvec/core"
	"github.com/arcnem-ai/texvec/store"
)

var initDB = store.InitDB
var ensureEmbeddingModel = core.EnsureEmbeddingModel
var ensureSummaryModel = core.EnsureSummaryModel

func runtimeOptions() core.RuntimeOptions {
	return core.RuntimeOptions{Verbose: verbose}
}

func chunkInputsFromDocument(document core.Document, embedder *core.TextEmbedder) ([]store.ChunkInput, error) {
	chunks := core.ChunkDocument(document.RawText, core.DefaultChunkTargetWords, core.DefaultChunkOverlapWords)
	if len(chunks) == 0 {
		chunks = []core.DocumentChunk{{
			Index:     0,
			StartLine: 1,
			EndLine:   strings.Count(document.RawText, "\n") + 1,
			Content:   document.NormalizedText,
		}}
	}

	inputs := make([]store.ChunkInput, 0, len(chunks))
	for _, chunk := range chunks {
		embedding, err := embedder.EmbedDocument(chunk.Content)
		if err != nil {
			return nil, fmt.Errorf("embed chunk %d: %w", chunk.Index, err)
		}

		inputs = append(inputs, store.ChunkInput{
			ChunkIndex: chunk.Index,
			StartLine:  chunk.StartLine,
			EndLine:    chunk.EndLine,
			Content:    chunk.Content,
			Embedding:  embedding,
		})
	}

	return inputs, nil
}

func withDB(fn func(*sql.DB) error) error {
	db, err := initDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return fn(db)
}
