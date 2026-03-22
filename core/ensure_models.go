package core

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arcnem-ai/texvec/config"
	"github.com/arcnem-ai/texvec/store"
)

func EnsureEmbeddingModel(db *sql.DB, modelName string) (EmbeddingModelInfo, error) {
	definition, ok := embeddingModelRegistry[modelName]
	if !ok {
		return EmbeddingModelInfo{}, fmt.Errorf("%s is not a valid embedding model", modelName)
	}

	dir, err := config.GetModelDir(modelName)
	if err != nil {
		return EmbeddingModelInfo{}, err
	}

	for _, asset := range definition.Assets {
		path := filepath.Join(dir, asset.FileName)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return EmbeddingModelInfo{}, err
		}

		if err := downloadAsset(modelName, path, asset.DownloadURL); err != nil {
			return EmbeddingModelInfo{}, err
		}
	}

	if _, err := store.InsertEmbeddingModel(db, modelName, definition.EmbeddingDim); err != nil {
		return EmbeddingModelInfo{}, err
	}

	return EmbeddingModelInfo{
		ID:            modelName,
		ModelPath:     filepath.Join(dir, definition.ModelFile),
		TokenizerPath: filepath.Join(dir, definition.TokenizerFile),
		Definition:    definition,
	}, nil
}

func EnsureSummaryModel(db *sql.DB, modelName string) (SummaryModelInfo, error) {
	definition, ok := summaryModelRegistry[modelName]
	if !ok {
		return SummaryModelInfo{}, fmt.Errorf("%s is not a valid summary model", modelName)
	}

	dir, err := config.GetModelDir(modelName)
	if err != nil {
		return SummaryModelInfo{}, err
	}

	for _, asset := range definition.Assets {
		path := filepath.Join(dir, asset.FileName)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return SummaryModelInfo{}, err
		}

		if err := downloadAsset(modelName, path, asset.DownloadURL); err != nil {
			return SummaryModelInfo{}, err
		}
	}

	if _, err := store.InsertSummaryModel(db, modelName); err != nil {
		return SummaryModelInfo{}, err
	}

	return SummaryModelInfo{
		ID:            modelName,
		EncoderPath:   filepath.Join(dir, definition.EncoderFile),
		DecoderPath:   filepath.Join(dir, definition.DecoderFile),
		TokenizerPath: filepath.Join(dir, definition.TokenizerFile),
		Definition:    definition,
	}, nil
}
