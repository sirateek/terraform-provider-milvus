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
	"github.com/sirateek/terraform-provider-milvus/internal/resources/alias"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/index"
)

func init() {
	RegisterResource(collection.NewMilvusCollectionResource)
	RegisterResource(index.NewMilvusIndexResource)
	RegisterResource(alias.NewMilvusAliasResource)
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
		MarkdownDescription: "The **Milvus** provider enables infrastructure-as-code management of [Milvus](https://milvus.io), an open-source, cloud-native vector database built for storing, indexing, and searching massive datasets of high-dimensional embedding vectors.\n\nUse this provider to declaratively manage Milvus resources — collections, indexes, and aliases — directly from Terraform, making it easy to version-control your vector database schema alongside the rest of your infrastructure.\n\n## Supported Resources\n\n| Resource | Description |\n|---|---|\n| `milvus_collection` | Manage Milvus collections and their schemas |\n| `milvus_index` | Create and manage indexes on collection fields |\n| `milvus_alias` | Manage aliases that point to collections |\n\n## Authentication\n\nThe provider supports connecting to both self-hosted Milvus clusters and Zilliz Cloud:\n\n- **Address / Username / Password** — for self-hosted Milvus\n- **API Key** — for Zilliz Cloud or API-key-enabled deployments\n- **TLS** — optionally enable encrypted connections\n\n## Example Usage\n\n```hcl\nprovider \"milvus\" {\n  address = \"localhost:19530\"\n}\n```",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				MarkdownDescription: "The `host:port` address of the Milvus server (e.g. `localhost:19530`). Can also be set via the `MILVUS_ADDRESS` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for Milvus authentication. Required when Milvus is running with authentication enabled. Can also be set via the `MILVUS_USERNAME` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for Milvus authentication. Used together with `username`. Can also be set via the `MILVUS_PASSWORD` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"db_name": schema.StringAttribute{
				MarkdownDescription: "The Milvus database to operate on. Defaults to the `default` database when not specified. Can also be set via the `MILVUS_DB_NAME` environment variable.",
				Optional:            true,
			},
			"enable_tls": schema.BoolAttribute{
				MarkdownDescription: "Whether to use TLS when connecting to Milvus. Set to `true` for Zilliz Cloud or any TLS-terminated endpoint. Defaults to `false`.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for Zilliz Cloud or API-key-enabled Milvus deployments. When set, this takes precedence over `username`/`password`. Can also be set via the `MILVUS_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"server_version": schema.StringAttribute{
				MarkdownDescription: "Target Milvus server version (e.g. `2.4.0`). When specified, the provider will validate compatibility against this version. Optional — omit to let the provider detect the version automatically.",
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
