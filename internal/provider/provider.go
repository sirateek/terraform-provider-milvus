// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/sirateek/terraform-provider-milvus/internal/client/milvus"
	config2 "github.com/sirateek/terraform-provider-milvus/internal/config"
)

// Blank import of resources to trigger resource registration during init()
// This must happen before the provider is used, but doesn't require
// a direct import at the package level
var _ struct{} = struct{}{} // dummy statement

func init() {
	// Trigger resources package initialization to register its resources
	// This is done indirectly through a side-effect import
	// We use an underscore import in the resources package itself
}

// Ensure MilvusProvider satisfies various provider interfaces.
var _ provider.Provider = &MilvusProvider{}
var _ provider.ProviderWithFunctions = &MilvusProvider{}
var _ provider.ProviderWithEphemeralResources = &MilvusProvider{}
var _ provider.ProviderWithActions = &MilvusProvider{}

// MilvusProvider defines the provider implementation.
type MilvusProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *MilvusProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "milvus"
	resp.Version = p.version
}

func (p *MilvusProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
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

func (p *MilvusProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
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

func (p *MilvusProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *MilvusProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *MilvusProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func (p *MilvusProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MilvusProvider{
			version: version,
		}
	}
}
