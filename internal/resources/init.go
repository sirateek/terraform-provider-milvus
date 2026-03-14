// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package resources

import (
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
)

func init() {
	// Register resources with the provider
	provider.RegisterResource(NewMilvusCollectionResource)
	provider.RegisterResource(NewMilvusIndexResource)
}
