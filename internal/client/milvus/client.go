// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package milvus

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/config"
)

func ProvideMilvusClient(config config.Milvus) (*milvusclient.Client, diag.Diagnostic) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Extract string values from pointers, using empty string as default
	address := ""
	if config.Address != nil {
		address = *config.Address
	}
	username := ""
	if config.Username != nil {
		username = *config.Username
	}
	password := ""
	if config.Password != nil {
		password = *config.Password
	}
	dbName := ""
	if config.DBName != nil {
		dbName = *config.DBName
	}
	apiKey := ""
	if config.APIKey != nil {
		apiKey = *config.APIKey
	}
	enableTLS := false
	if config.EnableTLS != nil {
		enableTLS = *config.EnableTLS
	}
	serverVersion := ""
	if config.ServerVersion != nil {
		serverVersion = *config.ServerVersion
	}
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address:       address,
		Username:      username,
		Password:      password,
		DBName:        dbName,
		EnableTLSAuth: enableTLS,
		APIKey:        apiKey,
		ServerVersion: serverVersion,
	})
	if err != nil {
		return nil, diag.NewAttributeErrorDiagnostic(
			path.Root("milvus_client"),
			"Failed to create milvus client",
			err.Error(),
		)
	}
	return client, nil
}
