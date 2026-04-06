// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package alias

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/milvus-io/milvus/pkg/v2/util/merr"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/alias/model"
)

var _ resource.Resource = &Alias{}
var _ resource.ResourceWithConfigure = &Alias{}
var _ resource.ResourceWithImportState = &Alias{}

type Alias struct {
	client *milvusclient.Client
}

// NewMilvusAliasResource provide the Milvus Alias resource representation.
func NewMilvusAliasResource() resource.Resource {
	return &Alias{}
}

func (a *Alias) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Milvus **alias** — a secondary, mutable name that points to a collection.\n\nAliases let you decouple your application's connection string from the physical collection name. You can atomically re-point an alias to a different collection (e.g. after a data refresh or during A/B testing) without touching application code.\n\n~> **Note:** The alias `name` is immutable. To rename an alias you must destroy and recreate it. The `collection_name` it points to can be updated in-place at any time.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the alias within the database. **Immutable** — changing this forces a new resource to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collection_name": schema.StringAttribute{
				MarkdownDescription: "Name of the collection this alias points to. Can be updated in-place; Milvus will atomically re-point the alias without downtime.",
				Required:            true,
			},
		},
	}
}

func (a *Alias) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_alias"
}

func (a *Alias) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), request, response)
}

func (a *Alias) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*milvusclient.Client)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *milvusclient.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	a.client = client
}

func (a *Alias) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan model.Alias

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	err := a.client.CreateAlias(ctx, milvusclient.NewCreateAliasOption(plan.CollectionName.ValueString(), plan.Name.ValueString()))
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating alias",
			fmt.Sprintf("Could not create alias %s for collection %s: %s", plan.Name.ValueString(), plan.CollectionName.ValueString(), err.Error()),
		)
		return
	}

	a.readAlias(ctx, &plan, response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (a *Alias) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state model.Alias

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	aliasResult, err := a.client.DescribeAlias(ctx, milvusclient.NewDescribeAliasOption(state.Name.ValueString()))
	if err != nil {
		if !errors.Is(err, merr.ErrAliasNotFound) {
			response.Diagnostics.AddError("Error reading alias", err.Error())
		}
		// Alias no longer exists; remove from state.
		response.State.RemoveResource(ctx)
		return
	}

	state.CollectionName = stringValue(aliasResult.CollectionName)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (a *Alias) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan model.Alias

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	err := a.client.AlterAlias(ctx, milvusclient.NewAlterAliasOption(plan.Name.ValueString(), plan.CollectionName.ValueString()))
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating alias",
			fmt.Sprintf("Could not alter alias %s to point to collection %s: %s", plan.Name.ValueString(), plan.CollectionName.ValueString(), err.Error()),
		)
		return
	}

	a.readAlias(ctx, &plan, response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (a *Alias) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state model.Alias

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	err := a.client.DropAlias(ctx, milvusclient.NewDropAliasOption(state.Name.ValueString()))
	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting alias",
			fmt.Sprintf("Could not drop alias %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}
