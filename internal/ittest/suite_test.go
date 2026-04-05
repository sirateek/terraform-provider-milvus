// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
	"github.com/stretchr/testify/suite"
)

type ProviderTestSuite struct {
	suite.Suite

	testCollectionName string
}

func TestProviderTestSuite(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	address := os.Getenv("MILVUS_ADDRESS")
	t.Logf("Address: %s", address)
	suite.Run(t, new(ProviderTestSuite))
}

func (s *ProviderTestSuite) SetupTest() {
	s.testCollectionName = fmt.Sprintf(
		"tf_test_%s",
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
	)

}

// testAccCheckCollectionAndIndexesDestroyed verifies that after destroy both the
// collection and every index that was attached to it are gone from Milvus.
func testAccCheckCollectionAndIndexesDestroyed(collectionName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		ctx := context.Background()

		// Confirm the collection itself is gone.
		exists, err := client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collectionName))
		if err != nil {
			return fmt.Errorf("error checking collection %s: %v", collectionName, err)
		}
		if exists {
			return fmt.Errorf("collection %s still exists after destroy", collectionName)
		}

		// Confirm every milvus_index resource in the prior state no longer exists.
		// If the collection is already gone Milvus cascades the index deletion, but
		// we still verify through the state so the test is explicit about every index.
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "milvus_index" {
				continue
			}

			rsCollectionName := rs.Primary.Attributes["collection_name"]
			if rsCollectionName != collectionName {
				continue
			}

			indexName := rs.Primary.Attributes["index_name"]
			if indexName == "" {
				indexName = rs.Primary.Attributes["field_name"]
			}

			_, describeErr := client.DescribeIndex(
				ctx,
				milvusclient.NewDescribeIndexOption(rsCollectionName, indexName),
			)
			if describeErr == nil {
				return fmt.Errorf("index %s (resource %s) still exists after collection %s was destroyed",
					indexName, name, collectionName)
			}
			// Any error from DescribeIndex means the index is gone — acceptable.
		}

		return nil
	}
}

func testAccCheckCollectionDestroyed(collectionName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		exists, err := client.HasCollection(
			context.Background(),
			milvusclient.NewHasCollectionOption(collectionName),
		)
		if err != nil {
			return fmt.Errorf("error checking collection %s: %v", collectionName, err)
		}
		if exists {
			return fmt.Errorf("collection %s still exists after destroy", collectionName)
		}
		return nil
	}
}

func testAccCheckCollectionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no collection ID is set")
		}

		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		collectionName, ok := rs.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("collection name not found in state")
		}

		exists, err := client.HasCollection(
			context.Background(),
			milvusclient.NewHasCollectionOption(collectionName),
		)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("collection %s does not exist", collectionName)
		}
		return nil
	}
}
