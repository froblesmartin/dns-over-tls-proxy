package config

import (
	"strings"

	"github.com/spf13/viper"
)

func InitializeConfig() {
	viper.SetDefault("DoTPort", "853")
	viper.SetDefault("DoTServer", "one.one.one.one")
	viper.SetDefault("ListeningPort", "53")

	viper.SetEnvPrefix("DOT_PROXY")
	viper.EnvKeyReplacer(strings.NewReplacer("", "_"))
	viper.AutomaticEnv()
}
