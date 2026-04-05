// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package collection

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/resources"
)

func TestAccResourceCollection_Basic(t *testing.T) {
	collectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { resources.testAccPreCheck(t) },
		ProtoV6ProviderFactories: resources.testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(collectionName),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceCollectionConfig_Basic(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "name", collectionName),
					resource.TestCheckResourceAttr("milvus_collection.test", "description", "Test collection"),
					resource.TestCheckResourceAttr("milvus_collection.test", "auto_id", "false"),
					// Verify computed attributes are populated
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
					resource.TestCheckResourceAttr("milvus_collection.test", "shard_num", "1"),
					resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Strong"),
					// Verify fields
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.name", "id"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.data_type", "Int64"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.is_primary_key", "true"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.name", "embedding"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.data_type", "FloatVector"),
					testAccCheckCollectionExists("milvus_collection.test"),
				),
			},
			// Import state testing — import by collection name, not the numeric id
			{
				ResourceName:      "milvus_collection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["milvus_collection.test"]
					if !ok {
						return "", fmt.Errorf("not found: milvus_collection.test")
					}
					return rs.Primary.Attributes["name"], nil
				},
			},
		},
	})
}

func TestAccResourceCollection_UpdateConsistencyLevel(t *testing.T) {
	collectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { resources.testAccPreCheck(t) },
		ProtoV6ProviderFactories: resources.testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(collectionName),
		Steps: []resource.TestStep{
			// Create with default consistency level
			{
				Config: testAccResourceCollectionConfig_Basic(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Strong"),
				),
			},
			// Update to Bounded (in-place update, no replace)
			{
				Config: testAccResourceCollectionConfig_WithConsistencyLevel(collectionName, "Bounded"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "name", collectionName),
					resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Bounded"),
					testAccCheckCollectionExists("milvus_collection.test"),
				),
			},
		},
	})
}

func TestAccResourceCollection_WithProperties(t *testing.T) {
	collectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { resources.testAccPreCheck(t) },
		ProtoV6ProviderFactories: resources.testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with properties
			{
				Config: testAccResourceCollectionConfig_WithProperties(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "name", collectionName),
					resource.TestCheckResourceAttr("milvus_collection.test", "enable_dynamic_field", "true"),
					resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Strong"),
				),
			},
		},
	})
}

func TestAccResourceCollection_MultipleFields(t *testing.T) {
	collectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { resources.testAccPreCheck(t) },
		ProtoV6ProviderFactories: resources.testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCollectionConfig_MultipleFields(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "name", collectionName),
					testAccCheckCollectionExists("milvus_collection.test"),
				),
			},
		},
	})
}

func testAccResourceCollectionConfig_Basic(name string) string {
	return fmt.Sprintf(`
provider "milvus" {
  address = "localhost:19530"
}

resource "milvus_collection" "test" {
  name        = "%s"
  description = "Test collection"

  fields = [
    {
      name           = "id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]
}
`, name)
}

func testAccResourceCollectionConfig_WithProperties(name string) string {
	return fmt.Sprintf(`
provider "milvus" {
  address = "localhost:19530"
}

resource "milvus_collection" "test" {
  name                   = "%s"
  description            = "Test collection with properties"
  enable_dynamic_field   = true
  shard_num              = 2
  consistency_level      = "Strong"

  fields = [
    {
      name           = "id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name       = "name"
      data_type  = "VarChar"
      max_length = 256
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]

  properties = {
    mmap_enabled            = true
    collection_ttl_seconds  = 3600
    allow_insert_auto_id    = true
  }
}
`, name)
}

func testAccResourceCollectionConfig_MultipleFields(name string) string {
	return fmt.Sprintf(`
provider "milvus" {
  address = "localhost:19530"
}

resource "milvus_collection" "test" {
  name                   = "%s"
  description            = "Test collection with multiple field types"
  enable_dynamic_field   = true

  fields = [
    {
      name           = "id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name       = "title"
      data_type  = "VarChar"
      max_length = 256
    },
    {
      name     = "is_active"
      data_type = "Bool"
    },
    {
      name     = "category"
      data_type = "Int32"
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]
}
`, name)
}

func testAccResourceCollectionConfig_WithConsistencyLevel(name, consistencyLevel string) string {
	return fmt.Sprintf(`
provider "milvus" {
  address = "localhost:19530"
}

resource "milvus_collection" "test" {
  name              = "%s"
  description       = "Test collection"
  consistency_level = "%s"

  fields = [
    {
      name           = "id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]
}
`, name, consistencyLevel)
}

// testAccCheckCollectionDestroyed verifies the collection is deleted after the test.
func testAccCheckCollectionDestroyed(collectionName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := resources.testAccProviderConfig.Client
		if client == nil {
			return fmt.Errorf("Provider not configured")
		}

		exists, err := client.HasCollection(context.Background(), milvusclient.NewHasCollectionOption(collectionName))
		if err != nil {
			return fmt.Errorf("Error checking collection %s: %v", collectionName, err)
		}
		if exists {
			return fmt.Errorf("Collection %s still exists after destroy", collectionName)
		}
		return nil
	}
}

func testAccCheckCollectionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No collection ID is set")
		}

		// Get the provider configured client
		client := resources.testAccProviderConfig.Client
		if client == nil {
			return fmt.Errorf("Provider not configured")
		}

		// Get the collection name from attributes
		collectionName, ok := rs.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Collection name not found in state")
		}

		// Check if collection exists using the collection name
		exists, err := client.HasCollection(context.Background(), milvusclient.NewHasCollectionOption(collectionName))
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Collection %s does not exist", collectionName)
		}

		return nil
	}
}
