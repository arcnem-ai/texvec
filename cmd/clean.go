package cmd

import (
	"fmt"
	"os"

	"github.com/arcnem-ai/texvec/config"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all texvec data",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := config.BaseDir()
		if err != nil {
			return err
		}

		if err := os.RemoveAll(dir); err != nil {
			return err
		}

		fmt.Printf("Removed %s\n", dir)
		return nil
	},
}
