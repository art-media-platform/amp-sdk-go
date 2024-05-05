package registry

import (
	"github.com/amp-3d/amp-sdk-go/amp"
)

func Global() amp.Registry {
	if gRegistry == nil {
		gRegistry = amp.NewRegistry()
	}
	amp.RegisterBuiltinTypes(gRegistry)
	return gRegistry
}

var gRegistry amp.Registry
