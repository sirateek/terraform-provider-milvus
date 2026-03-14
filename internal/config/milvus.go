package config

import "github.com/kelseyhightower/envconfig"

type Milvus struct {
	Address  string `envconfig:"ADDRESS"`
	Username string `envconfig:"USERNAME"`
	Password string `envconfig:"PASSWORD"`
	APIKey   string `envconfig:"API_KEY"`
	DBName   string `envconfig:"DB_NAME"`

	EnableTLS bool `envconfig:"ENABLE_TLS"`

	ServerVersion string `envconfig:"SERVER_VERSION"`
}

func ProvideMilvusConfig() (Milvus, error) {
	var cfg Milvus
	err := envconfig.Process("MILVUS", &cfg)
	return cfg, err
}
