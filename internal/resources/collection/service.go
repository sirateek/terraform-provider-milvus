// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package collection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection/converter"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection/model"
	"github.com/sirateek/terraform-provider-milvus/internal/util"
)

// addNewScalarFields calls AddCollectionField for each field present in plan
// but absent from state. The plan modifier guarantees no vector fields reach
// this path — those are handled by RequiresReplace.
func (r *MilvusCollectionResource) addNewScalarFields(
	ctx context.Context,
	plan,
	state model.MilvusCollectionResourceModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	var planFields, stateFields []model.FieldModel
	diags.Append(plan.Fields.ElementsAs(ctx, &planFields, false)...)
	diags.Append(state.Fields.ElementsAs(ctx, &stateFields, false)...)
	if diags.HasError() {
		return diags
	}

	stateFieldNames := make(map[string]struct{}, len(stateFields))
	for _, f := range stateFields {
		stateFieldNames[f.Name.ValueString()] = struct{}{}
	}

	for _, f := range planFields {
		if _, exists := stateFieldNames[f.Name.ValueString()]; exists {
			continue
		}
		field, ok := toEntityField(f)
		if !ok {
			diags.AddError("Fail to convert field to entity.Field", "Please report this issue to the provider developers.g")
			return diags
		}
		opt := milvusclient.NewAddCollectionFieldOption(plan.Name.ValueString(), field)
		if err := r.client.AddCollectionField(ctx, opt); err != nil {
			diags.AddError(
				"Error adding field to collection",
				fmt.Sprintf("Could not add field %q to collection %s: %s", f.Name.ValueString(), plan.Name.ValueString(), err.Error()),
			)
			return diags
		}
	}

	return diags
}

// readCollection fetches the collection from Milvus and updates the model.
func (r *MilvusCollectionResource) readCollection(ctx context.Context, in *model.MilvusCollectionResourceModel, diag diag.Diagnostics) {
	coll, err := r.client.DescribeCollection(ctx, milvusclient.NewDescribeCollectionOption(in.Name.ValueString()))
	if err != nil {
		diag.AddError(
			"Error reading collection",
			fmt.Sprintf("Could not describe collection %s: %s", in.Name.ValueString(), err.Error()),
		)
		return
	}

	// Set the ID to the Milvus-generated collection ID
	in.ID = types.Int64Value(coll.ID)

	// Update computed fields from the collection
	in.Description = types.StringValue(coll.Schema.Description)
	in.AutoID = types.BoolValue(coll.Schema.AutoID)
	in.EnableDynamicField = types.BoolValue(coll.Schema.EnableDynamicField)

	// delete_protection is provider-only and has no Milvus counterpart.
	// During import, the state has null here. Default it to false so that the
	// post-import plan produces no diff when the config also omits or sets false.
	// For normal reads the value is already populated from state, so IsNull()
	// is false and we leave it unchanged.
	if in.DeleteProtection.IsNull() {
		in.DeleteProtection = types.BoolValue(false)
	}

	// Map consistency level
	in.ConsistencyLevel = types.StringValue(consistencyLevelToString(coll.ConsistencyLevel))

	// Map properties - only if they were explicitly set in the plan
	if in.Properties != nil && coll.Properties != nil {
		var collPropertiesBridge model.ParsedMilvusCollectionProperties
		rawPropertyMap, marshalErr := json.Marshal(coll.Properties)
		if marshalErr != nil {
			diag.AddError(
				"Error parsing properties",
				"Could not parse properties: "+marshalErr.Error(),
			)
			return
		}
		unmarshalErr := json.Unmarshal(rawPropertyMap, &collPropertiesBridge)
		if unmarshalErr != nil {
			diag.AddError(
				"Error parsing properties",
				"Could not parse properties: "+unmarshalErr.Error(),
			)
			return
		}
		in.Properties = converter.ConverterInstance.ParsedToMilvusCollectionProperties(&collPropertiesBridge)
	}

	// Build fields list
	var fieldModels []model.FieldModel
	for _, field := range coll.Schema.Fields {
		fieldModel := model.FieldModel{
			Name:            types.StringValue(field.Name),
			DataType:        types.StringValue(model.GetStringValueFromMilvusFieldType(field.DataType)),
			IsPrimaryKey:    types.BoolValue(field.PrimaryKey),
			IsAutoID:        types.BoolValue(field.AutoID),
			IsPartitionKey:  types.BoolValue(field.IsPartitionKey),
			IsClusteringKey: types.BoolValue(field.IsClusteringKey),
			Nullable:        types.BoolValue(field.Nullable),
			Description:     types.StringValue(field.Description),
		}

		// Set optional TypeParam-backed fields (dim, max_length, max_capacity).
		if field.TypeParams != nil {
			if dim, ok := field.TypeParams["dim"]; ok {
				if dimVal, err := util.ToInt64(dim); err == nil {
					fieldModel.Dim = types.Int64Value(dimVal)
				}
			}
			if maxLength, ok := field.TypeParams["max_length"]; ok {
				if maxLengthVal, err := util.ToInt64(maxLength); err == nil {
					fieldModel.MaxLength = types.Int64Value(maxLengthVal)
				}
			}
			if maxCapacity, ok := field.TypeParams["max_capacity"]; ok {
				if maxCapacityVal, err := util.ToInt64(maxCapacity); err == nil {
					fieldModel.MaxCapacity = types.Int64Value(maxCapacityVal)
				}
			}
		}

		// ElementType is a direct field on entity.Field (not a TypeParam),
		// so it must be read outside the TypeParams block.
		if field.ElementType != entity.FieldTypeNone {
			fieldModel.ElementType = types.StringValue(model.GetStringValueFromMilvusFieldType(field.ElementType))
		}

		fieldModels = append(fieldModels, fieldModel)
	}

	// Convert fields to types.List
	if len(fieldModels) > 0 {
		fieldList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: fieldObjAttrTypes()}, fieldModels)
		diag.Append(diags...)
		in.Fields = fieldList
	}

	// Set other computed values
	in.ShardNum = types.Int64Value(int64(coll.ShardNum))
}
