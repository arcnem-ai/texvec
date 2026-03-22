package store

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"io"
	"time"
)

type ChunkInput struct {
	ChunkIndex int
	StartLine  int
	EndLine    int
	Content    string
	Embedding  []float32
}

func GetDocumentSummaryEmbedding(db *sql.DB, summaryID, embeddingModelID string) ([]float32, error) {
	var blob []byte
	err := db.QueryRow(
		`SELECT embedding FROM document_summary_embeddings WHERE summary_id = ? AND embedding_model_id = ?`,
		summaryID,
		embeddingModelID,
	).Scan(&blob)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return decodeEmbedding(blob)
}

func UpsertDocumentSummaryEmbedding(db *sql.DB, summaryID, embeddingModelID string, embedding []float32) error {
	blob, err := encodeEmbedding(embedding)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO document_summary_embeddings (summary_id, embedding_model_id, embedding, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(summary_id, embedding_model_id) DO UPDATE SET
			embedding = excluded.embedding,
			created_at = excluded.created_at`,
		summaryID,
		embeddingModelID,
		blob,
		time.Now().Unix(),
	)

	return err
}

func HasDocumentChunks(db *sql.DB, documentID, embeddingModelID string) (bool, error) {
	var exists int
	err := db.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM document_embedding_chunks
			WHERE document_id = ? AND embedding_model_id = ?
		)`,
		documentID,
		embeddingModelID,
	).Scan(&exists)

	return exists == 1, err
}

func ReplaceDocumentChunks(db *sql.DB, documentID, embeddingModelID string, chunks []ChunkInput) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`DELETE FROM document_embedding_chunks WHERE document_id = ? AND embedding_model_id = ?`,
		documentID,
		embeddingModelID,
	); err != nil {
		return err
	}

	now := time.Now().Unix()
	for _, chunk := range chunks {
		blob, err := encodeEmbedding(chunk.Embedding)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			`INSERT INTO document_embedding_chunks
			(document_id, embedding_model_id, chunk_index, start_line, end_line, content, embedding, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			documentID,
			embeddingModelID,
			chunk.ChunkIndex,
			chunk.StartLine,
			chunk.EndLine,
			chunk.Content,
			blob,
			now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetSummaryEmbeddingByPath(db *sql.DB, path, summaryModelID, embeddingModelID string) ([]float32, error) {
	var blob []byte
	err := db.QueryRow(
		`SELECT dse.embedding
		FROM document_summary_embeddings dse
		JOIN document_summaries ds ON ds.id = dse.summary_id
		JOIN documents d ON d.id = ds.document_id
		WHERE d.path = ? AND ds.summary_model_id = ? AND dse.embedding_model_id = ?`,
		path,
		summaryModelID,
		embeddingModelID,
	).Scan(&blob)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return decodeEmbedding(blob)
}

func encodeEmbedding(embedding []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, embedding); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeEmbedding(blob []byte) ([]float32, error) {
	reader := bytes.NewReader(blob)
	var values []float32
	for {
		var value float32
		if err := binary.Read(reader, binary.LittleEndian, &value); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}
