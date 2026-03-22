package config

import (
	"os"

	"github.com/spf13/viper"
)

func InitConfig() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	viper.SetConfigFile(configPath)
	viper.SetDefault("default_embedding_model", "all-minilm-l6-v2")
	viper.SetDefault("default_summary_model", "flan-t5-small")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := viper.SafeWriteConfigAs(configPath); err != nil {
			return err
		}
	}

	return viper.ReadInConfig()
}
