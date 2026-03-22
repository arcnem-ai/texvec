package cmd

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/arcnem-ai/texvec/config"
	"github.com/arcnem-ai/texvec/core"
	"github.com/arcnem-ai/texvec/store"
	"github.com/spf13/viper"
)

func setupCommandTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("TEXVEC_HOME", filepath.Join(dir, ".texvec"))
	viper.Reset()
	if err := config.InitConfig(); err != nil {
		t.Fatalf("init config: %v", err)
	}

	db, err := store.InitDB()
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		viper.Reset()
	})

	return db
}

func TestSetEmbeddingModelCommand(t *testing.T) {
	db := setupCommandTestDB(t)

	originalInitDB := initDB
	originalEnsureEmbedding := ensureEmbeddingModel
	initDB = func() (*sql.DB, error) { return db, nil }
	ensureEmbeddingModel = func(db *sql.DB, modelName string) (core.EmbeddingModelInfo, error) {
		return core.EmbeddingModelInfo{ID: modelName}, nil
	}
	defer func() {
		initDB = originalInitDB
		ensureEmbeddingModel = originalEnsureEmbedding
	}()

	if err := setEmbeddingModelCmd.RunE(setEmbeddingModelCmd, []string{"bge-small-en-v1.5"}); err != nil {
		t.Fatalf("run command: %v", err)
	}

	if got := viper.GetString("default_embedding_model"); got != "bge-small-en-v1.5" {
		t.Fatalf("expected updated embedding model, got %s", got)
	}
}

func TestSetSummaryModelCommand(t *testing.T) {
	db := setupCommandTestDB(t)

	originalInitDB := initDB
	originalEnsureSummary := ensureSummaryModel
	initDB = func() (*sql.DB, error) { return db, nil }
	ensureSummaryModel = func(db *sql.DB, modelName string) (core.SummaryModelInfo, error) {
		return core.SummaryModelInfo{ID: modelName}, nil
	}
	defer func() {
		initDB = originalInitDB
		ensureSummaryModel = originalEnsureSummary
	}()

	if err := setSummaryModelCmd.RunE(setSummaryModelCmd, []string{"flan-t5-small"}); err != nil {
		t.Fatalf("run command: %v", err)
	}

	if got := viper.GetString("default_summary_model"); got != "flan-t5-small" {
		t.Fatalf("expected updated summary model, got %s", got)
	}
}
