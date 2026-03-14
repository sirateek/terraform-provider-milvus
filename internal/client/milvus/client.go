package milvus

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/config"
)

func ProvideMilvusClient(config config.Milvus) (*milvusclient.Client, diag.Diagnostic) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address:       config.Address,
		Username:      config.Username,
		Password:      config.Password,
		DBName:        config.DBName,
		EnableTLSAuth: config.EnableTLS,
		APIKey:        config.APIKey,
		ServerVersion: config.ServerVersion,
	})
	if err != nil {
		return nil, diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Failed to create milvus client",
			err.Error(),
		)
	}
	return client, nil
}
