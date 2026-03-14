// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func TestAccResourceIndex_VectorFlat(t *testing.T) {
	collectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexConfig_VectorFlat(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_index.test", "collection_name", collectionName),
					resource.TestCheckResourceAttr("milvus_index.test", "field_name", "embedding"),
					resource.TestCheckResourceAttr("milvus_index.test", "index_type", "FLAT"),
					resource.TestCheckResourceAttr("milvus_index.test", "metric_type", "L2"),
					testAccCheckIndexExists("milvus_index.test"),
				),
			},
		},
	})
}

func TestAccResourceIndex_VectorHnsw(t *testing.T) {
	collectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexConfig_VectorHnsw(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_index.test", "index_type", "HNSW"),
					resource.TestCheckResourceAttr("milvus_index.test", "metric_type", "COSINE"),
					testAccCheckIndexExists("milvus_index.test"),
				),
			},
		},
	})
}

func TestAccResourceIndex_ScalarBitmap(t *testing.T) {
	collectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexConfig_ScalarBitmap(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_index.test", "field_name", "is_active"),
					resource.TestCheckResourceAttr("milvus_index.test", "index_type", "BITMAP"),
					testAccCheckIndexExists("milvus_index.test"),
				),
			},
		},
	})
}

func TestAccResourceIndex_IvfFlat(t *testing.T) {
	collectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexConfig_IvfFlat(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_index.test", "index_type", "IVF_FLAT"),
					resource.TestCheckResourceAttr("milvus_index.test", "metric_type", "COSINE"),
					testAccCheckIndexExists("milvus_index.test"),
				),
			},
		},
	})
}

func testAccResourceIndexConfig_VectorFlat(collectionName string) string {
	return fmt.Sprintf(`
provider "milvus" {
  address = "localhost:19530"
}

resource "milvus_collection" "test" {
  name = "%s"

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

resource "milvus_index" "test" {
  collection_name = milvus_collection.test.name
  field_name      = "embedding"
  index_type      = "FLAT"
  metric_type     = "L2"
  index_name      = "embedding_flat"
}
`, collectionName)
}

func testAccResourceIndexConfig_VectorHnsw(collectionName string) string {
	return fmt.Sprintf(`
provider "milvus" {
  address = "localhost:19530"
}

resource "milvus_collection" "test" {
  name = "%s"

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

resource "milvus_index" "test" {
  collection_name = milvus_collection.test.name
  field_name      = "embedding"
  index_type      = "HNSW"
  metric_type     = "COSINE"
  index_name      = "embedding_hnsw"

  index_params = {
    m               = 8
    ef_construction = 200
  }
}
`, collectionName)
}

func testAccResourceIndexConfig_ScalarBitmap(collectionName string) string {
	return fmt.Sprintf(`
provider "milvus" {
  address = "localhost:19530"
}

resource "milvus_collection" "test" {
  name = "%s"

  fields = [
    {
      name           = "id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name     = "is_active"
      data_type = "Bool"
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]
}

resource "milvus_index" "test" {
  collection_name = milvus_collection.test.name
  field_name      = "is_active"
  index_type      = "BITMAP"
  metric_type     = "L2"
  index_name      = "is_active_bitmap"
}
`, collectionName)
}

func testAccResourceIndexConfig_IvfFlat(collectionName string) string {
	return fmt.Sprintf(`
provider "milvus" {
  address = "localhost:19530"
}

resource "milvus_collection" "test" {
  name = "%s"

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

resource "milvus_index" "test" {
  collection_name = milvus_collection.test.name
  field_name      = "embedding"
  index_type      = "IVF_FLAT"
  metric_type     = "COSINE"
  index_name      = "embedding_ivf"

  index_params = {
    nlist = 128
  }
}
`, collectionName)
}

func testAccCheckIndexExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No index field_name is set")
		}

		// Get the provider configured client
		client := testAccProviderConfig.Client
		if client == nil {
			return fmt.Errorf("Provider not configured")
		}

		collectionName := rs.Primary.Attributes["collection_name"]
		indexName := rs.Primary.Attributes["index_name"]
		fieldName := rs.Primary.Attributes["field_name"]

		if indexName == "" {
			indexName = fieldName
		}

		// Try to describe the index
		opt := milvusclient.NewDescribeIndexOption(collectionName, indexName)
		_, err := client.DescribeIndex(context.Background(), opt)
		if err != nil {
			return fmt.Errorf("Index %s.%s does not exist: %v", collectionName, indexName, err)
		}

		return nil
	}
}
