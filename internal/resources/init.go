// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package resources

import (
	"github.com/sirateek/terraform-provider-milvus/internal/provider"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/collection"
	"github.com/sirateek/terraform-provider-milvus/internal/resources/index"
)

func init() {
	// Register resources with the provider
	provider.RegisterResource(collection.NewMilvusCollectionResource)
	provider.RegisterResource(index.NewMilvusIndexResource)
}
