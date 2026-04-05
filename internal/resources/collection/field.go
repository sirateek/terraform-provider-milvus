package collection

import "github.com/hashicorp/terraform-plugin-framework/types"

// FieldModel represents a single schema field in Terraform state.
type FieldModel struct {
	Name            types.String `tfsdk:"name"`
	DataType        types.String `tfsdk:"data_type"`
	IsPrimaryKey    types.Bool   `tfsdk:"is_primary_key"`
	IsAutoID        types.Bool   `tfsdk:"is_auto_id"`
	IsPartitionKey  types.Bool   `tfsdk:"is_partition_key"`
	IsClusteringKey types.Bool   `tfsdk:"is_clustering_key"`
	Nullable        types.Bool   `tfsdk:"nullable"`
	Dim             types.Int64  `tfsdk:"dim"`
	MaxLength       types.Int64  `tfsdk:"max_length"`
	MaxCapacity     types.Int64  `tfsdk:"max_capacity"`
	ElementType     types.String `tfsdk:"element_type"`
	Description     types.String `tfsdk:"description"`
}
