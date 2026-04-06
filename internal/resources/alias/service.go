// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package alias

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/alias/model"
)

// readAlias fetches the alias from Milvus and updates the model.
func (a *Alias) readAlias(ctx context.Context, in *model.Alias, diags diag.Diagnostics) {
	aliasResult, err := a.client.DescribeAlias(ctx, milvusclient.NewDescribeAliasOption(in.Name.ValueString()))
	if err != nil {
		diags.AddError(
			"Error reading alias",
			fmt.Sprintf("Could not describe alias %s: %s", in.Name.ValueString(), err.Error()),
		)
		return
	}

	in.CollectionName = stringValue(aliasResult.CollectionName)
}

// stringValue is a helper to convert a plain string to types.String.
func stringValue(s string) types.String {
	return types.StringValue(s)
}
