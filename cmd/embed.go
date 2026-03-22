package cmd

import (
	"database/sql"
	"fmt"

	"github.com/arcnem-ai/texvec/core"
	"github.com/arcnem-ai/texvec/store"
	"github.com/spf13/cobra"
)

var embedCmd = &cobra.Command{
	Use:   "embed [document]",
	Short: "Summarize, chunk, and embed a document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		document, err := core.LoadDocument(args[0])
		if err != nil {
			return err
		}

		return withDB(func(db *sql.DB) error {
			embeddingModelName := getCurrentEmbeddingModel(cmd)
			summaryModelName := getCurrentSummaryModel(cmd)

			embeddingInfo, err := ensureEmbeddingModel(db, embeddingModelName)
			if err != nil {
				return err
			}
			summaryInfo, err := ensureSummaryModel(db, summaryModelName)
			if err != nil {
				return err
			}

			documentRecord, hashChanged, err := store.UpsertDocument(db, document.Path, document.ContentHash)
			if err != nil {
				return err
			}

			summaryRecord, err := store.GetDocumentSummary(db, documentRecord.ID, summaryModelName)
			if err != nil {
				return err
			}

			needsSummary := hashChanged || summaryRecord == nil

			var summarizer *core.Summarizer
			if needsSummary {
				summarizer, err = core.NewSummarizer(summaryInfo, runtimeOptions())
				if err != nil {
					return err
				}
				defer summarizer.Close()

				summaryText, err := summarizer.Summarize(document.NormalizedText)
				if err != nil {
					return err
				}

				record, err := store.UpsertDocumentSummary(db, documentRecord.ID, summaryModelName, summaryText)
				if err != nil {
					return err
				}
				summaryRecord = &record
			}

			if summaryRecord == nil {
				return fmt.Errorf("failed to resolve summary row for %s", document.Path)
			}

			summaryEmbedding, err := store.GetDocumentSummaryEmbedding(db, summaryRecord.ID, embeddingModelName)
			if err != nil {
				return err
			}
			hasChunks, err := store.HasDocumentChunks(db, documentRecord.ID, embeddingModelName)
			if err != nil {
				return err
			}

			needsSummaryEmbedding := hashChanged || summaryEmbedding == nil
			needsChunks := hashChanged || !hasChunks

			if needsSummaryEmbedding || needsChunks {
				embedder, err := core.NewTextEmbedder(embeddingInfo, runtimeOptions())
				if err != nil {
					return err
				}
				defer embedder.Close()

				if needsSummaryEmbedding {
					embedding, err := embedder.EmbedDocument(summaryRecord.Content)
					if err != nil {
						return err
					}
					if err := store.UpsertDocumentSummaryEmbedding(db, summaryRecord.ID, embeddingModelName, embedding); err != nil {
						return err
					}
				}

				if needsChunks {
					chunks, err := chunkInputsFromDocument(document, embedder)
					if err != nil {
						return err
					}
					if err := store.ReplaceDocumentChunks(db, documentRecord.ID, embeddingModelName, chunks); err != nil {
						return err
					}
				}
			}

			fmt.Printf("Indexed document %s\n", document.Path)
			return nil
		})
	},
}

func init() {
	embedCmd.Flags().StringP("model", "m", "", "Embedding model to use")
	embedCmd.Flags().String("summary-model", "", "Summary model to use")
}
