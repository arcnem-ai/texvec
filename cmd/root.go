package cmd

import (
	"os"

	"github.com/arcnem-ai/texvec/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var verbose bool

func getCurrentEmbeddingModel(cmd *cobra.Command) string {
	if modelName, _ := cmd.Flags().GetString("model"); modelName != "" {
		return modelName
	}

	return viper.GetString("default_embedding_model")
}

func getCurrentSummaryModel(cmd *cobra.Command) string {
	if modelName, _ := cmd.Flags().GetString("summary-model"); modelName != "" {
		return modelName
	}

	return viper.GetString("default_summary_model")
}

var rootCmd = &cobra.Command{
	Use:   "texvec",
	Short: "Text similarity search using summaries and embeddings",
	Long: `texvec summarizes text documents, embeds the summaries and document chunks,
and stores everything in a local libsql database for cosine-distance search.

Supported document types: .txt, .md, .markdown
Supported embedding models: all-minilm-l6-v2, bge-small-en-v1.5
Supported summary models: flan-t5-small`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.InitConfig()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(setEmbeddingModelCmd)
	rootCmd.AddCommand(setSummaryModelCmd)
	rootCmd.AddCommand(summarizeCmd)
	rootCmd.AddCommand(embedCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(cleanCmd)
}
