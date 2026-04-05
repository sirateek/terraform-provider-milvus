// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package index

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// MilvusIndexResourceModel is the top-level Terraform state model.
type MilvusIndexResourceModel struct {
	CollectionName types.String           `tfsdk:"collection_name"`
	FieldName      types.String           `tfsdk:"field_name"`
	IndexName      types.String           `tfsdk:"index_name"`
	IndexType      types.String           `tfsdk:"index_type"`
	MetricType     types.String           `tfsdk:"metric_type"`
	IndexParams    *MilvusIndexParameters `tfsdk:"index_params"`
}

// MilvusIndexParameters contains index-specific parameters.
type MilvusIndexParameters struct {
	NList           types.Int64   `tfsdk:"nlist"`
	M               types.Int64   `tfsdk:"m"`
	NBits           types.Int64   `tfsdk:"nbits"`
	EFConstruction  types.Int64   `tfsdk:"ef_construction"`
	EF              types.Int64   `tfsdk:"ef"`
	WithRawData     types.Bool    `tfsdk:"with_raw_data"`
	DropRatio       types.Float64 `tfsdk:"drop_ratio"`
	IntermediateGrD types.Int64   `tfsdk:"intermediate_graph_degree"`
	GraphDegree     types.Int64   `tfsdk:"graph_degree"`
}

// MilvusIndexResource is the resource implementation.
type MilvusIndexResource struct {
	client *milvusclient.Client
}

var _ resource.Resource = &MilvusIndexResource{}
var _ resource.ResourceWithConfigure = &MilvusIndexResource{}
var _ resource.ResourceWithImportState = &MilvusIndexResource{}

// NewMilvusIndexResource is a helper function to simplify the provider implementation.
func NewMilvusIndexResource() resource.Resource {
	return &MilvusIndexResource{}
}

// Metadata returns the resource type name.
func (r *MilvusIndexResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index"
}

// Schema defines the schema for the resource.
func (r *MilvusIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Milvus index.",
		Attributes: map[string]schema.Attribute{
			"collection_name": schema.StringAttribute{
				MarkdownDescription: "The name of the collection to create the index on. Required and immutable.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"field_name": schema.StringAttribute{
				MarkdownDescription: "The name of the field to create the index on. Required and immutable.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index_name": schema.StringAttribute{
				MarkdownDescription: "The name of the index. If not specified, Milvus will generate one automatically.",
				Optional:            true,

				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index_type": schema.StringAttribute{
				MarkdownDescription: "The type of index (e.g., FLAT, IVF_FLAT, IVF_SQ8, IVF_PQ, HNSW, DISKANN, SCANN, AUTOINDEX, TRIE, SORTED, INVERTED, BITMAP, SPARSE_INVERTED, SPARSE_WAND, RTREE, MINHASH_LSH, SPARSE_CORD_INVERTED). Required and immutable.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metric_type": schema.StringAttribute{
				MarkdownDescription: "The metric type for vector distance (e.g., L2, COSINE, IP). Required for vector indexes.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index_params": schema.SingleNestedAttribute{
				MarkdownDescription: "Index-specific parameters. Optional and varies by index type.",
				Attributes: map[string]schema.Attribute{
					"nlist": schema.Int64Attribute{
						MarkdownDescription: "Number of clusters (inverted lists) for IVF-based indexes.",
						Optional:            true,
					},
					"m": schema.Int64Attribute{
						MarkdownDescription: "Number of subquantizers for IVF_PQ or degree parameter for HNSW.",
						Optional:            true,
					},
					"nbits": schema.Int64Attribute{
						MarkdownDescription: "Bits per subvector for IVF_PQ index.",
						Optional:            true,
					},
					"ef_construction": schema.Int64Attribute{
						MarkdownDescription: "Construction parameter for HNSW index.",
						Optional:            true,
					},
					"ef": schema.Int64Attribute{
						MarkdownDescription: "Search parameter for HNSW index.",
						Optional:            true,
					},
					"with_raw_data": schema.BoolAttribute{
						MarkdownDescription: "Whether to store raw data for SCANN index.",
						Optional:            true,
					},
					"drop_ratio": schema.Float64Attribute{
						MarkdownDescription: "Drop ratio for sparse indexes.",
						Optional:            true,
					},
					"intermediate_graph_degree": schema.Int64Attribute{
						MarkdownDescription: "Intermediate graph degree for GPU CAGRA index.",
						Optional:            true,
					},
					"graph_degree": schema.Int64Attribute{
						MarkdownDescription: "Graph degree for GPU CAGRA index.",
						Optional:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *MilvusIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *MilvusIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MilvusIndexResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build index object based on index type
	idx := r.buildIndex(plan)
	if idx == nil {
		resp.Diagnostics.AddError(
			"Unsupported Index Type",
			fmt.Sprintf("Index type '%s' is not supported", plan.IndexType.ValueString()),
		)
		return
	}

	// Build create index option
	opt := milvusclient.NewCreateIndexOption(plan.CollectionName.ValueString(), plan.FieldName.ValueString(), idx)

	// Set index name if provided
	if !plan.IndexName.IsNull() {
		opt = opt.WithIndexName(plan.IndexName.ValueString())
	}

	// Create the index
	task, err := r.client.CreateIndex(ctx, opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating index",
			fmt.Sprintf("Could not create index on %s.%s: %s", plan.CollectionName.ValueString(), plan.FieldName.ValueString(), err.Error()),
		)
		return
	}

	// Wait for index creation to complete
	err = task.Await(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for index creation",
			fmt.Sprintf("Index creation task failed: %s", err.Error()),
		)
		return
	}

	// Set state with the plan values as-is
	// If index_name was null in plan, it remains null in state (Milvus uses field name as default)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *MilvusIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MilvusIndexResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the index name to look up
	indexName := state.IndexName.ValueString()
	if indexName == "" {
		// If index name is not set, use field name (Milvus default)
		indexName = state.FieldName.ValueString()
	}

	// Try to describe the index to verify it still exists
	opt := milvusclient.NewDescribeIndexOption(state.CollectionName.ValueString(), indexName)
	_, err := r.client.DescribeIndex(ctx, opt)
	if err != nil {
		// If index doesn't exist, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Index exists and is verified - state is still valid
	// (Index properties are immutable in Milvus, so we don't need to refresh them)

	// Index is verified to exist, set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for indexes as they are immutable.
func (r *MilvusIndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Milvus indexes are immutable. To change an index, delete it and create a new one.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *MilvusIndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MilvusIndexResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine which index name to use
	indexName := state.IndexName.ValueString()
	if indexName == "" {
		// If index name is not set, construct it from field name
		indexName = state.FieldName.ValueString()
	}

	// Drop the index
	opt := milvusclient.NewDropIndexOption(state.CollectionName.ValueString(), indexName)
	err := r.client.DropIndex(ctx, opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting index",
			fmt.Sprintf("Could not delete index %s on %s.%s: %s", indexName, state.CollectionName.ValueString(), state.FieldName.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource by collection_name.field_name.
func (r *MilvusIndexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("field_name"), req, resp)
}

// Helper Functions

// buildIndex constructs a Milvus index object based on the plan.
func (r *MilvusIndexResource) buildIndex(plan MilvusIndexResourceModel) index.Index {
	indexType := plan.IndexType.ValueString()
	metricType := r.getMetricType(plan.MetricType.ValueString())
	var params *MilvusIndexParameters
	if plan.IndexParams != nil {
		params = plan.IndexParams
	}

	switch indexType {
	case "FLAT":
		return index.NewFlatIndex(metricType)

	case "IVF_FLAT":
		nlist := int64(100)
		if params != nil && !params.NList.IsNull() {
			nlist = params.NList.ValueInt64()
		}
		return index.NewIvfFlatIndex(metricType, int(nlist))

	case "IVF_SQ8":
		nlist := int64(100)
		if params != nil && !params.NList.IsNull() {
			nlist = params.NList.ValueInt64()
		}
		return index.NewIvfSQ8Index(metricType, int(nlist))

	case "IVF_PQ":
		nlist := int64(100)
		m := int64(8)
		nbits := int64(8)
		if params != nil {
			if !params.NList.IsNull() {
				nlist = params.NList.ValueInt64()
			}
			if !params.M.IsNull() {
				m = params.M.ValueInt64()
			}
			if !params.NBits.IsNull() {
				nbits = params.NBits.ValueInt64()
			}
		}
		return index.NewIvfPQIndex(metricType, int(nlist), int(m), int(nbits))

	case "HNSW":
		m := int64(16)
		efConstruction := int64(200)
		if params != nil {
			if !params.M.IsNull() {
				m = params.M.ValueInt64()
			}
			if !params.EFConstruction.IsNull() {
				efConstruction = params.EFConstruction.ValueInt64()
			}
		}
		return index.NewHNSWIndex(metricType, int(m), int(efConstruction))

	case "DISKANN":
		return index.NewDiskANNIndex(metricType)

	case "SCANN":
		nlist := int64(100)
		withRawData := false
		if params != nil {
			if !params.NList.IsNull() {
				nlist = params.NList.ValueInt64()
			}
			if !params.WithRawData.IsNull() {
				withRawData = params.WithRawData.ValueBool()
			}
		}
		return index.NewSCANNIndex(metricType, int(nlist), withRawData)

	case "AUTOINDEX":
		return index.NewAutoIndex(metricType)

	case "TRIE":
		return index.NewTrieIndex()

	case "SORTED":
		return index.NewSortedIndex()

	case "INVERTED":
		return index.NewInvertedIndex()

	case "BITMAP":
		return index.NewBitmapIndex()

	case "SPARSE_INVERTED":
		dropRatio := 0.2
		if params != nil && !params.DropRatio.IsNull() {
			dropRatio = params.DropRatio.ValueFloat64()
		}
		return index.NewSparseInvertedIndex(metricType, dropRatio)

	case "SPARSE_WAND":
		dropRatio := 0.2
		if params != nil && !params.DropRatio.IsNull() {
			dropRatio = params.DropRatio.ValueFloat64()
		}
		return index.NewSparseWANDIndex(metricType, dropRatio)

	case "RTREE":
		return index.NewRTreeIndex()

	case "MINHASH_LSH":
		return index.NewMinHashLSHIndex(metricType, 8)

	default:
		return nil
	}
}

// getMetricType converts a string to entity.MetricType.
func (r *MilvusIndexResource) getMetricType(metricTypeStr string) entity.MetricType {
	switch metricTypeStr {
	case "L2":
		return entity.L2
	case "COSINE":
		return entity.COSINE
	case "IP":
		return entity.IP
	case "HAMMING":
		return entity.HAMMING
	case "JACCARD":
		return entity.JACCARD
	default:
		return entity.L2
	}
}
