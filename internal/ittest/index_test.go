// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/ittest/testtemplate"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

// baseCollectionForIndex returns a CollectionTemplate with a primary key field
// and a FloatVector embedding field, suitable as a base for index tests.
func baseCollectionForIndex(name string, extraFields ...testtemplate.FieldTemplate) testtemplate.CollectionTemplate {
	fields := []testtemplate.FieldTemplate{
		{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
		{Name: "embedding", DataType: "FloatVector", Dim: testtemplate.IntPtr(768)},
	}
	fields = append(fields, extraFields...)
	return testtemplate.CollectionTemplate{
		Name:                  name,
		TerraformResourceName: "test",
		AutoID:                false,
		DeleteProtection:      false,
		ShardNum:              1,
		Fields:                fields,
	}
}

//func (s *ProviderTestSuite) TestCreateIndex_VectorFlat() {
//	resource.Test(s.T(), resource.TestCase{
//		PreCheck:                 func() { provider.PreCheck(s.T()) },
//		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
//		CheckDestroy:             testAccCheckCollectionAndIndexesDestroyed(s.testCollectionName),
//		Steps: []resource.TestStep{
//			{
//				Config: testtemplate.TerraformTemplate{
//					Collections: []testtemplate.CollectionTemplate{
//						baseCollectionForIndex(s.testCollectionName),
//					},
//					Indexes: []testtemplate.IndexTemplate{
//						{
//							TerraformResourceName: "test_vector_flat",
//							CollectionName:        "milvus_collection.test.name",
//							FieldName:             "embedding",
//							IndexName:             "embedding_flat",
//							IndexType:             "FLAT",
//							MetricType:            "L2",
//						},
//					},
//				}.Render(),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("milvus_index.test_vector_flat", "collection_name", s.testCollectionName),
//					resource.TestCheckResourceAttr("milvus_index.test_vector_flat", "field_name", "embedding"),
//					resource.TestCheckResourceAttr("milvus_index.test_vector_flat", "index_type", "FLAT"),
//					resource.TestCheckResourceAttr("milvus_index.test_vector_flat", "metric_type", "L2"),
//					testAccCheckIndexExists("milvus_index.test_vector_flat"),
//				),
//			},
//		},
//	})
//}
//
//func (s *ProviderTestSuite) TestCreateIndex_VectorHnsw() {
//	resource.Test(s.T(), resource.TestCase{
//		PreCheck:                 func() { provider.PreCheck(s.T()) },
//		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
//		CheckDestroy:             testAccCheckCollectionAndIndexesDestroyed(s.testCollectionName),
//		Steps: []resource.TestStep{
//			{
//				Config: testtemplate.TerraformTemplate{
//					Collections: []testtemplate.CollectionTemplate{
//						baseCollectionForIndex(s.testCollectionName),
//					},
//					Indexes: []testtemplate.IndexTemplate{
//						{
//							TerraformResourceName: "test",
//							CollectionName:        "milvus_collection.test.name",
//							FieldName:             "embedding",
//							IndexName:             "embedding_hnsw",
//							IndexType:             "HNSW",
//							MetricType:            "COSINE",
//							IndexParams: map[string]any{
//								"m":               8,
//								"ef_construction": 200,
//							},
//						},
//					},
//				}.Render(),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("milvus_index.test", "index_type", "HNSW"),
//					resource.TestCheckResourceAttr("milvus_index.test", "metric_type", "COSINE"),
//					testAccCheckIndexExists("milvus_index.test"),
//				),
//			},
//		},
//	})
//}
//
//func (s *ProviderTestSuite) TestCreateIndex_ScalarBitmap() {
//	resource.Test(s.T(), resource.TestCase{
//		PreCheck:                 func() { provider.PreCheck(s.T()) },
//		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
//		CheckDestroy:             testAccCheckCollectionAndIndexesDestroyed(s.testCollectionName),
//		Steps: []resource.TestStep{
//			{
//				Config: testtemplate.TerraformTemplate{
//					Collections: []testtemplate.CollectionTemplate{
//						baseCollectionForIndex(s.testCollectionName,
//							testtemplate.FieldTemplate{Name: "is_active", DataType: "Bool"},
//						),
//					},
//					Indexes: []testtemplate.IndexTemplate{
//						{
//							TerraformResourceName: "test",
//							CollectionName:        "milvus_collection.test.name",
//							FieldName:             "is_active",
//							IndexName:             "is_active_bitmap",
//							IndexType:             "BITMAP",
//							MetricType:            "L2",
//						},
//					},
//				}.Render(),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("milvus_index.test", "field_name", "is_active"),
//					resource.TestCheckResourceAttr("milvus_index.test", "index_type", "BITMAP"),
//					testAccCheckIndexExists("milvus_index.test"),
//				),
//			},
//		},
//	})
//}
//
//func (s *ProviderTestSuite) TestCreateIndex_IvfFlat() {
//	resource.Test(s.T(), resource.TestCase{
//		PreCheck:                 func() { provider.PreCheck(s.T()) },
//		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
//		CheckDestroy:             testAccCheckCollectionAndIndexesDestroyed(s.testCollectionName),
//		Steps: []resource.TestStep{
//			{
//				Config: testtemplate.TerraformTemplate{
//					Collections: []testtemplate.CollectionTemplate{
//						baseCollectionForIndex(s.testCollectionName),
//					},
//					Indexes: []testtemplate.IndexTemplate{
//						{
//							TerraformResourceName: "test",
//							CollectionName:        "milvus_collection.test.name",
//							FieldName:             "embedding",
//							IndexName:             "embedding_ivf",
//							IndexType:             "IVF_FLAT",
//							MetricType:            "COSINE",
//							IndexParams: map[string]any{
//								"nlist": 128,
//							},
//						},
//					},
//				}.Render(),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("milvus_index.test", "index_type", "IVF_FLAT"),
//					resource.TestCheckResourceAttr("milvus_index.test", "metric_type", "COSINE"),
//					testAccCheckIndexExists("milvus_index.test"),
//				),
//			},
//		},
//	})
//}

// TestDeleteCollectionWithIndex verifies that deleting a collection while a
// milvus_index resource still exists is blocked with a descriptive error.
//func (s *ProviderTestSuite) TestDeleteCollectionWithIndex() {
//	collectionConfig := testtemplate.TerraformTemplate{
//		Collections: []testtemplate.CollectionTemplate{
//			baseCollectionForIndex(s.testCollectionName),
//		},
//		Indexes: []testtemplate.IndexTemplate{
//			{
//				TerraformResourceName: "test",
//				CollectionName:        "milvus_collection.test.name",
//				FieldName:             "embedding",
//				IndexName:             "embedding_flat",
//				IndexType:             "FLAT",
//				MetricType:            "L2",
//			},
//		},
//	}
//
//	resource.Test(s.T(), resource.TestCase{
//		PreCheck:                 func() { provider.PreCheck(s.T()) },
//		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
//		CheckDestroy:             testAccCheckCollectionAndIndexesDestroyed(s.testCollectionName),
//		Steps: []resource.TestStep{
//			// Step 1: Create the collection and index
//			{
//				Config: collectionConfig.Render(),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccCheckCollectionExists("milvus_collection.test"),
//					testAccCheckIndexExists("milvus_index.test"),
//				),
//			},
//			// Step 2: Remove only the collection — must fail because the index still exists
//			{
//				Config: testtemplate.TerraformTemplate{
//					Indexes: []testtemplate.IndexTemplate{
//						{
//							TerraformResourceName: "test",
//							CollectionName:        fmt.Sprintf("%q", s.testCollectionName),
//							FieldName:             "embedding",
//							IndexName:             "embedding_flat",
//							IndexType:             "FLAT",
//							MetricType:            "L2",
//						},
//					},
//				}.Render(),
//				ExpectError: regexp.MustCompile(`still has indexes`),
//			},
//			// Step 3: Restore the full config so CheckDestroy can clean up cleanly
//			{
//				Config: collectionConfig.Render(),
//			},
//		},
//	})
//}

func testAccCheckIndexExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no index ID is set")
		}

		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		collectionName := rs.Primary.Attributes["collection_name"]
		indexName := rs.Primary.Attributes["index_name"]
		fieldName := rs.Primary.Attributes["field_name"]
		if indexName == "" {
			indexName = fieldName
		}

		_, err := client.DescribeIndex(
			context.Background(),
			milvusclient.NewDescribeIndexOption(collectionName, indexName),
		)
		if err != nil {
			return fmt.Errorf("index %s.%s does not exist: %v", collectionName, indexName, err)
		}

		return nil
	}
}
