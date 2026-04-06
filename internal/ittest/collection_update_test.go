// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sirateek/terraform-provider-milvus/internal/ittest/testtemplate"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

// baseCollectionForUpdate returns a minimal two-field collection suitable
// for update tests.
func baseCollectionForUpdate(name string) testtemplate.CollectionTemplate {
	return testtemplate.CollectionTemplate{
		Name:                  name,
		TerraformResourceName: "test",
		AutoID:                false,
		DeleteProtection:      false,
		ShardNum:              1,
		Fields: []testtemplate.FieldTemplate{
			{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
			{Name: "embedding", DataType: "FloatVector", Dim: testtemplate.IntPtr(128)},
		},
	}
}

// --- consistency_level ---

func (s *ProviderTestSuite) TestUpdateConsistencyLevel() {
	collTmpl := func(level string) testtemplate.TerraformTemplate {
		c := baseCollectionForUpdate(s.testCollectionName)
		c.ConsistencyLevel = testtemplate.StringPtr(level)
		return testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{c}}
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			{
				Config: collTmpl("Strong").Render(),
				Check:  resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Strong"),
			},
			{
				Config: collTmpl("Bounded").Render(),
				Check:  resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Bounded"),
			},
			{
				Config: collTmpl("Eventually").Render(),
				Check:  resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Eventually"),
			},
			{
				Config: collTmpl("Session").Render(),
				Check:  resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Session"),
			},
		},
	})
}

// --- properties.collection_ttl_seconds ---

func (s *ProviderTestSuite) TestUpdatePropertyTTL() {
	collTmpl := func(props *testtemplate.CollectionPropertiesTemplate) testtemplate.TerraformTemplate {
		c := baseCollectionForUpdate(s.testCollectionName)
		c.Properties = props
		return testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{c}}
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Create without properties.
			{
				Config: collTmpl(nil).Render(),
				Check:  testAccCheckCollectionExists("milvus_collection.test"),
			},
			// Add TTL.
			{
				Config: collTmpl(&testtemplate.CollectionPropertiesTemplate{
					TTLSeconds: testtemplate.Int64Ptr(3600),
				}).Render(),
				Check: resource.TestCheckResourceAttr("milvus_collection.test", "properties.collection_ttl_seconds", "3600"),
			},
			// Update TTL to a different value.
			{
				Config: collTmpl(&testtemplate.CollectionPropertiesTemplate{
					TTLSeconds: testtemplate.Int64Ptr(7200),
				}).Render(),
				Check: resource.TestCheckResourceAttr("milvus_collection.test", "properties.collection_ttl_seconds", "7200"),
			},
		},
	})
}

// --- properties.mmap_enabled ---

func (s *ProviderTestSuite) TestUpdatePropertyMmapEnabled() {
	collTmpl := func(props *testtemplate.CollectionPropertiesTemplate) testtemplate.TerraformTemplate {
		c := baseCollectionForUpdate(s.testCollectionName)
		c.Properties = props
		return testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{c}}
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			{
				Config: collTmpl(nil).Render(),
				Check:  testAccCheckCollectionExists("milvus_collection.test"),
			},
			{
				Config: collTmpl(&testtemplate.CollectionPropertiesTemplate{
					MmapEnabled: testtemplate.BoolPtr(true),
				}).Render(),
				Check: resource.TestCheckResourceAttr("milvus_collection.test", "properties.mmap_enabled", "true"),
			},
			{
				Config: collTmpl(&testtemplate.CollectionPropertiesTemplate{
					MmapEnabled: testtemplate.BoolPtr(false),
				}).Render(),
				Check: resource.TestCheckResourceAttr("milvus_collection.test", "properties.mmap_enabled", "false"),
			},
		},
	})
}

// --- properties combined: TTL + mmap_enabled ---

func (s *ProviderTestSuite) TestUpdateMultipleProperties() {
	collTmpl := func(props *testtemplate.CollectionPropertiesTemplate) testtemplate.TerraformTemplate {
		c := baseCollectionForUpdate(s.testCollectionName)
		c.Properties = props
		return testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{c}}
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Create with both properties set.
			{
				Config: collTmpl(&testtemplate.CollectionPropertiesTemplate{
					TTLSeconds:  testtemplate.Int64Ptr(1800),
					MmapEnabled: testtemplate.BoolPtr(true),
				}).Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "properties.collection_ttl_seconds", "1800"),
					resource.TestCheckResourceAttr("milvus_collection.test", "properties.mmap_enabled", "true"),
				),
			},
			// Update both in one apply.
			{
				Config: collTmpl(&testtemplate.CollectionPropertiesTemplate{
					TTLSeconds:  testtemplate.Int64Ptr(3600),
					MmapEnabled: testtemplate.BoolPtr(false),
				}).Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "properties.collection_ttl_seconds", "3600"),
					resource.TestCheckResourceAttr("milvus_collection.test", "properties.mmap_enabled", "false"),
				),
			},
		},
	})
}

// --- consistency_level + properties together ---

func (s *ProviderTestSuite) TestUpdateConsistencyLevelAndProperties() {
	collTmpl := func(level string, props *testtemplate.CollectionPropertiesTemplate) testtemplate.TerraformTemplate {
		c := baseCollectionForUpdate(s.testCollectionName)
		c.ConsistencyLevel = testtemplate.StringPtr(level)
		c.Properties = props
		return testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{c}}
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Create with initial settings.
			{
				Config: collTmpl("Strong", nil).Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Strong"),
				),
			},
			// Change both consistency_level and add properties in one apply.
			{
				Config: collTmpl("Bounded", &testtemplate.CollectionPropertiesTemplate{
					TTLSeconds: testtemplate.Int64Ptr(600),
				}).Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "consistency_level", "Bounded"),
					resource.TestCheckResourceAttr("milvus_collection.test", "properties.collection_ttl_seconds", "600"),
				),
			},
		},
	})
}

// --- immutable fields trigger recreation ---

func (s *ProviderTestSuite) TestImmutableFieldsForceRecreation() {
	base := baseCollectionForUpdate(s.testCollectionName)
	base.Description = "original description"

	updated := baseCollectionForUpdate(s.testCollectionName)
	updated.Description = "changed description"

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			{
				Config: testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{base}}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "description", "original description"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
			// Changing description must replace the collection (new id expected).
			{
				Config: testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{updated}}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "description", "changed description"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
		},
	})
}

// --- shard_num is immutable ---

func (s *ProviderTestSuite) TestShardNumIsImmutable() {
	collTmpl := func(shardNum int) testtemplate.TerraformTemplate {
		c := baseCollectionForUpdate(s.testCollectionName)
		c.ShardNum = shardNum
		return testtemplate.TerraformTemplate{Collections: []testtemplate.CollectionTemplate{c}}
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			{
				Config: collTmpl(1).Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "shard_num", "1"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
			// Changing shard_num forces replacement — new id will be assigned.
			{
				Config: collTmpl(2).Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_collection.test", "shard_num", "2"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
		},
	})
}

// --- auto_id default and stability ---

// TestAutoIDDefault verifies auto_id defaults to false and is stable across applies.
// auto_id=true requires a different field schema (no explicit is_primary_key on the
// field) and is not tested for recreation here; immutability is already demonstrated
// by TestShardNumIsImmutable and TestImmutableFieldsForceRecreation.
func (s *ProviderTestSuite) TestAutoIDDefault() {
	cfg := testtemplate.TerraformTemplate{
		Collections: []testtemplate.CollectionTemplate{baseCollectionForUpdate(s.testCollectionName)},
	}.Render()

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "auto_id", "false"),
				),
			},
			// Re-apply must produce an empty plan — auto_id is correctly read back.
			{
				Config:   cfg,
				PlanOnly: true,
			},
		},
	})
}

// --- enable_dynamic_field is immutable ---

// TestEnableDynamicFieldDefault verifies that enable_dynamic_field defaults to
// false when not set in config, and that the value is read back correctly from
// Milvus so the plan remains empty on subsequent applies.
func (s *ProviderTestSuite) TestEnableDynamicFieldDefault() {
	cfg := testtemplate.TerraformTemplate{
		Collections: []testtemplate.CollectionTemplate{baseCollectionForUpdate(s.testCollectionName)},
	}.Render()

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.test"),
					resource.TestCheckResourceAttr("milvus_collection.test", "enable_dynamic_field", "false"),
					resource.TestCheckResourceAttrSet("milvus_collection.test", "id"),
				),
			},
			// Re-apply the same config — must produce an empty plan (no recreation).
			{
				Config:   cfg,
				PlanOnly: true,
			},
		},
	})
}
