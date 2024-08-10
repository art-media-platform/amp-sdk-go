package registry

import (
	"github.com/art-media-platform/amp-sdk-go/amp"
)

func Global() amp.Registry {
	if gRegistry == nil {
		gRegistry = amp.NewRegistry()
	}
	amp.RegisterBuiltinTypes(gRegistry)
	return gRegistry
}

var gRegistry amp.Registry
