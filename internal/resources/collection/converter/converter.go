// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package converter

//go:generate go run github.com/jmattheis/goverter/cmd/goverter gen github.com/sirateek/terraform-provider-milvus/internal/resources/collection/converter

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection/model"
)

// goverter:converter
// goverter:output:file ./converter_generated.go
// goverter:output:package github.com/sirateek/terraform-provider-milvus/internal/resources/collection/converter
// goverter:extend TypesBoolToBoolPtr TypesInt64ToInt64Ptr TypesStringToStringPtr
// goverter:extend BoolPtrToTypesBool Int64PtrToTypesInt64 StringPtrToTypesString
type Converter interface {
	// MilvusCollectionPropertiesToParsed converts Terraform-typed properties to
	// native Go pointer types used for Milvus API calls.
	// Returns nil if in is nil.
	MilvusCollectionPropertiesToParsed(in *model.MilvusCollectionProperties) *model.ParsedMilvusCollectionProperties

	// ParsedToMilvusCollectionProperties converts native Go pointer types back to
	// Terraform-typed properties for state storage.
	// Returns nil if in is nil.
	ParsedToMilvusCollectionProperties(in *model.ParsedMilvusCollectionProperties) *model.MilvusCollectionProperties
}

// TypesBoolToBoolPtr converts types.Bool to *bool.
// Returns nil if the value is null or unknown.
func TypesBoolToBoolPtr(in types.Bool) *bool {
	if in.IsNull() || in.IsUnknown() {
		return nil
	}
	v := in.ValueBool()
	return &v
}

// BoolPtrToTypesBool converts *bool to types.Bool.
// Returns types.BoolNull() if the pointer is nil.
func BoolPtrToTypesBool(in *bool) types.Bool {
	if in == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*in)
}

// TypesInt64ToInt64Ptr converts types.Int64 to *int64.
// Returns nil if the value is null or unknown.
func TypesInt64ToInt64Ptr(in types.Int64) *int64 {
	if in.IsNull() || in.IsUnknown() {
		return nil
	}
	v := in.ValueInt64()
	return &v
}

// Int64PtrToTypesInt64 converts *int64 to types.Int64.
// Returns types.Int64Null() if the pointer is nil.
func Int64PtrToTypesInt64(in *int64) types.Int64 {
	if in == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*in)
}

// TypesStringToStringPtr converts types.String to *string.
// Returns nil if the value is null or unknown.
func TypesStringToStringPtr(in types.String) *string {
	if in.IsNull() || in.IsUnknown() {
		return nil
	}
	v := in.ValueString()
	return &v
}

// StringPtrToTypesString converts *string to types.String.
// Returns types.StringNull() if the pointer is nil.
func StringPtrToTypesString(in *string) types.String {
	if in == nil {
		return types.StringNull()
	}
	return types.StringValue(*in)
}
