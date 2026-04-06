// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/milvus-io/milvus/client/v2/entity"
)

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

// FieldTypeStringMap represent the mapping between string input and the Milvus's GRPC enum.
type FieldTypeStringMap struct {
	stringValue     string
	milvusFieldType entity.FieldType
	isVector        bool
}

func (f FieldTypeStringMap) IsMatchStringValue(s string) bool {
	return f.stringValue == s
}

func (f FieldTypeStringMap) IsVectorField() bool {
	return f.isVector
}

var (
	supportedFieldType = []FieldTypeStringMap{
		{
			stringValue:     "Float",
			milvusFieldType: entity.FieldTypeFloat,
			isVector:        false,
		},
		{
			stringValue:     "FloatVector",
			milvusFieldType: entity.FieldTypeFloatVector,
			isVector:        true,
		},
		{
			stringValue:     "BinaryVector",
			milvusFieldType: entity.FieldTypeBinaryVector,
			isVector:        true,
		},
		{
			stringValue:     "BFloat16Vector",
			milvusFieldType: entity.FieldTypeBFloat16Vector,
			isVector:        true,
		},
		{
			stringValue:     "Float16Vector",
			milvusFieldType: entity.FieldTypeFloat16Vector,
			isVector:        true,
		},
		{
			stringValue:     "Int8",
			milvusFieldType: entity.FieldTypeInt8,
		},
		{
			stringValue:     "Int16",
			milvusFieldType: entity.FieldTypeInt16,
		},
		{
			stringValue:     "Int32",
			milvusFieldType: entity.FieldTypeInt32,
		},
		{
			stringValue:     "Int64",
			milvusFieldType: entity.FieldTypeInt64,
		},
		{
			stringValue:     "Bool",
			milvusFieldType: entity.FieldTypeBool,
		},
		{
			stringValue:     "VarChar",
			milvusFieldType: entity.FieldTypeVarChar,
		},
		{
			stringValue:     "String",
			milvusFieldType: entity.FieldTypeString,
		},
		{
			stringValue:     "Array",
			milvusFieldType: entity.FieldTypeArray,
		},
		{
			stringValue:     "JSON",
			milvusFieldType: entity.FieldTypeJSON,
		},
	}
)

// IsVectorFieldType returns true if the given types.String data type value
// represents a vector field type that cannot be added to an existing collection
// without recreation.
func IsVectorFieldType(dataType types.String) bool {
	if dataType.IsNull() || dataType.IsUnknown() {
		return false
	}
	for _, v := range supportedFieldType {
		if v.IsMatchStringValue(dataType.ValueString()) {
			return v.IsVectorField()
		}
	}
	return false
}

func GetMilvusFieldTypeFromString(s string) (entity.FieldType, bool) {
	for _, v := range supportedFieldType {
		if v.IsMatchStringValue(s) {
			return v.milvusFieldType, true
		}
	}
	return entity.FieldTypeNone, false
}

func GetStringValueFromMilvusFieldType(ft entity.FieldType) string {
	for _, v := range supportedFieldType {
		if v.IsVectorField() && v.milvusFieldType == ft {
			return v.stringValue
		}
		if !v.IsVectorField() && v.milvusFieldType == ft {
			return v.stringValue
		}
	}
	return ""
}
