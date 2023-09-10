package arc

import (
	"net/url"

	"github.com/arcspace/go-arc-sdk/stdlib/task"
)

// App is how an app module is registered with an arc.Host so it can be invoked.
//
// An App is invoked by a client or other app via the app's UID or URI.
type App struct {

	// AppID identifies this app with form "{AppNameID}.{FamilyID}.{PublisherID}" -- e.g. "filesys.hosting.arcspace.systems"
	//   - PublisherID: typically the domain name of the publisher of this app -- e.g. "arcspace.systems"
	//   - FamilyID:    encompassing namespace ID used to group related apps (no spaces or punctuation)
	//   - AppNameID:   identifies this app within its parent family and domain (no spaces or punctuation)
	//
	// AppID form is consistent with a URL domain name (and subdomains).
	AppID        string
	UID          UID      // Universally unique and persistent ID for this module (and the module's "home" planet if present)
	Desc         string   // Human-readable description of this app
	Version      string   // "v{MajorVers}.{MinorID}.{RevID}"
	Dependencies []UID    // Module UIDs this app may access
	Invocations  []string // Additional aliases that invoke this app
	AttrDecl     []string // Attrs to be resolved and registered with a HostSession

	// NewAppInstance is the entry point for an App.
	// Called when an App is invoked on an active User session and is not yet running.
	NewAppInstance func() AppInstance
}

// AppContext is provided by the arc runtime to an AppInstance for support and context.
type AppContext interface {
	task.Context          // Allows select{} for graceful handling of app shutdown
	AssetPublisher        // Allows an app to publish assets for client consumption
	Session() HostSession // Access to underlying Session

	// Returns the absolute file system path of the app's local read-write directory.
	// This directory is scoped by the app's UID.
	LocalDataPath() string

	// Issues a mew cell ID guaranteed to be universally unique.
	// This should not be called concurrently with other IssueCellID() calls.
	IssueCellID() CellID

	// Gets the named cell and attribute from the user's home planet -- used high-level app settings.
	// The attr is scoped by both the app UID so key collision with other users or apps is not possible.
	// This is how an app can store and retrieve settings.
	GetAppCellAttr(attrSpec string, dst AttrElemVal) error

	// Write analog for GetAppCellAttr()
	PutAppCellAttr(attrSpec string, src AttrElemVal) error
}

// AppInstance is implemented by an App and invoked by arc.Host responding to a client pin request.
type AppInstance interface {
	AppContext // The arc runtime's app context support exposed

	// Instantiation callback made immediately after App.NewAppInstance() -- typically resolves app-specific type specs.
	OnNew(this AppContext) error

	// Celled when the app is pin the cell IAW with the given request.
	// If parent != nil, this is the context of the request.
	// If parent == nil, this app was invoked without out a parent cell / context.
	PinCell(parent PinnedCell, req PinReq) (PinnedCell, error)

	// Handles a meta message sent to this app, which could be any attr type.
	HandleURL(*url.URL) error

	// Called exactly once when an app is signaled to close.
	OnClosing()
}

// Cell is an interface for an app Cell
type Cell interface {

	// Returns this cell's immutable info.
	Info() CellID
}

// PinnedCell is how your app encapsulates a pinned cell to the host runtime and thus clients.
type PinnedCell interface {
	Cell

	// Apps spawn a PinnedCell as a child task.Context of arc.AppContext.Context or as a child of another PinnedCell.
	// This means an AppContext contains all its PinnedCells and thus Close() will close all PinnedCells.
	Context() task.Context

	// Pins the requested cell (typically a child cell).
	PinCell(req PinReq) (PinnedCell, error)

	// Pushes this cell and child cells to the client.
	// Exits when either:
	//   - ctx.Closing() is signaled,
	//   - state has been pushed to the client AND ctx.MaintainSync() == false, or
	//   - an error is encountered
	ServeState(ctx PinContext) error

	// Merges an incoming change set into this pinned cell. -- aka write operation
	MergeTx(tx *TxMsg) error
}

// Serialization abstraction
type PbValue interface {
	Size() int
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Unmarshal(dAtA []byte) error
}

// AttrElemVal wraps cell attribute element type name and serialization.
type AttrElemVal interface {

	// Returns the element type name (a degenerate AttrSpec).
	ElemTypeName() string

	// Marshals this value to the end of a buffer.
	MarshalToStore(in []byte) (out []byte, err error)

	// Unmarshals and merges value state from a buffer.
	Unmarshal(src []byte) error

	// Creates a default instance of this same AttrElemVal type
	New() AttrElemVal
}

// TxMsg is a multi-cell state update for a pinned cell or a container of meta attrs.
type TxMsg struct {
	ReqID     uint64    // allows replies to be routed to an originator if applicable
	Status    ReqStatus // status of the originating request if applicable
	CellOps   []CellOp  // Ordered operations in this tx
	DataStore []byte    // backing store
}

type AttrElem struct {
	AttrID      AttrUID     // attr being modified -- 0 denotes preceding CellOp's AttrID
	SeriesIndex SeriesIndex // series index (if applicable)
	DataOfs     int64       // Byte offset serialized location into parent TxMsg's data store
	DataLen     int64       // Byte length of serialized data
}

type CellOp struct {
	AttrElem
	OpCode      CellOpCode  // operation to perform
	CellID      CellID      // cell being modified -- 0 denotes preceding CellOp's CellID
	// AttrID      AttrUID     // attr being modified -- 0 denotes preceding CellOp's AttrID
	// SeriesIndex SeriesIndex // series index (if applicable)
	// DataOfs     int64       // Byte offset serialized location into parent TxMsg's data store
	// DataLen     int64       // Byte length of serialized data
}

// AttrSet is an ordered set of AttrSpec's that is used to select or mask a Cell's attributes.
type AttrSet struct {
	DefID AttrUID
	Attrs []AttrUID
}

// CellID is globally unique Cell identifier that globally identifies a cell.
//
// By convention, the the leading 8 bytes are a UTC16 timestamp and the trailing 8 bytes are pseudo-random.
// If the leading 8 bytes are 0, this denotes an ephemeral cell, meaning it does not no originate from a persistent store.
type CellID [2]uint64

// General purpose index of an cell attribute element in a series.
type SeriesIndex [2]uint64

// AttrUID is a universally unique identifier for an AttrSpec, generated from the MD5 of the canonic AttrSpec string.
type AttrUID [2]uint64
