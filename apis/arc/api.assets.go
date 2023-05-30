package arc

import (
	"io"
	"time"

	"github.com/arcspace/go-arc-sdk/stdlib/process"
)

// Options when publishing an asset
type PublishOpts struct {
	Expiry time.Duration // If <= 0, the publisher chooses the expiration period
}

// AssetPublisher publishes a MediaAsset to a randomly generated URL until the idle expiration is reached.
// If idleExpiry == 0, the publisher will choose an expiration period.
type AssetPublisher interface {
	PublishAsset(asset MediaAsset, opts PublishOpts) (URL string, err error)
}

// MediaAsset wraps any data asset that can be streamed and is typically audio or video.
type MediaAsset interface {

	// Short name or description of this asset for logging
	Label() string

	// Returns the media / MIME type of this asset
	MediaType() string

	// OnStart is called when this asset is live within the given context.
	// This MediaAsset should call ctx.Close() if a fatal error occurs or its underlying asset becomes unavailable.
	OnStart(ctx process.Context) error

	// Called when this asset is requested by a client for read access
	NewAssetReader() (AssetReader, error)
}

// AssetReader provides read and seek access to its parent MediaAsset.
//
// Close() should be called when the reader is no longer needed or when its parent MediaAsset becomes unavailable.
type AssetReader interface {
	io.ReadSeekCloser
}
