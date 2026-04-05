package collection

import "github.com/hashicorp/terraform-plugin-framework/types"

type MilvusCollectionProperties struct {
	TTLSeconds            types.Int64  `tfsdk:"collection_ttl_seconds"`
	MMapEnabled           types.Bool   `tfsdk:"mmap_enabled"`
	PartitionKeyIsolation types.Bool   `tfsdk:"partition_key_isolation"`
	DynamicFieldEnabled   types.Bool   `tfsdk:"dynamic_field_enabled"`
	AllowInsertAutoID     types.Bool   `tfsdk:"allow_insert_auto_id"`
	AllowUpdateAutoID     types.Bool   `tfsdk:"allow_update_auto_id"`
	Timezone              types.String `tfsdk:"timezone"`
}
