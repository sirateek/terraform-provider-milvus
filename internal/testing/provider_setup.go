//
// SPDX-License-Identifier: MPL-2.0

package testing

import (
	"context"
	"os"
	"testing"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/client/milvus"
	"github.com/sirateek/terraform-provider-milvus/internal/config"
)

// ProviderTestConfig holds the configured client for use in acceptance tests.
type ProviderTestConfig struct {
	Client *milvusclient.Client
}

// PreCheck verifies that the provider can be configured and a Milvus
// connection can be established before running acceptance tests.
func PreCheck(t *testing.T) *ProviderTestConfig {
	// Read configuration from environment variables
	address := os.Getenv("MILVUS_ADDRESS")
	if address == "" {
		address = "localhost:19530"
	}

	username := os.Getenv("MILVUS_USERNAME")
	password := os.Getenv("MILVUS_PASSWORD")
	dbName := os.Getenv("MILVUS_DB_NAME")
	if dbName == "" {
		dbName = "default"
	}

	enableTLS := os.Getenv("MILVUS_ENABLE_TLS") == "true"

	// Create Milvus client config with pointers
	milvusConfig := config.Milvus{
		Address: &address,
		Username: func() *string {
			if username == "" {
				return nil
			}
			return &username
		}(),
		Password: func() *string {
			if password == "" {
				return nil
			}
			return &password
		}(),
		DBName:    &dbName,
		EnableTLS: &enableTLS,
	}

	client, diag := milvus.ProvideMilvusClient(milvusConfig)
	if diag != nil {
		t.Fatalf("failed to create Milvus client: %s", diag.Summary())
	}

	// Test the connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Try a simple health check using DescribeCollection on a non-existent collection
	_, connErr := client.DescribeCollection(ctx, milvusclient.NewDescribeCollectionOption("_test_connection"))

	// We expect this to fail (collection doesn't exist), but it shows the connection works
	if connErr != nil {
		if !isCollectionNotFoundError(connErr.Error()) {
			// It's some other error, possibly connection-related
			t.Fatalf("failed to connect to Milvus at %s: %s", address, connErr)
		}
	}

	return &ProviderTestConfig{
		Client: client,
	}
}

// isCollectionNotFoundError checks if the error indicates a collection not found.
func isCollectionNotFoundError(errStr string) bool {
	return containsSubstring(errStr, "not found") ||
		containsSubstring(errStr, "does not exist") ||
		containsSubstring(errStr, "CollectionNotExist") ||
		containsSubstring(errStr, "can't find collection")
}

// containsSubstring checks if a string contains a substring.
func containsSubstring(str, substr string) bool {
	if len(str) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
