package config

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/kelseyhightower/envconfig"
)

type Milvus struct {
	Address  string `envconfig:"ADDRESS" tfsdk:"address"`
	Username string `envconfig:"USERNAME" tfsdk:"username"`
	Password string `envconfig:"PASSWORD" tfsdk:"password"`
	APIKey   string `envconfig:"API_KEY" tfsdk:"api_key"`
	DBName   string `envconfig:"DB_NAME" tfsdk:"db_name"`

	EnableTLS bool `envconfig:"ENABLE_TLS" tfsdk:"enable_tls"`

	ServerVersion string `envconfig:"SERVER_VERSION" tfsdk:"server_version"`
}

func (m *Milvus) MergeOnEmpty(other Milvus) Milvus {
	// Prioritize the value from our object than the other.
	if m.Address == "" {
		m.Address = other.Address
	}
	if m.Username == "" {
		m.Username = other.Username
	}
	if m.Password == "" {
		m.Password = other.Password
	}
	if m.DBName == "" {
		m.DBName = other.DBName
	}
	if m.APIKey == "" {
		m.APIKey = other.APIKey
	}
	if m.ServerVersion == "" {
		m.ServerVersion = other.ServerVersion
	}
	if !m.EnableTLS && other.EnableTLS {
		m.EnableTLS = other.EnableTLS
	}
	return *m
}

func ProvideMilvusConfig(ctx context.Context, req provider.ConfigureRequest) (Milvus, diag.Diagnostics) {
	var tfConfig Milvus
	diags := req.Config.Get(ctx, &tfConfig)

	var envConfig Milvus
	err := envconfig.Process("MILVUS", &envConfig)
	if err != nil {
		diags = append(
			diags,
			diag.NewAttributeErrorDiagnostic(
				path.Empty(),
				"Fail to load Milvus Config from Environment Variable",
				err.Error(),
			),
		)
	}
	return tfConfig.MergeOnEmpty(envConfig), diags
}
