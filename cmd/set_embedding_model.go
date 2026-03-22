package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setEmbeddingModelCmd = &cobra.Command{
	Use:   "set-embedding-model [model-name]",
	Short: "Set default embedding model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]

		return withDB(func(db *sql.DB) error {
			if _, err := ensureEmbeddingModel(db, modelName); err != nil {
				return err
			}

			viper.Set("default_embedding_model", modelName)
			if err := viper.WriteConfig(); err != nil {
				return err
			}

			fmt.Printf("Default embedding model set to %s\n", modelName)
			return nil
		})
	},
}
