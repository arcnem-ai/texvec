package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/arcnem-ai/texvec/store"
	"github.com/spf13/cobra"
)

var listLimit int

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List indexed documents",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB) error {
			modelName, _ := cmd.Flags().GetString("model")
			documents, err := store.ListDocuments(db, modelName, listLimit)
			if err != nil {
				return err
			}

			if len(documents) == 0 {
				fmt.Println("No documents found.")
				return nil
			}

			for _, document := range documents {
				fmt.Printf("%s  [%s]\n", document.Path, strings.Join(document.Models, ", "))
			}
			return nil
		})
	},
}

func init() {
	listCmd.Flags().IntVarP(&listLimit, "limit", "k", 0, "Max number of documents to show (0 = all)")
	listCmd.Flags().StringP("model", "m", "", "Filter by embedding model")
}
