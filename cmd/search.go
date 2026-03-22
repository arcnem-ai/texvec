package cmd

import (
	"database/sql"
	"fmt"

	"github.com/arcnem-ai/texvec/core"
	"github.com/arcnem-ai/texvec/store"
	"github.com/spf13/cobra"
)

var searchLimit int
var searchText string

var searchCmd = &cobra.Command{
	Use:   "search [document]",
	Short: "Find similar documents",
	Args: func(cmd *cobra.Command, args []string) error {
		if searchText != "" && len(args) > 0 {
			return fmt.Errorf("use either a document path or --text, not both")
		}
		if searchText == "" && len(args) != 1 {
			return fmt.Errorf("expected a document path or --text")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB) error {
			embeddingModelName := getCurrentEmbeddingModel(cmd)
			summaryModelName := getCurrentSummaryModel(cmd)

			embeddingInfo, err := ensureEmbeddingModel(db, embeddingModelName)
			if err != nil {
				return err
			}

			var queryEmbedding []float32
			excludePath := ""

			if searchText == "" {
				excludePath = args[0]
				queryEmbedding, err = store.GetSummaryEmbeddingByPath(db, args[0], summaryModelName, embeddingModelName)
				if err != nil {
					return err
				}
			}

			if queryEmbedding == nil {
				queryText, err := resolveQueryText(args)
				if err != nil {
					return err
				}

				if core.ShouldSummarizeQuery(queryText) {
					summaryInfo, err := ensureSummaryModel(db, summaryModelName)
					if err != nil {
						return err
					}

					summarizer, err := core.NewSummarizer(summaryInfo, runtimeOptions())
					if err != nil {
						return err
					}
					queryText, err = summarizer.Summarize(queryText)
					summarizer.Close()
					if err != nil {
						return err
					}
				}

				embedder, err := core.NewTextEmbedder(embeddingInfo, runtimeOptions())
				if err != nil {
					return err
				}
				defer embedder.Close()

				queryEmbedding, err = embedder.EmbedQuery(queryText)
				if err != nil {
					return err
				}
			}

			summaryRows, err := store.SearchSummaryEmbeddings(db, summaryModelName, embeddingModelName, queryEmbedding, searchLimit, excludePath)
			if err != nil {
				return err
			}
			chunkRows, err := store.SearchDocumentChunks(db, embeddingModelName, queryEmbedding, searchLimit, excludePath)
			if err != nil {
				return err
			}

			results := store.MergeSearchRows(searchLimit, summaryRows, chunkRows)
			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}

			for _, result := range results {
				fmt.Printf("%.4f  %s\n", result.Distance, result.Path)
			}
			return nil
		})
	},
}

func resolveQueryText(args []string) (string, error) {
	if searchText != "" {
		return core.NormalizeText(searchText), nil
	}

	document, err := core.LoadDocument(args[0])
	if err != nil {
		return "", err
	}

	return document.NormalizedText, nil
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "k", 5, "Number of results to return")
	searchCmd.Flags().StringP("model", "m", "", "Embedding model to use")
	searchCmd.Flags().String("summary-model", "", "Summary model to use")
	searchCmd.Flags().StringVar(&searchText, "text", "", "Search using raw text instead of a file")
}
