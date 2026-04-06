// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/ittest/testtemplate"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

func (s *ProviderTestSuite) TestCreateBasicCollection() {
	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testtemplate.TerraformTemplate{
					Collections: []testtemplate.CollectionTemplate{
						{
							Name:                  s.testCollectionName,
							TerraformResourceName: "test",
							Description:           "Test collection",
							AutoID:                false,
							DeleteProtection:      false,
							ShardNum:              1,
							Fields: []testtemplate.FieldTemplate{
								{
									Name:         "id",
									DataType:     "Int64",
									IsPrimaryKey: testtemplate.BoolPtr(true),
								},
								{
									Name:     "embedding",
									DataType: "FloatVector",
									Dim:      testtemplate.IntPtr(768),
								},
							},
						},
					},
				}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "name", s.testCollectionName),
					resource.TestCheckResourceAttr("milvus_collection.test", "description", "Test collection"),
					resource.TestCheckResourceAttr("milvus_collection.test", "auto_id", "false"),
					// Verify computed attributes are populated
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
					resource.TestCheckResourceAttr("milvus_collection.test", "shard_num", "1"),
					resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Strong"),
					// Verify field[0]: id (Int64 primary key)
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.name", "id"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.data_type", "Int64"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.is_primary_key", "true"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.is_auto_id", "false"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.is_partition_key", "false"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.is_clustering_key", "false"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.0.nullable", "false"),
					// Verify field[1]: embedding (FloatVector)
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.name", "embedding"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.data_type", "FloatVector"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.dim", "768"),
					resource.TestCheckResourceAttr("milvus_collection.test", "fields.1.is_primary_key", "false"),
					testAccCheckCollectionExists("milvus_collection.test"),
					testAccCheckBasicCollectionSchema("milvus_collection.test"),
				),
			},
		},
	})
}

func testAccCheckBasicCollectionSchema(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		collectionName := rs.Primary.Attributes["name"]

		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		coll, err := client.DescribeCollection(
			context.Background(),
			milvusclient.NewDescribeCollectionOption(collectionName),
		)
		if err != nil {
			return fmt.Errorf("failed to describe collection %s: %v", collectionName, err)
		}

		schema := coll.Schema
		if schema == nil {
			return fmt.Errorf("collection %s has no schema", collectionName)
		}

		type expectedField struct {
			name         string
			dataType     entity.FieldType
			isPrimaryKey bool
			dim          string // non-empty for vector fields
		}

		expected := []expectedField{
			{
				name:         "id",
				dataType:     entity.FieldTypeInt64,
				isPrimaryKey: true,
			},
			{
				name:     "embedding",
				dataType: entity.FieldTypeFloatVector,
				dim:      "768",
			},
		}

		// Build a name→field map from the actual schema for easy lookup.
		fieldMap := make(map[string]*entity.Field, len(schema.Fields))
		for _, f := range schema.Fields {
			fieldMap[f.Name] = f
		}

		for _, exp := range expected {
			f, exists := fieldMap[exp.name]
			if !exists {
				return fmt.Errorf("schema field %q not found in collection %s", exp.name, collectionName)
			}
			if f.DataType != exp.dataType {
				return fmt.Errorf("field %q: expected data type %v, got %v", exp.name, exp.dataType, f.DataType)
			}
			if f.PrimaryKey != exp.isPrimaryKey {
				return fmt.Errorf("field %q: expected is_primary_key=%v, got %v", exp.name, exp.isPrimaryKey, f.PrimaryKey)
			}
			if exp.dim != "" {
				actualDim, hasDim := f.TypeParams["dim"]
				if !hasDim {
					return fmt.Errorf("field %q: expected dim=%s but TypeParams has no dim", exp.name, exp.dim)
				}
				if actualDim != exp.dim {
					return fmt.Errorf("field %q: expected dim=%s, got %s", exp.name, exp.dim, actualDim)
				}
			}
		}

		return nil
	}
}

func (s *ProviderTestSuite) TestDeleteProtectionCollection() {
	collectionTemplate := func(deleteProtection bool) testtemplate.TerraformTemplate {
		return testtemplate.TerraformTemplate{
			Collections: []testtemplate.CollectionTemplate{
				{
					Name:                  s.testCollectionName,
					TerraformResourceName: "test_delete_protection",
					Description:           "Test collection",
					AutoID:                false,
					DeleteProtection:      deleteProtection,
					ShardNum:              1,
					Fields: []testtemplate.FieldTemplate{
						{
							Name:         "id",
							DataType:     "Int64",
							IsPrimaryKey: testtemplate.BoolPtr(true),
						},
						{
							Name:     "embedding",
							DataType: "FloatVector",
							Dim:      testtemplate.IntPtr(768),
						},
					},
				},
			},
		}
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Step 1: Create with delete_protection = true
			{
				Config: collectionTemplate(true).Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test_delete_protection"),
					resource.TestCheckResourceAttr("milvus_collection.test_delete_protection", "delete_protection", "true"),
				),
			},
			// Step 2: Attempt removal — must fail because delete_protection is still true
			{
				Config:      testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{}}.Render(),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Collection %s is protected from deletion", s.testCollectionName)),
			},
			// Step 3: Disable delete_protection so the post-test destroy succeeds
			{
				Config: collectionTemplate(false).Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test_delete_protection", "delete_protection", "false"),
				),
			},
		},
	})
}

func (s *ProviderTestSuite) TestCreateAllFieldTypesCollection() {
	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			{
				Config: testtemplate.TerraformTemplate{
					Collections: []testtemplate.CollectionTemplate{
						{
							Name:                  s.testCollectionName,
							TerraformResourceName: "test_all_fields",
							Description:           "Test all field types",
							AutoID:                false,
							DeleteProtection:      false,
							ShardNum:              1,
							Fields: []testtemplate.FieldTemplate{
								{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
								{Name: "int8_field", DataType: "Int8"},
								{Name: "int16_field", DataType: "Int16"},
								{Name: "int32_field", DataType: "Int32"},
								{Name: "float_field", DataType: "Float"},
								{Name: "bool_field", DataType: "Bool"},
								{Name: "varchar_field", DataType: "VarChar", MaxLength: testtemplate.IntPtr(256)},
								{Name: "json_field", DataType: "JSON"},
								{Name: "array_field", DataType: "Array", ElementType: testtemplate.StringPtr("Int64"), MaxCapacity: testtemplate.IntPtr(100)},
								{Name: "float_vector", DataType: "FloatVector", Dim: testtemplate.IntPtr(128)},
								{Name: "binary_vector", DataType: "BinaryVector", Dim: testtemplate.IntPtr(128)},
								{Name: "float16_vector", DataType: "Float16Vector", Dim: testtemplate.IntPtr(128)},
								{Name: "bfloat16_vector", DataType: "BFloat16Vector", Dim: testtemplate.IntPtr(128)},
							},
						},
					},
				}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test_all_fields"),
					testAccCheckAllFieldTypesSchema("milvus_collection.test_all_fields"),
				),
			},
		},
	})
}

func testAccCheckAllFieldTypesSchema(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		collectionName := rs.Primary.Attributes["name"]

		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		coll, err := client.DescribeCollection(
			context.Background(),
			milvusclient.NewDescribeCollectionOption(collectionName),
		)
		if err != nil {
			return fmt.Errorf("failed to describe collection %s: %v", collectionName, err)
		}

		if coll.Schema == nil {
			return fmt.Errorf("collection %s has no schema", collectionName)
		}

		type expectedField struct {
			dataType    entity.FieldType
			isPrimary   bool
			dim         string
			maxLength   string
			maxCapacity string
			elementType entity.FieldType
		}

		expected := map[string]expectedField{
			"id":              {dataType: entity.FieldTypeInt64, isPrimary: true},
			"int8_field":      {dataType: entity.FieldTypeInt8},
			"int16_field":     {dataType: entity.FieldTypeInt16},
			"int32_field":     {dataType: entity.FieldTypeInt32},
			"float_field":     {dataType: entity.FieldTypeFloat},
			"bool_field":      {dataType: entity.FieldTypeBool},
			"varchar_field":   {dataType: entity.FieldTypeVarChar, maxLength: "256"},
			"json_field":      {dataType: entity.FieldTypeJSON},
			"array_field":     {dataType: entity.FieldTypeArray, maxCapacity: "100", elementType: entity.FieldTypeInt64},
			"float_vector":    {dataType: entity.FieldTypeFloatVector, dim: "128"},
			"binary_vector":   {dataType: entity.FieldTypeBinaryVector, dim: "128"},
			"float16_vector":  {dataType: entity.FieldTypeFloat16Vector, dim: "128"},
			"bfloat16_vector": {dataType: entity.FieldTypeBFloat16Vector, dim: "128"},
		}

		fieldMap := make(map[string]*entity.Field, len(coll.Schema.Fields))
		for _, f := range coll.Schema.Fields {
			fieldMap[f.Name] = f
		}

		for name, exp := range expected {
			f, exists := fieldMap[name]
			if !exists {
				return fmt.Errorf("field %q not found in schema", name)
			}
			if f.DataType != exp.dataType {
				return fmt.Errorf("field %q: expected data type %v, got %v", name, exp.dataType, f.DataType)
			}
			if f.PrimaryKey != exp.isPrimary {
				return fmt.Errorf("field %q: expected is_primary_key=%v, got %v", name, exp.isPrimary, f.PrimaryKey)
			}
			if exp.dim != "" {
				if got := f.TypeParams["dim"]; got != exp.dim {
					return fmt.Errorf("field %q: expected dim=%s, got %s", name, exp.dim, got)
				}
			}
			if exp.maxLength != "" {
				if got := f.TypeParams["max_length"]; got != exp.maxLength {
					return fmt.Errorf("field %q: expected max_length=%s, got %s", name, exp.maxLength, got)
				}
			}
			if exp.maxCapacity != "" {
				if got := f.TypeParams["max_capacity"]; got != exp.maxCapacity {
					return fmt.Errorf("field %q: expected max_capacity=%s, got %s", name, exp.maxCapacity, got)
				}
			}
			if exp.elementType != entity.FieldTypeNone {
				if f.ElementType != exp.elementType {
					return fmt.Errorf("field %q: expected element_type=%v, got %v", name, exp.elementType, f.ElementType)
				}
			}
		}

		return nil
	}
}
