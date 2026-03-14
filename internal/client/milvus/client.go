package milvus

import (
	"context"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/config"
)

func ProvideMilvusClient(config config.Milvus) (*milvusclient.Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address:  config.Address,
		Username: config.Username,
		Password: config.Password,
	})
}
