package store

import "database/sql"

type DocumentRecord struct {
	ID     string
	Path   string
	Models []string
}

func ListDocuments(db *sql.DB, modelFilter string, limit int) ([]DocumentRecord, error) {
	query := `
		SELECT d.id, d.path, indexed_rows.embedding_model_id
		FROM documents d
		JOIN (
			SELECT ds.document_id, dse.embedding_model_id
			FROM document_summaries ds
			JOIN document_summary_embeddings dse ON dse.summary_id = ds.id
			UNION
			SELECT document_id, embedding_model_id
			FROM document_embedding_chunks
		) indexed_rows ON indexed_rows.document_id = d.id
	`
	var args []any

	if modelFilter != "" {
		query += ` WHERE indexed_rows.embedding_model_id = ?`
		args = append(args, modelFilter)
	}

	query += ` ORDER BY d.path, indexed_rows.embedding_model_id`
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documentMap := map[string]*DocumentRecord{}
	var order []string

	for rows.Next() {
		var id, path, model string
		if err := rows.Scan(&id, &path, &model); err != nil {
			return nil, err
		}

		if record, ok := documentMap[id]; ok {
			record.Models = append(record.Models, model)
			continue
		}

		documentMap[id] = &DocumentRecord{
			ID:     id,
			Path:   path,
			Models: []string{model},
		}
		order = append(order, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	records := make([]DocumentRecord, len(order))
	for i, id := range order {
		records[i] = *documentMap[id]
	}

	return records, nil
}
