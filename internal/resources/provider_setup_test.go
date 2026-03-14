// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	testingpkg "github.com/sirateek/terraform-provider-milvus/internal/testing"
)

// testAccProviderConfig is a provider configuration object used in acceptance testing.
// This holds a reference to the configured Milvus client for verification purposes.
var testAccProviderConfig *testingpkg.ProviderTestConfig

// testAccProtoV6ProviderFactories will be populated with provider factories at test runtime
var testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

func init() {
	// Initialize with reference to testing package's factories
	// These will be populated by provider/testing.go's init() function
	testAccProtoV6ProviderFactories = testingpkg.ProtoV6ProviderFactories
}

// testAccPreCheck verifies that the provider can be configured and a Milvus
// connection can be established before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	// Use the testing helper to check prerequisites and get the Milvus client
	testAccProviderConfig = testingpkg.PreCheck(t)

	// Verify the factories are populated (provider must be imported somewhere)
	if len(testAccProtoV6ProviderFactories) == 0 {
		t.Fatalf("Provider factories not registered. Make sure provider package is initialized.")
	}
}
