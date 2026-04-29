package config

import "github.com/spf13/viper"

type ObserverConfig struct {
	Group      bool `mapstructure:"group"`
	Role       bool `mapstructure:"role"`
	Permission bool `mapstructure:"permission"`
	Subject    bool `mapstructure:"subject"`
	Access     bool `mapstructure:"access"`
}

func defaultObserverConfig(key string, vConfig *viper.Viper) {
	vConfig.SetDefault(key+".group", true)
	vConfig.SetDefault(key+".role", true)
	vConfig.SetDefault(key+".permission", true)
	vConfig.SetDefault(key+".subject", true)
	vConfig.SetDefault(key+".access", true)
}
