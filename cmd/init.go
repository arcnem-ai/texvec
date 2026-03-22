package cmd

import (
	"fmt"

	"github.com/arcnem-ai/texvec/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize texvec environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := core.EnsureRuntime(); err != nil {
			return err
		}

		db, err := initDB()
		if err != nil {
			return err
		}
		defer db.Close()

		embeddingModel := viper.GetString("default_embedding_model")
		if _, err := ensureEmbeddingModel(db, embeddingModel); err != nil {
			return err
		}

		summaryModel := viper.GetString("default_summary_model")
		if _, err := ensureSummaryModel(db, summaryModel); err != nil {
			return err
		}

		fmt.Println("texvec initialized successfully.")
		return nil
	},
}
