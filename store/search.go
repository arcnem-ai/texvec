package store

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"sort"
)

type SearchRow struct {
	DocumentID string
	Path       string
	Distance   float64
}

func SearchSummaryEmbeddings(
	db *sql.DB,
	summaryModelID string,
	embeddingModelID string,
	query []float32,
	limit int,
	excludePath string,
) ([]SearchRow, error) {
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

	return readSearchRows(rows)
}

func SearchDocumentChunks(
	db *sql.DB,
	embeddingModelID string,
	query []float32,
	limit int,
	excludePath string,
) ([]SearchRow, error) {
	queryBlob, err := encodeSearchQuery(query)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`SELECT
			d.id,
			d.path,
			MIN(vector_distance_cos(dec.embedding, ?)) AS distance
		FROM document_embedding_chunks dec
		JOIN documents d ON d.id = dec.document_id
		WHERE dec.embedding_model_id = ?
			AND (? = '' OR d.path != ?)
		GROUP BY d.id, d.path
		ORDER BY distance ASC
		LIMIT ?`,
		queryBlob,
		embeddingModelID,
		excludePath,
		excludePath,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readSearchRows(rows)
}

func MergeSearchRows(limit int, groups ...[]SearchRow) []SearchRow {
	best := make(map[string]SearchRow)

	for _, group := range groups {
		for _, row := range group {
			existing, ok := best[row.DocumentID]
			if !ok || row.Distance < existing.Distance {
				best[row.DocumentID] = row
			}
		}
	}

	merged := make([]SearchRow, 0, len(best))
	for _, row := range best {
		merged = append(merged, row)
	}

	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Distance == merged[j].Distance {
			return merged[i].Path < merged[j].Path
		}
		return merged[i].Distance < merged[j].Distance
	})

	if limit > 0 && len(merged) > limit {
		return merged[:limit]
	}

	return merged
}

func encodeSearchQuery(query []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, query); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readSearchRows(rows *sql.Rows) ([]SearchRow, error) {
	var results []SearchRow
	for rows.Next() {
		var row SearchRow
		if err := rows.Scan(&row.DocumentID, &row.Path, &row.Distance); err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	return results, rows.Err()
}
