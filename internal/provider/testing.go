//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/milvus-io/milvus/pkg/v2/util/merr"
	"github.com/sirateek/terraform-provider-milvus/internal/client/milvus"
	"github.com/sirateek/terraform-provider-milvus/internal/config"
	testingpkg "github.com/sirateek/terraform-provider-milvus/internal/testing"
)

// ProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var ProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

func init() {
	ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"milvus": providerserver.NewProtocol6WithError(New("test")()),
	}

	// Register the provider factory with the testing package to make it available
	// to test files without creating circular imports.
	// We use an indirect import through a function call to avoid the circular dependency.
	registerProviderFactories()
}

// registerProviderFactories registers this provider's factories with the testing package.
// This allows test files to import the testing package (not the provider package)
// and still have access to the provider factories, avoiding circular imports.
func registerProviderFactories() {
	testingpkg.RegisterProviderFactory("milvus", ProtoV6ProviderFactories["milvus"])
}

// ProviderConfig holds the configured client for use in acceptance tests.
type ProviderConfig struct {
	Client *milvusclient.Client
}

// AccTestProviderConfig is a provider configuration object used in acceptance testing.
// It holds a reference to the configured Milvus client for verification purposes.
var AccTestProviderConfig *ProviderConfig

// PreCheck verifies that the provider can be configured and a Milvus
// connection can be established before running acceptance tests.
func PreCheck(t *testing.T) {
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
	// This will fail but shows the connection is alive
	_, connErr := client.DescribeCollection(ctx, milvusclient.NewDescribeCollectionOption("_test_connection"))

	// We expect this to fail (collection doesn't exist), but it shows the connection works
	// If it fails with connection error, we'll catch it

	if connErr != nil {
		// Check if it's a "collection not found" error, which is expected
		if !errors.Is(connErr, merr.ErrCollectionNotFound) {
			// It's some other error, possibly connection-related
			t.Fatalf("failed to connect to Milvus at %s: %s", address, connErr)
		}
	}

	// Store the client in the provider config for use in tests
	AccTestProviderConfig = &ProviderConfig{
		Client: client,
	}
}
