package collection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// fieldsListPlanModifier requires replacement when:
//   - Milvus only support adding a new scalar field to the existing collection.
//   - Milvus do not support adding / removing / modifying any Vector field. Doing so will result in collection recreation.
//
// Adding scalar fields is allowed without replacement via AddCollectionField.
type fieldsListPlanModifier struct{}

func (m fieldsListPlanModifier) Description(_ context.Context) string {
	return "Requires replacement if a vector field is added or any existing field is removed/modified."
}

func (m fieldsListPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m fieldsListPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
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

	var stateFields, planFields []FieldModel
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &stateFields, false)...)
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &planFields, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateFieldNames := make(map[string]struct{}, len(stateFields))
	for _, f := range stateFields {
		stateFieldNames[f.Name.ValueString()] = struct{}{}
	}

	planFieldNames := make(map[string]struct{}, len(planFields))
	for _, f := range planFields {
		name := f.Name.ValueString()
		planFieldNames[name] = struct{}{}

		if _, exists := stateFieldNames[name]; !exists {
			// New field: require replace if it is a vector type.
			if isVectorFieldType(f.DataType) {
				resp.RequiresReplace = true
				return
			}
		}
	}

	// Require replace if any existing field was removed.
	for name := range stateFieldNames {
		if _, exists := planFieldNames[name]; !exists {
			resp.RequiresReplace = true
			return
		}
	}
}
