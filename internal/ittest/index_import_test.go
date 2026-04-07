// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/ittest/testtemplate"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

var errInvalidIndexImportID = regexp.MustCompile(`Invalid import ID format`)

// TestImportIndex verifies that a milvus_index created outside Terraform can be
// imported using the <collection_name>/<index_name> format, that all attributes
// (index_type, metric_type) are populated from Milvus after import, and that
// a subsequent plan produces no diff.
func (s *ProviderTestSuite) TestImportIndex() {
	indexName := fmt.Sprintf("idx_%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	// Config that matches the externally-created collection and index.
	cfg := testtemplate.TerraformTemplate{
		Collections: []testtemplate.CollectionTemplate{
			baseCollectionForIndex(s.testCollectionName),
		},
		Indexes: []testtemplate.IndexTemplate{
			{
				TerraformResourceName: "test",
				CollectionName:        fmt.Sprintf("%q", s.testCollectionName),
				FieldName:             "embedding",
				IndexName:             indexName,
				IndexType:             "FLAT",
				MetricType:            "L2",
			},
		},
	}.Render()

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionAndIndexesDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Step 1: Create collection + index externally via Milvus client.
			{
				PreConfig: func() {
					createIndexExternally(s, s.testCollectionName, "embedding", indexName)
				},
				Config:             cfg,
				ResourceName:       "milvus_index.test",
				ImportState:        true,
				ImportStateId:      fmt.Sprintf("%s/%s", s.testCollectionName, indexName),
				ImportStatePersist: true,
				// index_params is null after import (FLAT has none) and null in
				// config too — no ignore needed.
				ImportStateVerifyIgnore: []string{"field_name"},
			},
			// Step 2: Plan with matching config — must be empty (no diff).
			{
				Config:   cfg,
				PlanOnly: true,
			},
			// Step 3: Apply and verify state is fully correct.
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists("milvus_index.test"),
					resource.TestCheckResourceAttr("milvus_index.test", "collection_name", s.testCollectionName),
					resource.TestCheckResourceAttr("milvus_index.test", "field_name", "embedding"),
					resource.TestCheckResourceAttr("milvus_index.test", "index_name", indexName),
					resource.TestCheckResourceAttr("milvus_index.test", "index_type", "FLAT"),
					resource.TestCheckResourceAttr("milvus_index.test", "metric_type", "L2"),
					testAccCheckIndexImportedState(s.testCollectionName, indexName),
				),
			},
		},
	})
}

// TestImportIndexInvalidID verifies that the provider returns a descriptive
// error when the import ID is not in the required <collection>/<index> format.
func (s *ProviderTestSuite) TestImportIndexInvalidID() {
	cfg := testtemplate.TerraformTemplate{
		Collections: []testtemplate.CollectionTemplate{
			baseCollectionForIndex(s.testCollectionName),
		},
		Indexes: []testtemplate.IndexTemplate{
			{
				TerraformResourceName: "test",
				CollectionName:        fmt.Sprintf("%q", s.testCollectionName),
				FieldName:             "embedding",
				IndexName:             "embedding_flat",
				IndexType:             "FLAT",
				MetricType:            "L2",
			},
		},
	}.Render()

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        cfg,
				ResourceName:  "milvus_index.test",
				ImportState:   true,
				ImportStateId: "missing-slash",
				ExpectError:   errInvalidIndexImportID,
			},
			{
				Config:        cfg,
				ResourceName:  "milvus_index.test",
				ImportState:   true,
				ImportStateId: "/no-collection",
				ExpectError:   errInvalidIndexImportID,
			},
			{
				Config:        cfg,
				ResourceName:  "milvus_index.test",
				ImportState:   true,
				ImportStateId: "no-index/",
				ExpectError:   errInvalidIndexImportID,
			},
		},
	})
}

// createIndexExternally creates a collection and a FLAT index on the embedding
// field directly via the Milvus client, bypassing Terraform entirely.
func createIndexExternally(s *ProviderTestSuite, collectionName, fieldName, indexName string) {
	s.T().Helper()

	client := provider.AccTestProviderConfig.Client
	s.Require().NotNil(client, "Milvus client not initialised — did PreCheck run?")

	// Create collection.
	schema := entity.NewSchema().
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName(fieldName).WithDataType(entity.FieldTypeFloatVector).WithDim(768))

	err := client.CreateCollection(
		context.Background(),
		milvusclient.NewCreateCollectionOption(collectionName, schema).
			WithConsistencyLevel(entity.ClStrong),
	)
	s.Require().NoError(err, "failed to create collection %q externally", collectionName)

	// Create FLAT index.
	task, err := client.CreateIndex(
		context.Background(),
		milvusclient.NewCreateIndexOption(collectionName, fieldName, index.NewFlatIndex(entity.L2)).
			WithIndexName(indexName),
	)
	s.Require().NoError(err, "failed to create index %q externally", indexName)
	s.Require().NoError(task.Await(context.Background()), "index creation task failed")

	s.T().Logf("Created collection %q and index %q externally", collectionName, indexName)
}

// testAccCheckIndexImportedState verifies that the key Milvus-sourced attributes
// are correctly populated in state after an index import.
func testAccCheckIndexImportedState(collectionName, indexName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var rs *terraform.ResourceState
		for _, r := range s.RootModule().Resources {
			if r.Type == "milvus_index" && r.Primary.Attributes["index_name"] == indexName {
				rs = r
				break
			}
		}
		if rs == nil {
			return fmt.Errorf("milvus_index with index_name=%q not found in state", indexName)
		}

		attrs := rs.Primary.Attributes

		if attrs["collection_name"] != collectionName {
			return fmt.Errorf("collection_name: expected %q, got %q", collectionName, attrs["collection_name"])
		}
		if attrs["index_type"] == "" {
			return fmt.Errorf("index_type: expected a non-empty value after import, got empty")
		}
		if attrs["metric_type"] == "" {
			return fmt.Errorf("metric_type: expected a non-empty value after import, got empty")
		}

		return nil
	}
}
