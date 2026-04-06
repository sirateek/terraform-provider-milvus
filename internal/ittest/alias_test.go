// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package ittest

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/ittest/testtemplate"
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

// baseCollectionForAlias returns a CollectionTemplate suitable as a base for alias tests.
func baseCollectionForAlias(name, resourceName string) testtemplate.CollectionTemplate {
	return testtemplate.CollectionTemplate{
		Name:                  name,
		TerraformResourceName: resourceName,
		AutoID:                false,
		DeleteProtection:      false,
		ShardNum:              1,
		Fields: []testtemplate.FieldTemplate{
			{Name: "id", DataType: "Int64", IsPrimaryKey: testtemplate.BoolPtr(true)},
			{Name: "embedding", DataType: "FloatVector", Dim: testtemplate.IntPtr(128)},
		},
	}
}

func (s *ProviderTestSuite) TestCreateAlias() {
	aliasName := fmt.Sprintf("alias_%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAliasDestroyed(aliasName),
			testAccCheckCollectionDestroyed(s.testCollectionName),
		),
		Steps: []resource.TestStep{
			{
				Config: testtemplate.TerraformTemplate{
					Collections: []testtemplate.CollectionTemplate{
						baseCollectionForAlias(s.testCollectionName, "test"),
					},
					Aliases: []testtemplate.AliasTemplate{
						{
							TerraformResourceName: "test",
							AliasName:             aliasName,
							CollectionName:        "milvus_collection.test.name",
						},
					},
				}.Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_alias.test", "name", aliasName),
					resource.TestCheckResourceAttr("milvus_alias.test", "collection_name", s.testCollectionName),
					testAccCheckAliasExists("milvus_alias.test"),
					testAccCheckAliasPointsTo(aliasName, s.testCollectionName),
				),
			},
		},
	})
}

func (s *ProviderTestSuite) TestUpdateAlias() {
	aliasName := fmt.Sprintf("alias_%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	secondCollectionName := fmt.Sprintf("tf_test_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	aliasConfig := func(targetCollectionRef string) testtemplate.TerraformTemplate {
		return testtemplate.TerraformTemplate{
			Collections: []testtemplate.CollectionTemplate{
				baseCollectionForAlias(s.testCollectionName, "coll_a"),
				baseCollectionForAlias(secondCollectionName, "coll_b"),
			},
			Aliases: []testtemplate.AliasTemplate{
				{
					TerraformResourceName: "test",
					AliasName:             aliasName,
					CollectionName:        targetCollectionRef,
				},
			},
		}
	}

	resource.Test(s.T(), resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(s.T()) },
		ProtoV6ProviderFactories: provider.ProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAliasDestroyed(aliasName),
			testAccCheckCollectionDestroyed(s.testCollectionName),
			testAccCheckCollectionDestroyed(secondCollectionName),
		),
		Steps: []resource.TestStep{
			// Step 1: Create alias pointing to first collection.
			{
				Config: aliasConfig("milvus_collection.coll_a.name").Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_alias.test", "collection_name", s.testCollectionName),
					testAccCheckAliasPointsTo(aliasName, s.testCollectionName),
				),
			},
			// Step 2: Update alias to point to second collection.
			{
				Config: aliasConfig("milvus_collection.coll_b.name").Render(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("milvus_alias.test", "collection_name", secondCollectionName),
					testAccCheckAliasPointsTo(aliasName, secondCollectionName),
				),
			},
		},
	})
}

func testAccCheckAliasExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		aliasName := rs.Primary.Attributes["name"]
		if aliasName == "" {
			return fmt.Errorf("alias name not set")
		}

		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		_, err := client.DescribeAlias(context.Background(), milvusclient.NewDescribeAliasOption(aliasName))
		if err != nil {
			return fmt.Errorf("alias %s does not exist: %v", aliasName, err)
		}

		return nil
	}
}

func testAccCheckAliasPointsTo(aliasName, expectedCollection string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		aliasResult, err := client.DescribeAlias(context.Background(), milvusclient.NewDescribeAliasOption(aliasName))
		if err != nil {
			return fmt.Errorf("failed to describe alias %s: %v", aliasName, err)
		}

		if aliasResult.CollectionName != expectedCollection {
			return fmt.Errorf("alias %s points to %q, want %q", aliasName, aliasResult.CollectionName, expectedCollection)
		}

		return nil
	}
}

func testAccCheckAliasDestroyed(aliasName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.AccTestProviderConfig.Client
		if client == nil {
			return fmt.Errorf("provider not configured")
		}

		_, err := client.DescribeAlias(context.Background(), milvusclient.NewDescribeAliasOption(aliasName))
		if err == nil {
			return fmt.Errorf("alias %s still exists after destroy", aliasName)
		}

		return nil
	}
}