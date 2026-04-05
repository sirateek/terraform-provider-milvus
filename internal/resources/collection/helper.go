// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package collection

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection/model"
)

// toEntitySchema converts Terraform plan to entity.Schema.
func toEntitySchema(_ context.Context, resourceModel model.MilvusCollectionResourceModel, fields []model.FieldModel) (*entity.Schema, bool) {
	collSchema := &entity.Schema{
		CollectionName:     resourceModel.Name.ValueString(),
		Description:        resourceModel.Description.ValueString(),
		AutoID:             resourceModel.AutoID.ValueBool(),
		EnableDynamicField: resourceModel.EnableDynamicField.ValueBool(),
	}

	// Add fields
	for _, f := range fields {
		field, ok := toEntityField(f)
		if !ok {
			return nil, false
		}
		if field != nil {
			collSchema.Fields = append(collSchema.Fields, field)
		}
	}

	return collSchema, true
}

// toEntityField converts a FieldModel to entity.Field.
func toEntityField(f model.FieldModel) (*entity.Field, bool) {
	dataType, ok := model.GetMilvusFieldTypeFromString(f.DataType.ValueString())
	if !ok {
		return nil, false
	}
	field := &entity.Field{
		Name:            f.Name.ValueString(),
		DataType:        dataType,
		PrimaryKey:      f.IsPrimaryKey.ValueBool(),
		AutoID:          f.IsAutoID.ValueBool(),
		IsPartitionKey:  f.IsPartitionKey.ValueBool(),
		IsClusteringKey: f.IsClusteringKey.ValueBool(),
		Nullable:        f.Nullable.ValueBool(),
		Description:     f.Description.ValueString(),
	}

	// Set type parameters
	typeParams := make(map[string]string)

	if !f.Dim.IsNull() {
		typeParams["dim"] = fmt.Sprintf("%d", f.Dim.ValueInt64())
	}

	if !f.MaxLength.IsNull() {
		typeParams["max_length"] = fmt.Sprintf("%d", f.MaxLength.ValueInt64())
	}

	if !f.MaxCapacity.IsNull() {
		typeParams["max_capacity"] = fmt.Sprintf("%d", f.MaxCapacity.ValueInt64())
	}

	if !f.ElementType.IsNull() {
		elementType, fieldTypeFound := model.GetMilvusFieldTypeFromString(f.ElementType.ValueString())
		if !fieldTypeFound {
			return nil, false
		}
		field.ElementType = elementType
	}

	if len(typeParams) > 0 {
		field.TypeParams = typeParams
	}

	return field, true
}

// consistencyLevelFromString maps string to entity.ConsistencyLevel.
func consistencyLevelFromString(s string) entity.ConsistencyLevel {
	switch s {
	case "Strong":
		return entity.ClStrong
	case "Bounded":
		return entity.ClBounded
	case "Session":
		return entity.ClSession
	case "Eventually":
		return entity.ClEventually
	default:
		return entity.ClStrong
	}
}

// consistencyLevelToString maps entity.ConsistencyLevel to string.
func consistencyLevelToString(cl entity.ConsistencyLevel) string {
	switch cl {
	case entity.ClStrong:
		return "Strong"
	case entity.ClBounded:
		return "Bounded"
	case entity.ClSession:
		return "Session"
	case entity.ClEventually:
		return "Eventually"
	default:
		return "Strong"
	}
}

// fieldObjAttrTypes returns the attribute type map for FieldModel.
func fieldObjAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":              types.StringType,
		"data_type":         types.StringType,
		"is_primary_key":    types.BoolType,
		"is_auto_id":        types.BoolType,
		"is_partition_key":  types.BoolType,
		"is_clustering_key": types.BoolType,
		"nullable":          types.BoolType,
		"dim":               types.Int64Type,
		"max_length":        types.Int64Type,
		"max_capacity":      types.Int64Type,
		"element_type":      types.StringType,
		"description":       types.StringType,
	}
}

// compareTerraformCollectionPropertyPlanAndState returns a MilvusCollectionProperties
// containing only the fields that differ between plan and state.
// Unchanged fields are left as their zero values, which are null in the framework types.
// New fields added to MilvusCollectionProperties in the future are automatically included
// without any changes to this function.
func compareTerraformCollectionPropertyPlanAndState(plan, state *model.MilvusCollectionProperties) *model.MilvusCollectionProperties {
	var result model.MilvusCollectionProperties

	planVal := reflect.ValueOf(plan)
	stateVal := reflect.ValueOf(state)
	resultVal := reflect.ValueOf(&result).Elem()

	for i := range planVal.NumField() {
		planAttr, ok := planVal.Field(i).Interface().(attr.Value)
		if !ok {
			continue
		}
		stateAttr, ok := stateVal.Field(i).Interface().(attr.Value)
		if !ok {
			continue
		}

		if !planAttr.Equal(stateAttr) {
			resultVal.Field(i).Set(planVal.Field(i))
		}
	}

	return &result
}
