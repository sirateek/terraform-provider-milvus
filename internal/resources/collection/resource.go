// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package collection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/milvus-io/milvus/pkg/v2/util/merr"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection/converter"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection/model"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection/plan"
)

var _ resource.Resource = &MilvusCollectionResource{}
var _ resource.ResourceWithConfigure = &MilvusCollectionResource{}
var _ resource.ResourceWithImportState = &MilvusCollectionResource{}

// MilvusCollectionResource is the resource implementation.
type MilvusCollectionResource struct {
	client *milvusclient.Client
}

// NewMilvusCollectionResource is a helper function to simplify the provider implementation.
func NewMilvusCollectionResource() resource.Resource {
	return &MilvusCollectionResource{}
}

// Metadata returns the resource type name.
func (r *MilvusCollectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection"
}

// Schema defines the schema for the resource.
func (r *MilvusCollectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Milvus **collection** — the top-level container for vector and scalar data, analogous to a table in a relational database.\n\nA collection has a fixed schema (fields) that is defined at creation time. Most schema properties are **immutable** and require the collection to be recreated if changed. The `consistency_level` and `properties` block can be updated in-place.\n\n~> **Delete protection** is enabled by default (`delete_protection = true`). Set it to `false` before destroying a collection to prevent accidental data loss.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Milvus-generated numeric identifier for the collection. Read-only.",
				Computed:            true,
			},
			"delete_protection": schema.BoolAttribute{
				MarkdownDescription: "When `true` (default), the provider will refuse to destroy this collection, protecting against accidental deletion. Set to `false` to allow `terraform destroy` or removal from configuration. Note that some Milvus operations (e.g. schema changes that require recreation) also require this to be `false`.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the collection within the database. **Immutable** — changing this forces a new collection to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Human-readable description of the collection. **Immutable** — changing this forces a new collection to be created. Defaults to an empty string.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_id": schema.BoolAttribute{
				MarkdownDescription: "When `true`, Milvus automatically generates a unique ID for each inserted entity, so the primary key field does not need to be supplied by the client. **Immutable** — changing this forces a new collection to be created. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_dynamic_field": schema.BoolAttribute{
				MarkdownDescription: "When `true`, entities may contain fields not defined in the schema. Extra fields are stored in a reserved JSON column (`$meta`), enabling schema-on-write flexibility. **Immutable** — changing this forces a new collection to be created. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"shard_num": schema.Int64Attribute{
				MarkdownDescription: "Number of shards (write channels) for the collection. Higher values increase write throughput but also resource usage. **Immutable** — changing this forces a new collection to be created. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"consistency_level": schema.StringAttribute{
				MarkdownDescription: "Read consistency guarantee for the collection. Accepted values:\n\n| Value | Behaviour |\n|---|---|\n| `Strong` | Always reads the latest data. Highest consistency, highest latency. |\n| `Bounded` | Reads data that may be slightly behind the latest write. Good balance. |\n| `Session` | Guarantees that a client sees its own writes. |\n| `Eventually` | No consistency guarantee. Lowest latency. |\n\nCan be updated in-place. Defaults to `Strong`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Strong"),
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Optional runtime properties for the collection. All properties in this block can be updated in-place without recreating the collection.",
				Attributes: map[string]schema.Attribute{
					"collection_ttl_seconds": schema.Int64Attribute{
						MarkdownDescription: "Time-To-Live for collection data in seconds. Entities older than this value are automatically expired and removed by Milvus. The deletion runs asynchronously — queries may still return expired data briefly after the TTL elapses. Omit or set to `0` to disable TTL.",
						Optional:            true,
					},
					"mmap_enabled": schema.BoolAttribute{
						MarkdownDescription: "When `true`, Milvus uses memory-mapped I/O (mmap) for indexes and raw data, allowing the working set to exceed available RAM by paging from disk on demand. Useful for large collections where cost of RAM is a constraint. May increase tail latency under memory pressure.",
						Optional:            true,
					},
					"partition_key_isolation": schema.BoolAttribute{
						MarkdownDescription: "When `true`, Milvus builds a separate index per distinct value of the partition key field. Search requests that filter on the partition key will only scan the relevant sub-index, significantly reducing unnecessary data access. Requires a field with `is_partition_key = true`.",
						Optional:            true,
					},
					"dynamic_field_enabled": schema.BoolAttribute{
						MarkdownDescription: "Overrides the dynamic field setting at the properties level. Prefer the top-level `enable_dynamic_field` attribute for new collections.",
						Optional:            true,
					},
					"allow_insert_auto_id": schema.BoolAttribute{
						MarkdownDescription: "When `true`, allows inserting entities without providing the primary key value even if `auto_id` was not enabled at collection creation. Use with caution — only effective on supported Milvus versions.",
						Optional:            true,
					},
					"allow_update_auto_id": schema.BoolAttribute{
						MarkdownDescription: "When `true`, allows updating the auto-generated primary key. Only applicable when `auto_id = true`.",
						Optional:            true,
					},
					"timezone": schema.StringAttribute{
						MarkdownDescription: "IANA timezone identifier (e.g. `America/New_York`) used for time-based TTL calculations. Defaults to UTC when not set.",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"fields": schema.ListNestedAttribute{
				MarkdownDescription: "Ordered list of fields that define the collection schema. At least one primary key field and one vector field are required. Scalar fields can be added after creation; vector fields and the primary key are **immutable**.\n\n~> Removing or changing an existing field's type forces the entire collection to be recreated.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the field. Must be unique within the collection.",
							Required:            true,
						},
						"data_type": schema.StringAttribute{
							MarkdownDescription: "Data type of the field. Supported scalar types: `Bool`, `Int8`, `Int16`, `Int32`, `Int64`, `Float`, `Double`, `VarChar`, `JSON`, `Array`. Supported vector types: `FloatVector`, `BinaryVector`, `Float16Vector`, `BFloat16Vector`, `SparseFloatVector`.",
							Required:            true,
						},
						"is_primary_key": schema.BoolAttribute{
							MarkdownDescription: "Marks this field as the primary key. Exactly one field in the schema must have this set to `true`. Defaults to `false`.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"is_auto_id": schema.BoolAttribute{
							MarkdownDescription: "When `true` on the primary key field, Milvus auto-generates IDs on insert. Equivalent to setting `auto_id = true` at the collection level. Defaults to `false`.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"is_partition_key": schema.BoolAttribute{
							MarkdownDescription: "Designates this field as the partition key. Milvus physically routes entities by the hash of this field's value. Only one field may be the partition key. Defaults to `false`.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"is_clustering_key": schema.BoolAttribute{
							MarkdownDescription: "Designates this field as the clustering key. Milvus uses this to co-locate entities with similar key values, improving range-query performance. Only one field may be the clustering key. Defaults to `false`.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"nullable": schema.BoolAttribute{
							MarkdownDescription: "When `true`, this field accepts `null` values on insert. Not applicable to primary key or vector fields. Defaults to `false`.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"dim": schema.Int64Attribute{
							MarkdownDescription: "Dimensionality of the vector field (number of floats/bytes per vector). Required for all vector types (`FloatVector`, `BinaryVector`, `Float16Vector`, `BFloat16Vector`). For `BinaryVector`, the value must be a multiple of 8.",
							Optional:            true,
						},
						"max_length": schema.Int64Attribute{
							MarkdownDescription: "Maximum byte length for `VarChar` fields. Required when `data_type = \"VarChar\"`. Maximum allowed value is `65535`.",
							Optional:            true,
						},
						"max_capacity": schema.Int64Attribute{
							MarkdownDescription: "Maximum number of elements in an `Array` field. Required when `data_type = \"Array\"`.",
							Optional:            true,
						},
						"element_type": schema.StringAttribute{
							MarkdownDescription: "Scalar type of each element in an `Array` field (e.g. `Int64`, `VarChar`). Required when `data_type = \"Array\"`.",
							Optional:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Human-readable description of the field. Optional. Defaults to an empty string.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					plan.FieldsListPlanModifier{},
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
	var resourceModel model.MilvusCollectionResourceModel

	// Read Terraform resourceModel data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &resourceModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract fields from resourceModel
	var fields []model.FieldModel
	resp.Diagnostics.Append(resourceModel.Fields.ElementsAs(ctx, &fields, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build entity schema
	collSchema, ok := toEntitySchema(ctx, resourceModel, fields)
	if !ok {
		resp.Diagnostics.AddError("Fail to build entity schema", "Please check the fields in the plan.")
		return
	}

	// Build create collection option
	opt := milvusclient.NewCreateCollectionOption(resourceModel.Name.ValueString(), collSchema)

	// Set optional fields
	if !resourceModel.ShardNum.IsNull() {
		opt.WithShardNum(int32(resourceModel.ShardNum.ValueInt64()))
	}

	if !resourceModel.ConsistencyLevel.IsNull() {
		consistencyLevel := consistencyLevelFromString(resourceModel.ConsistencyLevel.ValueString())
		opt.WithConsistencyLevel(consistencyLevel)
	}

	// Handle properties
	if resourceModel.Properties != nil {
		// Convert to Parsed
		parsedMilvusCollectionProperties := converter.
			ConverterInstance.
			MilvusCollectionPropertiesToParsed(resourceModel.Properties)
		marshaledProperties, err := json.Marshal(parsedMilvusCollectionProperties)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing properties",
				fmt.Sprintf("Could not parse properties: %s", err.Error()),
			)
			return
		}

		var parsedData map[string]any
		err = json.Unmarshal(marshaledProperties, &parsedData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing properties",
				fmt.Sprintf("Could not parse properties: %s", err.Error()),
			)
			return
		}

		for key, value := range parsedData {
			opt.WithProperty(key, value)
		}
	}

	// Create the collection
	err := r.client.CreateCollection(ctx, opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating collection",
			fmt.Sprintf("Could not create collection %s: %s", resourceModel.Name.ValueString(), err.Error()),
		)
		return
	}

	// Read back the created collection to populate computed values
	r.readCollection(ctx, &resourceModel, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &resourceModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *MilvusCollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.MilvusCollectionResourceModel

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

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *MilvusCollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData, stateData model.MilvusCollectionResourceModel

	// Read Terraform planData and stateData data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add new scalar fields if any were added to the planData
	resp.Diagnostics.Append(r.addNewScalarFields(ctx, planData, stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update consistency_level if changed. Milvus accepts it as a property with
	// the integer enum value (Strong=0, Session=1, Bounded=2, Eventually=3).
	if planData.ConsistencyLevel.ValueString() != stateData.ConsistencyLevel.ValueString() {
		cl := consistencyLevelFromString(planData.ConsistencyLevel.ValueString())
		opt := milvusclient.NewAlterCollectionPropertiesOption(planData.Name.ValueString()).
			WithProperty("consistency_level", int(cl))
		if err := r.client.AlterCollectionProperties(ctx, opt); err != nil {
			resp.Diagnostics.AddError(
				"Error updating collection consistency level",
				fmt.Sprintf("Could not update consistency_level for collection %s: %s", planData.Name.ValueString(), err.Error()),
			)
			return
		}
	}

	// Update properties if changed
	changedProps := make(map[string]any)
	if planData.Properties != nil || stateData.Properties != nil {
		diffProperties := compareTerraformCollectionPropertyPlanAndState(planData.Properties, stateData.Properties)
		milvusCollectionPropertyBridge := converter.ConverterInstance.MilvusCollectionPropertiesToParsed(diffProperties)
		marshaledProperties, err := json.Marshal(milvusCollectionPropertyBridge)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing properties",
				fmt.Sprintf("Could not parse properties: %s", err.Error()),
			)
			return
		}

		unmarshalErr := json.Unmarshal(marshaledProperties, &changedProps)
		if unmarshalErr != nil {
			resp.Diagnostics.AddError(
				"Error parsing properties",
				fmt.Sprintf("Could not parse properties: %s", unmarshalErr.Error()),
			)
			return
		}
	}

	if len(changedProps) > 0 {
		opt := milvusclient.NewAlterCollectionPropertiesOption(planData.Name.ValueString())
		for key, val := range changedProps {
			opt.WithProperty(key, val)
		}

		err := r.client.AlterCollectionProperties(ctx, opt)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating collection properties",
				fmt.Sprintf("Could not update properties for collection %s: %s", planData.Name.ValueString(), err.Error()),
			)
			return
		}
	}

	// Read back the updated collection
	r.readCollection(ctx, &planData, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set stateData
	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *MilvusCollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.MilvusCollectionResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeleteProtection.ValueBool() {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Collection %s is protected from deletion",
				state.Name.ValueString()), "Cannot delete the resource that has `delete_protection` on. Please confirm your action by first updating the `delete_protection` to false first.",
		)
		return
	}

	collectionName := state.Name.ValueString()

	// Prevent deletion when indexes still exist on the collection.
	// The caller must remove all milvus_index resources linked to this collection
	// before the collection itself can be destroyed, ensuring no index Terraform
	// resource is left orphaned in state.
	indexNames, err := r.client.ListIndexes(ctx, milvusclient.NewListIndexOption(collectionName))
	if err != nil && !errors.Is(err, merr.ErrIndexNotFound) {
		resp.Diagnostics.AddError(
			"Error listing indexes for collection",
			fmt.Sprintf("Could not list indexes on collection %s: %s", collectionName, err.Error()),
		)
		return
	}
	if len(indexNames) > 0 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Collection %s still has indexes", collectionName),
			fmt.Sprintf(
				"Cannot delete collection %s while the following indexes exist: %v. "+
					"Remove all milvus_index resources linked to this collection first.",
				collectionName, indexNames,
			),
		)
		return
	}

	// Drop the collection
	if err := r.client.DropCollection(ctx, milvusclient.NewDropCollectionOption(collectionName)); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting collection",
			fmt.Sprintf("Could not delete collection %s: %s", collectionName, err.Error()),
		)
		return
	}
}

// ImportState imports the resource by name.
func (r *MilvusCollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
