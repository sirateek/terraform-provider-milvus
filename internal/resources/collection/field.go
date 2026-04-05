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

type FieldDataType types.String

// Supported Vector Field Data Type
var (
	FloatVectorFieldDataType    = types.StringValue("FloatVector")
	BinaryVectorFieldDataType   = types.StringValue("BinaryVector")
	Float16VectorFieldDataType  = types.StringValue("Float16Vector")
	BFloat16VectorFieldDataType = types.StringValue("BFloat16Vector")
	SparseVectorFieldDataType   = types.StringValue("SparseVector")
	Int8VectorFieldDataType     = types.StringValue("Int8Vector")
)

// isVectorFieldType returns true for vector data types that cannot be added
// to an existing collection without recreation.
func isVectorFieldType(dataType types.String) bool {
	switch dataType.ValueString() {
	case FloatVectorFieldDataType.ValueString(),
		BinaryVectorFieldDataType.ValueString(),
		Float16VectorFieldDataType.ValueString(),
		BFloat16VectorFieldDataType.ValueString(),
		SparseVectorFieldDataType.ValueString(),
		Int8VectorFieldDataType.ValueString():
		return true
	}
	return false
}
