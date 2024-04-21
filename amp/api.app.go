package amp

import (
	"net/url"

	"github.com/amp-space/amp-sdk-go/stdlib/task"
)

// App is how an app module is registered with an amp.Host so it can be invoked.
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

	// Gets the named attribute from the user's home planet -- used high-level app settings.
	// The attr is scoped by both the app UID so key collision with other users or apps is not possible.
	// This is how an app can store and retrieve settings.
	GetAppAttr(attrSpec string, dst ElemVal) error

	// Write analog for GetAppAttr()
	PutAppAttr(attrSpec string, src ElemVal) error
}

// AppInstance is implemented by an App and invoked by amp.Host responding to a client pin request.
type AppInstance interface {
	AppContext // The arc runtime's app context support exposed

	// Instantiation callback made immediately after App.NewAppInstance() -- typically resolves app-specific type specs.
	OnNew(this AppContext) error

	// Celled when the app is pin the cell IAW with the given request.
	// If parent != nil, this is the context of the request.
	// If parent == nil, this app was invoked without out a parent cell / context.
	PinCell(parent PinnedCell, req PinOp) (PinnedCell, error)

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

	// Apps spawn a PinnedCell as a child task.Context of amp.AppContext.Context or as a child of another PinnedCell.
	// This means an AppContext contains all its PinnedCells and thus Close() will close all PinnedCells.
	Context() task.Context

	// Pins the requested cell (typically a child cell).
	PinCell(req PinOp) (PinnedCell, error)

	// Pushes this cell and child cells to the client.
	// Exits when either:
	//   - ctx.Closing() is signaled,
	//   - state has been pushed to the client AND ctx.MaintainSync() == false, or
	//   - an error is encountered
	ServeState(ctx PinContext) error

	// Merges a set of incoming changes into this pinned cell. -- a "write" operation
	MergeTx(tx *TxMsg) error
}

// Serialization abstraction
type PbValue interface {
	Size() int
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Unmarshal(dAtA []byte) error
}

// ElemVal wraps cell attribute element name and serialization.
type ElemVal interface {

	// Returns the element type name (a scalar AttrSpec).
	ElemTypeName() string

	// Marshals this ElemVal to a buffer, reallocating if needed.
	MarshalToStore(in []byte) (out []byte, err error)

	// Unmarshals and merges value state from a buffer.
	Unmarshal(src []byte) error

	// Creates a default instance of this same ElemVal type
	New() ElemVal
}

// TxMsg is workhorse generic transport serialization sent between client and host.
type TxMsg struct {
	TxInfo 
	refCount  int32     // see AddRef() / ReleaseRef()
	Ops       []TxOp    // ordered operations to perform on the target
	DataStore []byte    // marshalled data store for Ops serialized data
}

// TxOp is an atomic operation on a target cell and is a unit of change (or message) for any target.
// Values are typically LSM sorted, so use low order bytes before high order bytes.
// Note that x0 is the most significant and x2 is least significant bytes.
type TxOp struct {
	OpCode       TxOpCode
	TargetID     CellID      // Target cell to operate on
	ParentID     CellID      // Parent cell of the target cell
	AttrID       AttrID      // Attribute to operate on
	SI           SeriesIndex // Index of the data being mutated
	DataStoreOfs int64       // Offset into TxMsg.DataStore
	DataLen      int64       // Length of data in TxMsg.DataStore
}

type AttrDef struct {
	AttrSpec
	Prototype ElemVal
}

// SeriesIndex
type SeriesIndex [2]uint64

// CellID is globally unique Cell identifier that globally identifies a cell.
//
// By convention, the the leading 8 bytes are a UTC16 timestamp and the trailing 8 bytes are pseudo-random.
type CellID [3]uint64

// AttrID is a UID for the canonic string representation of an AttrSpec
// Leading bits are reserved to express pin detail level or layer.
type AttrID UID

func (id AttrID) String() string {
	return UID(id).String()
}

// Leading bits of AttrID are reserved to express pin detail level or layer.
const PinLayerBits = 3
