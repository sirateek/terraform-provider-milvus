// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MilvusCollectionResourceModel is the top-level Terraform state model.
type MilvusCollectionResourceModel struct {
	ID                 types.Int64                 `tfsdk:"id"`
	DeleteProtection   types.Bool                  `tfsdk:"delete_protection"`
	Name               types.String                `tfsdk:"name"`
	Description        types.String                `tfsdk:"description"`
	AutoID             types.Bool                  `tfsdk:"auto_id"`
	EnableDynamicField types.Bool                  `tfsdk:"enable_dynamic_field"`
	ShardNum           types.Int64                 `tfsdk:"shard_num"`
	ConsistencyLevel   types.String                `tfsdk:"consistency_level"`
	Properties         *MilvusCollectionProperties `tfsdk:"properties"`
	Fields             types.List                  `tfsdk:"fields"`
}

type MilvusCollectionProperties struct {
	TTLSeconds            types.Int64  `tfsdk:"collection_ttl_seconds"`
	MMapEnabled           types.Bool   `tfsdk:"mmap_enabled"`
	PartitionKeyIsolation types.Bool   `tfsdk:"partition_key_isolation"`
	DynamicFieldEnabled   types.Bool   `tfsdk:"dynamic_field_enabled"`
	AllowInsertAutoID     types.Bool   `tfsdk:"allow_insert_auto_id"`
	AllowUpdateAutoID     types.Bool   `tfsdk:"allow_update_auto_id"`
	Timezone              types.String `tfsdk:"timezone"`
}

// ParsedMilvusCollectionProperties is the parsed Milvus collection properties.
// It aims to provide ease of Milvus collection property change difference.
type ParsedMilvusCollectionProperties struct {
	TTLSeconds            *int64  `json:"collection.ttl.seconds,omitempty"`
	MMapEnabled           *bool   `json:"mmap.enabled,omitempty"`
	PartitionKeyIsolation *bool   `json:"partitionkey.isolation,omitempty"`
	DynamicFieldEnabled   *bool   `json:"dynamicfield.enabled,omitempty"`
	AllowInsertAutoID     *bool   `json:"allow_insert_auto_id,omitempty"`
	AllowUpdateAutoID     *bool   `json:"allow_update_auto_id,omitempty"`
	Timezone              *string `json:"timezone,omitempty"`
}
