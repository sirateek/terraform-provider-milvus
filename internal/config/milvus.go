package config

import "github.com/kelseyhightower/envconfig"

type Milvus struct {
	Address  string `envconfig:"ADDRESS"`
	Username string `envconfig:"USERNAME"`
	Password string `envconfig:"PASSWORD"`
}

func ProvideMilvusConfig() (Milvus, error) {
	var cfg Milvus
	err := envconfig.Process("MILVUS", &cfg)
	return cfg, err
}
