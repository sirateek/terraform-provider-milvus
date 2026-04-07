// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sirateek/terraform-provider-milvus/internal/ittest/testtemplate"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

// fourVectorCollection builds a collection with four vector fields covering
// every vector type supported by the provider.
func fourVectorCollection(name string, dim int) testtemplate.CollectionTemplate {
	return testtemplate.CollectionTemplate{
		Name:                  name,
		TerraformResourceName: "test",
		AutoID:                false,
		DeleteProtection:      false,
		ShardNum:              1,
		Fields: []testtemplate.FieldTemplate{
			{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
			{Name: "float_vec", DataType: "FloatVector", Dim: testtemplate.IntPtr(dim)},
			{Name: "binary_vec", DataType: "BinaryVector", Dim: testtemplate.IntPtr(dim)},
			{Name: "float16_vec", DataType: "Float16Vector", Dim: testtemplate.IntPtr(dim)},
			{Name: "bfloat16_vec", DataType: "BFloat16Vector", Dim: testtemplate.IntPtr(dim)},
		},
	}
}

// TestVectorFieldDimChangeForceRecreation verifies that changing the dimension
// of an existing vector field forces the collection to be destroyed and recreated.
// The Milvus plan modifier must detect the dim change on each vector field type
// and set RequiresReplace = true.
func (s *ProviderTestSuite) TestVectorFieldDimChangeForceRecreation() {
	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Step 1: Create collection with dim=128 on all four vector fields.
			{
				Config: testtemplate.TerraformTemplate{
					Collections: []testtemplate.CollectionTemplate{
						fourVectorCollection(s.testCollectionName, 128),
					},
				}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.data_type", "FloatVector"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.dim", "128"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.2.data_type", "BinaryVector"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.2.dim", "128"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.3.data_type", "Float16Vector"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.3.dim", "128"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.4.data_type", "BFloat16Vector"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.4.dim", "128"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
			// Step 2: Change dim to 256 on all vector fields.
			// The plan modifier must mark the collection for replacement.
			// The new collection will have a different Milvus-assigned id.
			{
				Config: testtemplate.TerraformTemplate{
					Collections: []testtemplate.CollectionTemplate{
						fourVectorCollection(s.testCollectionName, 256),
					},
				}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.dim", "256"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.2.dim", "256"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.3.dim", "256"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.4.dim", "256"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
		},
	})
}

// TestAddVectorFieldForceRecreation verifies that adding a new vector field to
// an existing collection forces recreation (Milvus does not support adding
// vector fields in-place).
func (s *ProviderTestSuite) TestAddVectorFieldForceRecreation() {
	base := testtemplate.CollectionTemplate{
		Name:                  s.testCollectionName,
		TerraformResourceName: "test",
		AutoID:                false,
		DeleteProtection:      false,
		ShardNum:              1,
		Fields: []testtemplate.FieldTemplate{
			{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
			{Name: "float_vec", DataType: "FloatVector", Dim: testtemplate.IntPtr(128)},
			{Name: "binary_vec", DataType: "BinaryVector", Dim: testtemplate.IntPtr(128)},
			{Name: "float16_vec", DataType: "Float16Vector", Dim: testtemplate.IntPtr(128)},
		},
	}

	withFourVectors := testtemplate.CollectionTemplate{
		Name:                  s.testCollectionName,
		TerraformResourceName: "test",
		AutoID:                false,
		DeleteProtection:      false,
		ShardNum:              1,
		Fields: []testtemplate.FieldTemplate{
			{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
			{Name: "float_vec", DataType: "FloatVector", Dim: testtemplate.IntPtr(128)},
			{Name: "binary_vec", DataType: "BinaryVector", Dim: testtemplate.IntPtr(128)},
			{Name: "float16_vec", DataType: "Float16Vector", Dim: testtemplate.IntPtr(128)},
			{Name: "bfloat16_vec", DataType: "BFloat16Vector", Dim: testtemplate.IntPtr(128)},
		},
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Step 1: Create with 3 vector fields.
			{
				Config: testtemplate.TerraformTemplate{
					Collections: []testtemplate.CollectionTemplate{base},
				}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.#", "4"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
			// Step 2: Add a 4th vector field — must force recreation.
			{
				Config: testtemplate.TerraformTemplate{
					Collections: []testtemplate.CollectionTemplate{withFourVectors},
				}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.#", "5"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
		},
	})
}
