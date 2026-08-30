package config

import (
	"fmt"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type InfraCfg struct {
	DBConnectionString string `envconfig:"DB_CONNECTION_STRING"`
	MinIOEndpoint      string `envconfig:"MINIO_ENDPOINT"`
	MinIORootUser      string `envconfig:"MINIO_ROOT_USER"`
	MinIORootPassword  string `envconfig:"MINIO_ROOT_PASSWORD"`
}

type Config struct {
	Infra InfraCfg
}

var load = sync.OnceValues(func() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
})

func C() *Config {
	cfg, err := load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}
