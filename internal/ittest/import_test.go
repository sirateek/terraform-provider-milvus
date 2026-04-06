// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/ittest/testtemplate"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

// TestImportCollection verifies that a collection created outside Terraform
// can be imported into state, that all Milvus-sourced attributes are populated
// correctly, and that a subsequent update does NOT recreate the collection.
//
// The no-recreation check works by inserting a PlanOnly step after import with
// the same config. If auto_id or enable_dynamic_field were not read back from
// Milvus correctly, the plan would show a diff (forcing recreation) and the
// PlanOnly step would fail because it expects an empty plan.
func (s *ProviderTestSuite) TestImportCollection() {
	// cfg matches the collection that will be created externally.
	cfg := testtemplate.TerraformTemplate{
		Collections: []testtemplate.CollectionTemplate{
			{
				Name:                  s.testCollectionName,
				TerraformResourceName: "imported",
				DeleteProtection:      false,
				ShardNum:              1,
				Fields: []testtemplate.FieldTemplate{
					{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
					{Name: "embedding", DataType: "FloatVector", Dim: testtemplate.IntPtr(128)},
				},
			},
		},
	}.Render()

	// cfgUpdated changes consistency_level in-place (no recreation expected).
	cfgUpdated := testtemplate.TerraformTemplate{
		Collections: []testtemplate.CollectionTemplate{
			{
				Name:                  s.testCollectionName,
				TerraformResourceName: "imported",
				DeleteProtection:      false,
				ShardNum:              1,
				ConsistencyLevel:      testtemplate.StringPtr("Bounded"),
				Fields: []testtemplate.FieldTemplate{
					{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
					{Name: "embedding", DataType: "FloatVector", Dim: testtemplate.IntPtr(128)},
				},
			},
		},
	}.Render()

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroyed(s.testCollectionName),
		Steps: []resource.TestStep{
			// Step 1: Import the externally-created collection.
			{
				PreConfig: func() {
					createCollectionExternally(s, s.testCollectionName)
				},
				Config:             cfg,
				ResourceName:       "milvus_collection.imported",
				ImportState:        true,
				ImportStateId:      s.testCollectionName,
				ImportStatePersist: true,
				// delete_protection is provider-only and not stored in Milvus.
				// auto_id and enable_dynamic_field are now read back from
				// coll.Schema after the service.go fix, so they are no longer ignored.
				ImportStateVerifyIgnore: []string{"delete_protection"},
			},
			// Step 2: Plan-only with the same config.
			// The plan must be empty — if auto_id or enable_dynamic_field were
			// not correctly restored from Milvus, Terraform would plan recreation
			// here and this step would fail.
			{
				Config:   cfg,
				PlanOnly: true,
			},
			// Step 3: Apply an in-place update (consistency_level only).
			// Verifies the collection is updated without being destroyed and recreated.
			{
				Config: cfgUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCollectionExists("milvus_collection.imported"),
					resource.TestCheckResourceAttr("milvus_collection.imported", "consistency_level", "Bounded"),
					testAccCheckImportedCollectionState(s.testCollectionName),
				),
			},
		},
	})
}

// createCollectionExternally creates a simple two-field collection directly
// through the Milvus client, bypassing Terraform entirely.
func createCollectionExternally(s *ProviderTestSuite, collectionName string) {
	s.T().Helper()

	client := provider.AccTestProviderConfig.Client
	s.Require().NotNil(client, "Milvus client not initialised — did PreCheck run?")

	schema := entity.NewSchema().
		WithField(
			entity.NewField().
				WithName("id").
				WithDataType(entity.FieldTypeInt64).
				WithIsPrimaryKey(true),
		).
		WithField(
			entity.NewField().
				WithName("embedding").
				WithDataType(entity.FieldTypeFloatVector).
				WithDim(128),
		)

	// Use Strong consistency to match the provider schema default so that a
	// PlanOnly step after import produces an empty plan.
	opt := milvusclient.NewCreateCollectionOption(collectionName, schema).
		WithConsistencyLevel(entity.ClStrong)
	err := client.CreateCollection(context.Background(), opt)
	s.Require().NoError(err, fmt.Sprintf("failed to create collection %q externally", collectionName))

	s.T().Logf("Created collection %q externally via Milvus client", collectionName)
}

// testAccCheckImportedCollectionState verifies that the key Milvus-sourced
// attributes are correctly populated in state after an import.
func testAccCheckImportedCollectionState(collectionName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var rs *terraform.ResourceState
		for _, r := range s.RootModule().Resources {
			if r.Type == "milvus_collection" && r.Primary.Attributes["name"] == collectionName {
				rs = r
				break
			}
		}
		if rs == nil {
			return fmt.Errorf("milvus_collection %q not found in state", collectionName)
		}

		attrs := rs.Primary.Attributes

		if attrs["id"] == "" {
			return fmt.Errorf("id: expected a non-empty value after import, got empty")
		}
		if attrs["shard_num"] == "" {
			return fmt.Errorf("shard_num: expected a non-empty value after import, got empty")
		}
		if attrs["consistency_level"] == "" {
			return fmt.Errorf("consistency_level: expected a non-empty value after import, got empty")
		}
		if attrs["auto_id"] == "" {
			return fmt.Errorf("auto_id: expected a non-empty value after import, got empty")
		}
		if attrs["enable_dynamic_field"] == "" {
			return fmt.Errorf("enable_dynamic_field: expected a non-empty value after import, got empty")
		}

		// Verify the field count in state matches what Milvus reports.
		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		coll, err := client.DescribeCollection(
			context.Background(),
			milvusclient.NewDescribeCollectionOption(collectionName),
		)
		if err != nil {
			return fmt.Errorf("failed to describe collection %q: %v", collectionName, err)
		}

		wantFieldCount := len(coll.Schema.Fields)
		gotFieldCount := 0
		for k := range attrs {
			if len(k) > 8 && k[:7] == "fields." && k[len(k)-5:] == ".name" {
				gotFieldCount++
			}
		}
		if gotFieldCount != wantFieldCount {
			return fmt.Errorf("fields: expected %d fields in state, got %d", wantFieldCount, gotFieldCount)
		}

		return nil
	}
}
