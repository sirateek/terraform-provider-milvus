//
// SPDX-License-Identifier: MPL-2.0

package testing

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ProtoV6ProviderFactories holds the provider factories registered for testing.
var ProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

var (
	// Mutex protects access to ProtoV6ProviderFactories.
	factoriesMutex sync.RWMutex
)

func init() {
	ProtoV6ProviderFactories = make(map[string]func() (tfprotov6.ProviderServer, error))
}

// RegisterProviderFactory registers a provider factory for testing.
// This is called by the provider package's init() function to register
// its factory without creating circular imports.
func RegisterProviderFactory(name string, factory func() (tfprotov6.ProviderServer, error)) {
	factoriesMutex.Lock()
	defer factoriesMutex.Unlock()
	ProtoV6ProviderFactories[name] = factory
}
