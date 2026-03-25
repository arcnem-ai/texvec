package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/arcnem-ai/texvec/core"
	"github.com/arcnem-ai/texvec/store"
	"github.com/spf13/cobra"
)

var searchLimit int
var searchChunkLimit int
var searchText string

const searchChunkPreviewRunes = 140

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
		if searchChunkLimit < 1 {
			return fmt.Errorf("--chunks must be at least 1")
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

			results, err := store.SearchSummaryEmbeddings(db, summaryModelName, embeddingModelName, queryEmbedding, searchLimit, excludePath)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}

			evidence, err := store.SearchDocumentChunksForDocuments(
				db,
				embeddingModelName,
				queryEmbedding,
				searchResultDocumentIDs(results),
				searchChunkLimit,
			)
			if err != nil {
				return err
			}

			for i := range results {
				results[i].Chunks = evidence[results[i].DocumentID]
			}

			for _, result := range results {
				fmt.Println(formatSearchResult(result))
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
	searchCmd.Flags().IntVarP(&searchChunkLimit, "chunks", "c", 1, "Number of supporting chunks to show per result")
	searchCmd.Flags().StringP("model", "m", "", "Embedding model to use")
	searchCmd.Flags().String("summary-model", "", "Summary model to use")
	searchCmd.Flags().StringVar(&searchText, "text", "", "Search using raw text instead of a file")
}

func searchResultDocumentIDs(results []store.SearchResult) []string {
	ids := make([]string, 0, len(results))
	for _, result := range results {
		ids = append(ids, result.DocumentID)
	}

	return ids
}

func formatSearchResult(result store.SearchResult) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%.4f  %s", result.Distance, result.Path))

	for _, chunk := range result.Chunks {
		builder.WriteString(
			fmt.Sprintf(
				"\n        %.4f  %s: %s",
				chunk.Distance,
				formatLineRange(chunk.StartLine, chunk.EndLine),
				truncateRunes(chunk.Content, searchChunkPreviewRunes),
			),
		)
	}

	return builder.String()
}

func formatLineRange(startLine, endLine int) string {
	if startLine <= 0 || endLine <= 0 {
		return "match"
	}
	if startLine == endLine {
		return fmt.Sprintf("line %d", startLine)
	}

	return fmt.Sprintf("lines %d-%d", startLine, endLine)
}

func truncateRunes(text string, limit int) string {
	if limit <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	if limit <= 3 {
		return string(runes[:limit])
	}

	return string(runes[:limit-3]) + "..."
}
