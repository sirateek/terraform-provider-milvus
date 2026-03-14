// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// resourceFactories holds the registered resource factories
var (
	resourceFactoriesMutex sync.RWMutex
	resourceFactories      []func() resource.Resource
)

// RegisterResource registers a resource factory with the provider
// This is called by the resources package to register its resources
// without creating a circular import dependency
func RegisterResource(factory func() resource.Resource) {
	resourceFactoriesMutex.Lock()
	defer resourceFactoriesMutex.Unlock()
	resourceFactories = append(resourceFactories, factory)
}

// getResources returns all registered resource factories
func getResources() []func() resource.Resource {
	resourceFactoriesMutex.RLock()
	defer resourceFactoriesMutex.RUnlock()

	// Return a copy to avoid external modification
	result := make([]func() resource.Resource, len(resourceFactories))
	copy(result, resourceFactories)
	return result
}

// Resources returns the provider's resources
// This method is part of the provider.Provider interface implementation
func (p *MilvusProvider) Resources(ctx context.Context) []func() resource.Resource {
	return getResources()
}
