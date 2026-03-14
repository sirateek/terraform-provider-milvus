package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// CollectionFieldModel represents a single schema field in Terraform state
type CollectionFieldModel struct {
	Name           types.String `tfsdk:"name"`
	DataType       types.String `tfsdk:"data_type"`
	IsPrimaryKey   types.Bool   `tfsdk:"is_primary_key"`
	IsAutoID       types.Bool   `tfsdk:"is_auto_id"`
	IsPartitionKey types.Bool   `tfsdk:"is_partition_key"`
	IsClusteringKey types.Bool   `tfsdk:"is_clustering_key"`
	Nullable       types.Bool   `tfsdk:"nullable"`
	Dim            types.Int64  `tfsdk:"dim"`
	MaxLength      types.Int64  `tfsdk:"max_length"`
	MaxCapacity    types.Int64  `tfsdk:"max_capacity"`
	ElementType    types.String `tfsdk:"element_type"`
	Description    types.String `tfsdk:"description"`
}

// MilvusCollectionResourceModel is the top-level Terraform state model
type MilvusCollectionResourceModel struct {
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	AutoID             types.Bool   `tfsdk:"auto_id"`
	EnableDynamicField types.Bool   `tfsdk:"enable_dynamic_field"`
	ShardNum           types.Int64  `tfsdk:"shard_num"`
	ConsistencyLevel   types.String `tfsdk:"consistency_level"`
	Properties         types.Map    `tfsdk:"properties"`
	Fields             types.List   `tfsdk:"fields"`
}

// MilvusCollectionResource is the resource implementation
type MilvusCollectionResource struct {
	client *milvusclient.Client
}

var _ resource.Resource = &MilvusCollectionResource{}
var _ resource.ResourceWithConfigure = &MilvusCollectionResource{}
var _ resource.ResourceWithImportState = &MilvusCollectionResource{}

// NewMilvusCollectionResource is a helper function to simplify the provider implementation.
func NewMilvusCollectionResource() resource.Resource {
	return &MilvusCollectionResource{}
}

// Metadata returns the resource type name.
func (r *MilvusCollectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection"
}

// Schema defines the schema for the resource.
func (r *MilvusCollectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Milvus collection.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the collection. Required and immutable.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the collection. Optional and immutable.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_id": schema.BoolAttribute{
				MarkdownDescription: "Whether to automatically generate IDs for primary key field. Optional and immutable.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_dynamic_field": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable dynamic field. Optional and immutable.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"shard_num": schema.Int64Attribute{
				MarkdownDescription: "Number of shards for the collection. Optional and immutable after creation.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"consistency_level": schema.StringAttribute{
				MarkdownDescription: "Consistency level of the collection. Valid values: Strong, Bounded, Session, Eventually. Optional and mutable.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Strong"),
			},
			"properties": schema.MapAttribute{
				MarkdownDescription: "Additional properties for the collection. Optional and mutable.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"fields": schema.ListNestedAttribute{
				MarkdownDescription: "Schema fields for the collection. Required and immutable.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Field name. Required.",
							Required:            true,
						},
						"data_type": schema.StringAttribute{
							MarkdownDescription: "Field data type (e.g., Int64, FloatVector, VarChar). Required.",
							Required:            true,
						},
						"is_primary_key": schema.BoolAttribute{
							MarkdownDescription: "Whether this field is the primary key. Optional and computed.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"is_auto_id": schema.BoolAttribute{
							MarkdownDescription: "Whether this field auto-generates IDs. Optional and computed.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"is_partition_key": schema.BoolAttribute{
							MarkdownDescription: "Whether this field is a partition key. Optional and computed.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"is_clustering_key": schema.BoolAttribute{
							MarkdownDescription: "Whether this field is a clustering key. Optional and computed.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"nullable": schema.BoolAttribute{
							MarkdownDescription: "Whether this field is nullable. Optional and computed.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"dim": schema.Int64Attribute{
							MarkdownDescription: "Dimension for vector fields. Optional.",
							Optional:            true,
						},
						"max_length": schema.Int64Attribute{
							MarkdownDescription: "Maximum length for varchar fields. Optional.",
							Optional:            true,
						},
						"max_capacity": schema.Int64Attribute{
							MarkdownDescription: "Maximum capacity for array fields. Optional.",
							Optional:            true,
						},
						"element_type": schema.StringAttribute{
							MarkdownDescription: "Element type for array fields. Optional.",
							Optional:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the field. Optional and computed.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *MilvusCollectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*milvusclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *milvusclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *MilvusCollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MilvusCollectionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract fields from plan
	var fields []CollectionFieldModel
	resp.Diagnostics.Append(plan.Fields.ElementsAs(ctx, &fields, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build entity schema
	collSchema := toEntitySchema(ctx, plan, fields)

	// Build create collection option
	opt := milvusclient.NewCreateCollectionOption(plan.Name.ValueString(), collSchema)

	// Set optional fields
	if !plan.ShardNum.IsNull() {
		opt.WithShardNum(int32(plan.ShardNum.ValueInt64()))
	}

	if !plan.ConsistencyLevel.IsNull() {
		consistencyLevel := consistencyLevelFromString(plan.ConsistencyLevel.ValueString())
		opt.WithConsistencyLevel(consistencyLevel)
	}

	// Handle properties
	if !plan.Properties.IsNull() {
		props := make(map[string]string)
		resp.Diagnostics.Append(plan.Properties.ElementsAs(ctx, &props, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range props {
			opt.WithProperty(k, v)
		}
	}

	// Create the collection
	err := r.client.CreateCollection(ctx, opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating collection",
			fmt.Sprintf("Could not create collection %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Read back the created collection to populate computed values
	r.readCollection(ctx, &plan, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *MilvusCollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MilvusCollectionResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if collection exists
	exists, err := r.client.HasCollection(ctx, milvusclient.NewHasCollectionOption(state.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error checking collection",
			fmt.Sprintf("Could not check if collection %s exists: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}

	r.readCollection(ctx, &state, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *MilvusCollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state MilvusCollectionResourceModel

	// Read Terraform plan and state data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only consistency_level and properties can be updated
	// Other fields have RequiresReplace, so if they changed, Terraform would replace the resource

	// Update properties if changed
	if !plan.Properties.Equal(state.Properties) {
		planProps := make(map[string]string)
		stateProps := make(map[string]string)

		resp.Diagnostics.Append(plan.Properties.ElementsAs(ctx, &planProps, false)...)
		resp.Diagnostics.Append(state.Properties.ElementsAs(ctx, &stateProps, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Find changed properties
		changedProps := make(map[string]string)
		for key, val := range planProps {
			if stateProps[key] != val {
				changedProps[key] = val
			}
		}

		if len(changedProps) > 0 {
			opt := milvusclient.NewAlterCollectionPropertiesOption(plan.Name.ValueString())
			for key, val := range changedProps {
				opt.WithProperty(key, val)
			}

			err := r.client.AlterCollectionProperties(ctx, opt)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating collection properties",
					fmt.Sprintf("Could not update properties for collection %s: %s", plan.Name.ValueString(), err.Error()),
				)
				return
			}
		}
	}

	// Read back the updated collection
	r.readCollection(ctx, &plan, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *MilvusCollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MilvusCollectionResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the collection
	opt := milvusclient.NewDropCollectionOption(state.Name.ValueString())
	err := r.client.DropCollection(ctx, opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting collection",
			fmt.Sprintf("Could not delete collection %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource by name.
func (r *MilvusCollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// Helper Functions

// readCollection fetches the collection from Milvus and updates the model
func (r *MilvusCollectionResource) readCollection(ctx context.Context, model *MilvusCollectionResourceModel, diag diag.Diagnostics) {
	coll, err := r.client.DescribeCollection(ctx, milvusclient.NewDescribeCollectionOption(model.Name.ValueString()))
	if err != nil {
		diag.AddError(
			"Error reading collection",
			fmt.Sprintf("Could not describe collection %s: %s", model.Name.ValueString(), err.Error()),
		)
		return
	}

	// Update computed fields from the collection
	model.Description = types.StringValue(coll.Schema.Description)

	// Map consistency level
	model.ConsistencyLevel = types.StringValue(consistencyLevelToString(coll.ConsistencyLevel))

	// Map properties
	if coll.Properties != nil && len(coll.Properties) > 0 {
		propsValue, diags := types.MapValueFrom(ctx, types.StringType, coll.Properties)
		diag.Append(diags...)
		model.Properties = propsValue
	} else {
		model.Properties = types.MapNull(types.StringType)
	}

	// Build fields list
	var fieldModels []CollectionFieldModel
	for _, field := range coll.Schema.Fields {
		fieldModel := CollectionFieldModel{
			Name:           types.StringValue(field.Name),
			DataType:       types.StringValue(fieldTypeToString(field.DataType)),
			IsPrimaryKey:   types.BoolValue(field.PrimaryKey),
			IsAutoID:       types.BoolValue(field.AutoID),
			IsPartitionKey: types.BoolValue(field.IsPartitionKey),
			IsClusteringKey: types.BoolValue(field.IsClusteringKey),
			Nullable:       types.BoolValue(field.Nullable),
			Description:    types.StringValue(field.Description),
		}

		// Set optional integer fields
		if field.TypeParams != nil {
			if dim, ok := field.TypeParams["dim"]; ok {
				if dimVal, err := toInt64(dim); err == nil {
					fieldModel.Dim = types.Int64Value(dimVal)
				}
			}
			if maxLength, ok := field.TypeParams["max_length"]; ok {
				if maxLengthVal, err := toInt64(maxLength); err == nil {
					fieldModel.MaxLength = types.Int64Value(maxLengthVal)
				}
			}
			if maxCapacity, ok := field.TypeParams["max_capacity"]; ok {
				if maxCapacityVal, err := toInt64(maxCapacity); err == nil {
					fieldModel.MaxCapacity = types.Int64Value(maxCapacityVal)
				}
			}
			if elementType, ok := field.TypeParams["element_type"]; ok {
				fieldModel.ElementType = types.StringValue(fmt.Sprintf("%v", elementType))
			}
		}

		fieldModels = append(fieldModels, fieldModel)
	}

	// Convert fields to types.List
	if len(fieldModels) > 0 {
		fieldList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: fieldObjAttrTypes()}, fieldModels)
		diag.Append(diags...)
		model.Fields = fieldList
	}

	// Set other computed values
	model.ShardNum = types.Int64Value(int64(coll.ShardNum))
}

// toEntitySchema converts Terraform plan to entity.Schema
func toEntitySchema(_ context.Context, plan MilvusCollectionResourceModel, fields []CollectionFieldModel) *entity.Schema {
	collSchema := &entity.Schema{
		CollectionName:     plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		AutoID:             plan.AutoID.ValueBool(),
		EnableDynamicField: plan.EnableDynamicField.ValueBool(),
	}

	// Add fields
	for _, f := range fields {
		field := toEntityField(f)
		if field != nil {
			collSchema.Fields = append(collSchema.Fields, field)
		}
	}

	return collSchema
}

// toEntityField converts a CollectionFieldModel to entity.Field
func toEntityField(f CollectionFieldModel) *entity.Field {
	field := &entity.Field{
		Name:        f.Name.ValueString(),
		DataType:    fieldTypeFromString(f.DataType.ValueString()),
		PrimaryKey:  f.IsPrimaryKey.ValueBool(),
		AutoID:      f.IsAutoID.ValueBool(),
		Nullable:    f.Nullable.ValueBool(),
		Description: f.Description.ValueString(),
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
		typeParams["element_type"] = f.ElementType.ValueString()
	}

	if len(typeParams) > 0 {
		field.TypeParams = typeParams
	}

	return field
}

// fieldTypeFromString maps string to entity.FieldType
func fieldTypeFromString(s string) entity.FieldType {
	switch s {
	case "Float":
		return entity.FieldTypeFloat
	case "FloatVector":
		return entity.FieldTypeFloatVector
	case "BinaryVector":
		return entity.FieldTypeBinaryVector
	case "BFloat16Vector":
		return entity.FieldTypeBFloat16Vector
	case "Float16Vector":
		return entity.FieldTypeFloat16Vector
	case "Int8":
		return entity.FieldTypeInt8
	case "Int16":
		return entity.FieldTypeInt16
	case "Int32":
		return entity.FieldTypeInt32
	case "Int64":
		return entity.FieldTypeInt64
	case "Bool":
		return entity.FieldTypeBool
	case "VarChar":
		return entity.FieldTypeVarChar
	case "String":
		return entity.FieldTypeString
	case "Array":
		return entity.FieldTypeArray
	case "JSON":
		return entity.FieldTypeJSON
	default:
		return entity.FieldTypeFloat
	}
}

// fieldTypeToString maps entity.FieldType to string
func fieldTypeToString(ft entity.FieldType) string {
	switch ft {
	case entity.FieldTypeFloat:
		return "Float"
	case entity.FieldTypeFloatVector:
		return "FloatVector"
	case entity.FieldTypeBinaryVector:
		return "BinaryVector"
	case entity.FieldTypeBFloat16Vector:
		return "BFloat16Vector"
	case entity.FieldTypeFloat16Vector:
		return "Float16Vector"
	case entity.FieldTypeInt8:
		return "Int8"
	case entity.FieldTypeInt16:
		return "Int16"
	case entity.FieldTypeInt32:
		return "Int32"
	case entity.FieldTypeInt64:
		return "Int64"
	case entity.FieldTypeBool:
		return "Bool"
	case entity.FieldTypeVarChar:
		return "VarChar"
	case entity.FieldTypeString:
		return "String"
	case entity.FieldTypeArray:
		return "Array"
	case entity.FieldTypeJSON:
		return "JSON"
	default:
		return "Float"
	}
}

// consistencyLevelFromString maps string to entity.ConsistencyLevel
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

// consistencyLevelToString maps entity.ConsistencyLevel to string
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

// fieldObjAttrTypes returns the attribute type map for CollectionFieldModel
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

// toInt64 is a helper to convert interface{} to int64
func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case string:
		var i int64
		_, err := fmt.Sscanf(val, "%d", &i)
		return i, err
	default:
		return 0, fmt.Errorf("cannot convert %v to int64", v)
	}
}
