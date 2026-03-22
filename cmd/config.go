package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, key := range viper.AllKeys() {
			fmt.Printf("%s: %v\n", key, viper.Get(key))
		}
		return nil
	},
}
