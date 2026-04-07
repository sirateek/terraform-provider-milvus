// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package index

import (
	"context"
	"fmt"
	"strings"

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
		MarkdownDescription: "Manages a Milvus **index** on a collection field.\n\nIndexes accelerate vector similarity search and scalar filtering. Each field can have at most one index. All index properties are **immutable** — to change an index you must delete it and create a new one.\n\n~> A `milvus_collection` cannot be destroyed while any `milvus_index` resource still references it. Remove all indexes before (or together with) the collection.",
		Attributes: map[string]schema.Attribute{
			"collection_name": schema.StringAttribute{
				MarkdownDescription: "Name of the collection that contains the field to index. **Immutable** — changing this forces a new index to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"field_name": schema.StringAttribute{
				MarkdownDescription: "Name of the field to build the index on. **Immutable** — changing this forces a new index to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index_name": schema.StringAttribute{
				MarkdownDescription: "Optional display name for the index. When omitted, the field name is used as the index name. **Immutable** — changing this forces a new index to be created.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index_type": schema.StringAttribute{
				MarkdownDescription: "Algorithm used to build the index. **Immutable** — changing this forces a new index to be created.\n\n**Vector indexes:** `FLAT`, `IVF_FLAT`, `IVF_SQ8`, `IVF_PQ`, `HNSW`, `DISKANN`, `SCANN`, `AUTOINDEX`, `SPARSE_INVERTED`, `SPARSE_WAND`, `MINHASH_LSH`\n\n**Scalar indexes:** `TRIE` (VarChar), `SORTED` (numeric), `INVERTED`, `BITMAP`, `RTREE`",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metric_type": schema.StringAttribute{
				MarkdownDescription: "Distance metric used for vector similarity calculations. **Immutable** — changing this forces a new index to be created.\n\n| Value | Use case |\n|---|---|\n| `L2` | Euclidean distance. Common for dense float vectors. |\n| `COSINE` | Cosine similarity. Good for normalised embeddings. |\n| `IP` | Inner product. Equivalent to cosine when vectors are unit-normalised. |\n| `HAMMING` | Bit-level Hamming distance. For binary vectors. |\n| `JACCARD` | Jaccard similarity. For binary vectors. |\n\nFor scalar indexes this field is still required by the API but has no effect.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index_params": schema.SingleNestedAttribute{
				MarkdownDescription: "Index-type-specific tuning parameters. Only the parameters relevant to the chosen `index_type` need to be set; all others are ignored.",
				Attributes: map[string]schema.Attribute{
					"nlist": schema.Int64Attribute{
						MarkdownDescription: "Number of cluster centroids (inverted lists) for IVF-based indexes (`IVF_FLAT`, `IVF_SQ8`, `IVF_PQ`, `SCANN`). Higher values improve recall at the cost of slower build time and larger memory usage. Typical range: `64`–`65536`.",
						Optional:            true,
					},
					"m": schema.Int64Attribute{
						MarkdownDescription: "For `HNSW`: maximum number of bidirectional links per node (graph degree). Higher `m` improves recall but increases memory usage and build time. Typical range: `4`–`64`.\n\nFor `IVF_PQ`: number of sub-quantizers. Must divide the vector dimension evenly.",
						Optional:            true,
					},
					"nbits": schema.Int64Attribute{
						MarkdownDescription: "For `IVF_PQ`: bits used to encode each sub-vector. Higher values improve accuracy at the cost of memory. Valid values: `8` (default).",
						Optional:            true,
					},
					"ef_construction": schema.Int64Attribute{
						MarkdownDescription: "For `HNSW`: size of the dynamic candidate list during graph construction. Higher values produce a better-quality graph (higher recall) at the cost of longer build time. Must be ≥ `m`. Typical range: `8`–`512`.",
						Optional:            true,
					},
					"ef": schema.Int64Attribute{
						MarkdownDescription: "For `HNSW`: size of the dynamic candidate list during search. Higher values improve recall at the cost of query latency. Must be ≥ the `top_k` requested at query time.",
						Optional:            true,
					},
					"with_raw_data": schema.BoolAttribute{
						MarkdownDescription: "For `SCANN`: when `true`, raw vector data is stored alongside the index to enable reranking, improving recall at a small storage cost.",
						Optional:            true,
					},
					"drop_ratio": schema.Float64Attribute{
						MarkdownDescription: "For sparse vector indexes (`SPARSE_INVERTED`, `SPARSE_WAND`): fraction of the smallest magnitude values to drop during indexing. Range `0.0`–`1.0`. Higher values reduce index size but may lower recall. Defaults to `0.2`.",
						Optional:            true,
					},
					"intermediate_graph_degree": schema.Int64Attribute{
						MarkdownDescription: "For GPU `CAGRA` index: degree of the intermediate kNN graph constructed during the build phase. Affects build quality and time.",
						Optional:            true,
					},
					"graph_degree": schema.Int64Attribute{
						MarkdownDescription: "For GPU `CAGRA` index: degree of the final search graph. Must be less than or equal to `intermediate_graph_degree`.",
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
	desc, err := r.client.DescribeIndex(ctx, opt)
	if err != nil {
		// If index doesn't exist, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// After import, index_type and metric_type are not in state yet — populate
	// them from the DescribeIndex response so the post-import plan is empty.
	if state.IndexType.IsNull() || state.IndexType.IsUnknown() {
		params := desc.Params()
		if v, ok := params[index.IndexTypeKey]; ok {
			state.IndexType = types.StringValue(v)
		}
		if v, ok := params[index.MetricTypeKey]; ok {
			state.MetricType = types.StringValue(v)
		}
		// index_name is the authoritative name Milvus stores.
		state.IndexName = types.StringValue(desc.Name())
	}

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

// ImportState imports a milvus_index resource.
//
// The import ID must be in the format: <collection_name>/<index_name>
// where index_name is the explicit index name, or the field name if no
// index name was set when the index was created.
//
// Example:
//
//	terraform import milvus_index.my_index my_collection/embedding_flat
func (r *MilvusIndexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf(
				"Expected import ID in the format <collection_name>/<index_name>, got %q.\n\n"+
					"Use the index_name if one was set explicitly, otherwise use the field_name.",
				req.ID,
			),
		)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("collection_name"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("field_name"), parts[1])
	resp.State.SetAttribute(ctx, path.Root("index_name"), parts[1])
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
