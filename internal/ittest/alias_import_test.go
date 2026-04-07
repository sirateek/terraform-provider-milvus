// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/ittest/testtemplate"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

// TestImportAlias verifies that a milvus_alias created outside Terraform can be
// imported using the alias name as the import ID, that collection_name is
// populated from Milvus after import, and that a subsequent plan produces no diff.
func (s *ProviderTestSuite) TestImportAlias() {
	aliasName := fmt.Sprintf("alias_%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	// Config that matches the externally-created collection and alias.
	cfg := testtemplate.TerraformTemplate{
		Collections: []testtemplate.CollectionTemplate{
			baseCollectionForAlias(s.testCollectionName, "test"),
		},
		Aliases: []testtemplate.AliasTemplate{
			{
				TerraformResourceName: "test",
				AliasName:             aliasName,
				CollectionName:        fmt.Sprintf("%q", s.testCollectionName),
			},
		},
	}.Render()

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAliasDestroyed(aliasName),
			testAccCheckCollectionDestroyed(s.testCollectionName),
		),
		Steps: []resource.TestStep{
			// Step 1: Create collection + alias externally via Milvus client, then
			// import the collection so it is tracked in state.
			{
				PreConfig: func() {
					createAliasExternally(s, s.testCollectionName, aliasName)
				},
				Config:                  cfg,
				ResourceName:            "milvus_collection.test",
				ImportState:             true,
				ImportStateId:           s.testCollectionName,
				ImportStatePersist:      true,
				ImportStateVerifyIgnore: []string{"delete_protection"},
			},
			// Step 2: Import the alias. Collection is already in state from step 1.
			{
				Config:             cfg,
				ResourceName:       "milvus_alias.test",
				ImportState:        true,
				ImportStateId:      aliasName,
				ImportStatePersist: true,
			},
			// Step 3: Plan with matching config — must be empty (no diff).
			{
				Config:   cfg,
				PlanOnly: true,
			},
			// Step 4: Apply and verify state is fully correct.
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAliasExists("milvus_alias.test"),
					resource.TestCheckResourceAttr("milvus_alias.test", "name", aliasName),
					resource.TestCheckResourceAttr("milvus_alias.test", "collection_name", s.testCollectionName),
					testAccCheckAliasImportedState(aliasName, s.testCollectionName),
				),
			},
		},
	})
}

// createAliasExternally creates a collection and alias directly via the Milvus
// client, bypassing Terraform entirely.
func createAliasExternally(s *ProviderTestSuite, collectionName, aliasName string) {
	s.T().Helper()

	client := provider.AccTestProviderConfig.Client
	s.Require().NotNil(client, "Milvus client not initialised — did PreCheck run?")

	// Create collection.
	schema := entity.NewSchema().
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("embedding").WithDataType(entity.FieldTypeFloatVector).WithDim(128))

	err := client.CreateCollection(
		context.Background(),
		milvusclient.NewCreateCollectionOption(collectionName, schema).
			WithConsistencyLevel(entity.ClStrong),
	)
	s.Require().NoError(err, "failed to create collection %q externally", collectionName)

	// Create alias.
	err = client.CreateAlias(
		context.Background(),
		milvusclient.NewCreateAliasOption(collectionName, aliasName),
	)
	s.Require().NoError(err, "failed to create alias %q externally", aliasName)

	s.T().Logf("Created collection %q and alias %q externally", collectionName, aliasName)
}

// testAccCheckAliasImportedState verifies that collection_name is correctly
// populated in state after an alias import.
func testAccCheckAliasImportedState(aliasName, expectedCollection string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var rs *terraform.ResourceState
		for _, r := range s.RootModule().Resources {
			if r.Type == "milvus_alias" && r.Primary.Attributes["name"] == aliasName {
				rs = r
				break
			}
		}
		if rs == nil {
			return fmt.Errorf("milvus_alias with name=%q not found in state", aliasName)
		}

		attrs := rs.Primary.Attributes

		if attrs["collection_name"] != expectedCollection {
			return fmt.Errorf("collection_name: expected %q, got %q", expectedCollection, attrs["collection_name"])
		}

		return nil
	}
}
