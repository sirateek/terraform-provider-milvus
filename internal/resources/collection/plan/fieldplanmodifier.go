// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package plan

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection/model"
)

// vectorFieldEqual returns true when all structural attributes of two vector
// FieldModels are identical. Only attributes that Milvus stores and returns
// for vector fields are compared; boolean flags that default to false are
// excluded because Milvus does not echo them back and would produce false diffs.
func vectorFieldEqual(a, b model.FieldModel) bool {
	return a.Name.Equal(b.Name) &&
		a.DataType.Equal(b.DataType) &&
		a.Dim.Equal(b.Dim) &&
		a.Description.Equal(b.Description)
}

// scalarFieldEqual returns true when all structural attributes of two scalar
// FieldModels are identical. mmap_enabled is intentionally excluded: it is a
// runtime property settable via AlterCollectionFieldProperty without recreation.
func scalarFieldEqual(a, b model.FieldModel) bool {
	return a.Name.Equal(b.Name) &&
		a.DataType.Equal(b.DataType) &&
		a.Nullable.Equal(b.Nullable) &&
		a.MaxLength.Equal(b.MaxLength) &&
		a.MaxCapacity.Equal(b.MaxCapacity) &&
		a.ElementType.Equal(b.ElementType) &&
		a.Description.Equal(b.Description) &&
		a.IsPrimaryKey.Equal(b.IsPrimaryKey) &&
		a.IsPartitionKey.Equal(b.IsPartitionKey) &&
		a.IsClusteringKey.Equal(b.IsClusteringKey)
}

// FieldsListPlanModifier requires replacement when:
//   - A vector field is added, removed, or any of its attributes changed.
//   - A scalar field is removed or any of its structural attributes changed.
//     (Adding a new scalar field is allowed via AddCollectionField — no replace needed.)
//   - mmap_enabled is excluded from scalar field comparison because it can be
//     updated in-place via AlterCollectionFieldProperty.
type FieldsListPlanModifier struct{}

func (m FieldsListPlanModifier) Description(_ context.Context) string {
	return "Requires replacement if a vector field is added or any existing field is removed/modified."
}

func (m FieldsListPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m FieldsListPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// No state yet (create) or plan is unknown — nothing to do.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.Equal(req.PlanValue) {
		return
	}

	var stateFields, planFields []model.FieldModel
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &stateFields, false)...)
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &planFields, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateFieldsByName := make(map[string]model.FieldModel, len(stateFields))
	for _, f := range stateFields {
		stateFieldsByName[f.Name.ValueString()] = f
	}

	planFieldNames := make(map[string]struct{}, len(planFields))
	for _, f := range planFields {
		name := f.Name.ValueString()
		planFieldNames[name] = struct{}{}

		stateField, exists := stateFieldsByName[name]
		if !exists {
			// New field: require replace if it is a vector type.
			if model.IsVectorFieldType(f.DataType) {
				resp.RequiresReplace = true
				return
			}
			// New scalar field — allowed via AddCollectionField, no replace needed.
			continue
		}

		if model.IsVectorFieldType(f.DataType) {
			// Existing vector field: any attribute change requires replacement.
			if !vectorFieldEqual(f, stateField) {
				resp.RequiresReplace = true
				return
			}
		} else {
			// Existing scalar field: structural attribute change requires replacement.
			// mmap_enabled is intentionally excluded (handled in-place).
			if !scalarFieldEqual(f, stateField) {
				resp.RequiresReplace = true
				return
			}
		}
	}

	// Require replace if any existing field was removed.
	for name := range stateFieldsByName {
		if _, exists := planFieldNames[name]; !exists {
			resp.RequiresReplace = true
			return
		}
	}
}
