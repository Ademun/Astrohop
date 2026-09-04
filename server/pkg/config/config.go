package config

import (
	"github.com/kelseyhightower/envconfig"
)

type InfraCfg struct {
	DBConnectionString string `envconfig:"DB_CONNECTION_STRING"`
}

type Config struct {
	Infra InfraCfg
}

func GetConfig() *Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		panic(err)
	}
	return &cfg
}
