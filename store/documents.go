package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID          string
	Path        string
	ContentHash string
}

type SummaryRecord struct {
	ID             string
	DocumentID     string
	SummaryModelID string
	Content        string
}

func UpsertDocument(db *sql.DB, path, contentHash string) (Document, bool, error) {
	var document Document
	err := db.QueryRow(
		`SELECT id, path, content_hash FROM documents WHERE path = ?`,
		path,
	).Scan(&document.ID, &document.Path, &document.ContentHash)

	switch err {
	case nil:
		if document.ContentHash == contentHash {
			return document, false, nil
		}

		document.ContentHash = contentHash
		_, err = db.Exec(
			`UPDATE documents SET content_hash = ?, updated_at = ? WHERE id = ?`,
			contentHash,
			time.Now().Unix(),
			document.ID,
		)
		return document, true, err
	case sql.ErrNoRows:
		document = Document{
			ID:          uuid.NewString(),
			Path:        path,
			ContentHash: contentHash,
		}
		now := time.Now().Unix()
		_, err = db.Exec(
			`INSERT INTO documents (id, path, content_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			document.ID,
			document.Path,
			document.ContentHash,
			now,
			now,
		)
		return document, true, err
	default:
		return Document{}, false, err
	}
}

func GetDocumentByPath(db *sql.DB, path string) (*Document, error) {
	var document Document
	err := db.QueryRow(
		`SELECT id, path, content_hash FROM documents WHERE path = ?`,
		path,
	).Scan(&document.ID, &document.Path, &document.ContentHash)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &document, nil
}

func GetDocumentSummary(db *sql.DB, documentID, summaryModelID string) (*SummaryRecord, error) {
	var summary SummaryRecord
	err := db.QueryRow(
		`SELECT id, document_id, summary_model_id, content
		FROM document_summaries
		WHERE document_id = ? AND summary_model_id = ?`,
		documentID,
		summaryModelID,
	).Scan(&summary.ID, &summary.DocumentID, &summary.SummaryModelID, &summary.Content)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func UpsertDocumentSummary(db *sql.DB, documentID, summaryModelID, content string) (SummaryRecord, error) {
	existing, err := GetDocumentSummary(db, documentID, summaryModelID)
	if err != nil {
		return SummaryRecord{}, err
	}

	now := time.Now().Unix()
	if existing == nil {
		record := SummaryRecord{
			ID:             uuid.NewString(),
			DocumentID:     documentID,
			SummaryModelID: summaryModelID,
			Content:        content,
		}
		_, err := db.Exec(
			`INSERT INTO document_summaries
			(id, document_id, summary_model_id, content, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			record.ID,
			record.DocumentID,
			record.SummaryModelID,
			record.Content,
			now,
			now,
		)
		return record, err
	}

	if existing.Content != content {
		_, err := db.Exec(
			`UPDATE document_summaries SET content = ?, updated_at = ? WHERE id = ?`,
			content,
			now,
			existing.ID,
		)
		if err != nil {
			return SummaryRecord{}, err
		}
		existing.Content = content
	}

	return *existing, nil
}
