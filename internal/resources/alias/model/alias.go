package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type Alias struct {
	Name           types.String `tfsdk:"name"`
	CollectionName types.String `tfsdk:"collection_name"`
}
