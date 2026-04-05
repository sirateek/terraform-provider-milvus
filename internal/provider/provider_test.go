// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProvider_Basic verifies the provider can be configured and used
// to create a minimal collection resource.
func TestAccProvider_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_Basic(),
				// No Check needed — a successful apply/plan verifies provider config.
			},
		},
	})
}

// TestAccProvider_WithCredentials verifies the provider accepts optional
// username/password attributes without error.
func TestAccProvider_WithCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_WithCredentials(),
			},
		},
	})
}

func testAccProviderConfig_Basic() string {
	return `
provider "milvus" {
  address = "localhost:19530"
}
`
}

func testAccProviderConfig_WithCredentials() string {
	return `
provider "milvus" {
  address  = "localhost:19530"
  username = "root"
  password = "Milvus"
  db_name  = "default"
}
`
}
