// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/sirateek/terraform-provider-milvus/internal/client/milvus"
	config2 "github.com/sirateek/terraform-provider-milvus/internal/config"
	"github.com/sirateek/terraform-provider-milvus/internal/resources"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ScaffoldingProvider{}
var _ provider.ProviderWithFunctions = &ScaffoldingProvider{}
var _ provider.ProviderWithEphemeralResources = &ScaffoldingProvider{}
var _ provider.ProviderWithActions = &ScaffoldingProvider{}

// ScaffoldingProvider defines the provider implementation.
type ScaffoldingProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *ScaffoldingProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scaffolding"
	resp.Version = p.version
}

func (p *ScaffoldingProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				MarkdownDescription: "Address of Milvus",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username of Milvus",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of Milvus",
				Optional:            true,
				Sensitive:           true,
			},
			"db_name": schema.StringAttribute{
				MarkdownDescription: "Database name of Milvus to manage",
				Optional:            true,
			},
			"enable_tls": schema.BoolAttribute{
				MarkdownDescription: "Enable TLS for Milvus connection",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key of Milvus",
				Optional:            true,
				Sensitive:           true,
			},
			"server_version": schema.StringAttribute{
				MarkdownDescription: "Version of Milvus to manage",
				Optional:            true,
			},
		},
	}
}

func (p *ScaffoldingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	config, diags := config2.ProvideMilvusConfig(ctx, req)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, milvusClientDiag := milvus.ProvideMilvusClient(config)
	resp.Diagnostics.Append(milvusClientDiag)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ScaffoldingProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewMilvusCollectionResource,
	}
}

func (p *ScaffoldingProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *ScaffoldingProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *ScaffoldingProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func (p *ScaffoldingProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ScaffoldingProvider{
			version: version,
		}
	}
}
