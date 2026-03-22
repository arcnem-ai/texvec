package cmd

import (
	"database/sql"
	"fmt"

	"github.com/arcnem-ai/texvec/core"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize [document]",
	Short: "Generate and print a summary for a document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		document, err := core.LoadDocument(args[0])
		if err != nil {
			return err
		}

		return withDB(func(db *sql.DB) error {
			summaryInfo, err := ensureSummaryModel(db, getCurrentSummaryModel(cmd))
			if err != nil {
				return err
			}

			summarizer, err := core.NewSummarizer(summaryInfo, runtimeOptions())
			if err != nil {
				return err
			}
			defer summarizer.Close()

			summary, err := summarizer.Summarize(document.NormalizedText)
			if err != nil {
				return err
			}

			fmt.Println(summary)
			return nil
		})
	},
}

func init() {
	summarizeCmd.Flags().String("summary-model", "", "Summary model to use")
}
