package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setSummaryModelCmd = &cobra.Command{
	Use:   "set-summary-model [model-name]",
	Short: "Set default summary model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]

		return withDB(func(db *sql.DB) error {
			if _, err := ensureSummaryModel(db, modelName); err != nil {
				return err
			}

			viper.Set("default_summary_model", modelName)
			if err := viper.WriteConfig(); err != nil {
				return err
			}

			fmt.Printf("Default summary model set to %s\n", modelName)
			return nil
		})
	},
}
