package store

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"strings"
)

type SearchResult struct {
	DocumentID string
	Path       string
	Distance   float64
	Chunks     []ChunkMatch
}

type ChunkMatch struct {
	ChunkIndex int
	StartLine  int
	EndLine    int
	Content    string
	Distance   float64
}

func SearchSummaryEmbeddings(
	db *sql.DB,
	summaryModelID string,
	embeddingModelID string,
	query []float32,
	limit int,
	excludePath string,
) ([]SearchResult, error) {
	queryBlob, err := encodeSearchQuery(query)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`SELECT
			d.id,
			d.path,
			vector_distance_cos(dse.embedding, ?) AS distance
		FROM document_summary_embeddings dse
		JOIN document_summaries ds ON ds.id = dse.summary_id
		JOIN documents d ON d.id = ds.document_id
		WHERE ds.summary_model_id = ?
			AND dse.embedding_model_id = ?
			AND (? = '' OR d.path != ?)
		ORDER BY distance ASC
		LIMIT ?`,
		queryBlob,
		summaryModelID,
		embeddingModelID,
		excludePath,
		excludePath,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readSummarySearchResults(rows)
}

func SearchDocumentChunksForDocuments(
	db *sql.DB,
	embeddingModelID string,
	query []float32,
	documentIDs []string,
	limit int,
) (map[string][]ChunkMatch, error) {
	if len(documentIDs) == 0 || limit < 1 {
		return map[string][]ChunkMatch{}, nil
	}

	queryBlob, err := encodeSearchQuery(query)
	if err != nil {
		return nil, err
	}

	placeholders := strings.TrimRight(strings.Repeat("?, ", len(documentIDs)), ", ")
	args := make([]any, 0, len(documentIDs)+3)
	args = append(args, queryBlob, embeddingModelID)
	for _, documentID := range documentIDs {
		args = append(args, documentID)
	}
	args = append(args, limit)

	rows, err := db.Query(
		fmt.Sprintf(
			`WITH scored AS (
				SELECT
					dec.document_id,
					dec.chunk_index,
					dec.start_line,
					dec.end_line,
					dec.content,
					vector_distance_cos(dec.embedding, ?) AS distance
				FROM document_embedding_chunks dec
				WHERE dec.embedding_model_id = ?
					AND dec.document_id IN (%s)
			),
			ranked AS (
				SELECT
					document_id,
					chunk_index,
					start_line,
					end_line,
					content,
					distance,
					ROW_NUMBER() OVER (
						PARTITION BY document_id
						ORDER BY distance ASC, chunk_index ASC
					) AS rank
				FROM scored
			)
			SELECT
				document_id,
				chunk_index,
				start_line,
				end_line,
				content,
				distance
			FROM ranked
			WHERE rank <= ?
			ORDER BY document_id ASC, rank ASC`,
			placeholders,
		),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readChunkMatches(rows)
}

func encodeSearchQuery(query []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, query); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readSummarySearchResults(rows *sql.Rows) ([]SearchResult, error) {
	var results []SearchResult
	for rows.Next() {
		var row SearchResult
		if err := rows.Scan(&row.DocumentID, &row.Path, &row.Distance); err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

func readChunkMatches(rows *sql.Rows) (map[string][]ChunkMatch, error) {
	results := make(map[string][]ChunkMatch)
	for rows.Next() {
		var documentID string
		var row ChunkMatch
		if err := rows.Scan(
			&documentID,
			&row.ChunkIndex,
			&row.StartLine,
			&row.EndLine,
			&row.Content,
			&row.Distance,
		); err != nil {
			return nil, err
		}
		results[documentID] = append(results[documentID], row)
	}

	return results, rows.Err()
}
