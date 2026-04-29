package config

import (
	"github.com/spf13/viper"
)

// SecurityConfig holds configuration for security adapters.
type SecurityConfig struct {
}

func defaultSecurityConfig(key string, vConfig *viper.Viper) {
}
