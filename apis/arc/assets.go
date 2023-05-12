package arc

import (
	"io"

	"github.com/arcspace/go-arc-sdk/stdlib/process"
)

type MediaAsset interface {

	// Helpful short description of this asset
	Label() string

	// Returns the media / MIME type of this asset
	MediaType() string

	// OnStart is called when this asset is live within the given context.
	// This MediaAsset should:
	//  - close and perform cleanup if/when ctx.Closing() is signaled
	//  - call ctx.Close() if it encounters a fatal error or is no longer available for access.
	OnStart(ctx process.Context) error

	// Called when this asset is requested by a client for read access
	NewAssetReader() (AssetReader, error)
}

// Provides read access to its parent MediaAsset
type AssetReader interface {
	io.ReadSeekCloser
}
