package arc

import (
	"reflect"

	"github.com/arcspace/go-arc-sdk/stdlib/process"
)

type CellID uint64

// U64 is a convenience method that converts a CellID to a uint64.
func (ID CellID) U64() uint64 { return uint64(ID) }

// Host is the highest level controller.
// Child processes attach to it and start new host sessions as needed.
type Host interface {
	process.Context

	HostPlanet() Planet

	Registry() Registry

	// StartNewSession creates a new HostSession and binds its Msg transport to the given steam.
	StartNewSession(parent HostService, via ServerStream) (HostSession, error)
}

// HostService attaches to a arc.Host as a child process, extending host functionality.
// For example. it wraps a Grpc-based Msg transport as well as a dll-based Msg transport implementation.
type HostService interface {
	process.Context

	// Returns short string identifying this service
	ServiceURI() string

	// Returns the parent Host this extension is attached to.
	Host() Host

	// StartService attaches a child process to the given host and starts this HostService.
	StartService(on Host) error

	// GracefulStop initiates a polite stop of this extension and blocks until it's in a "soft" closed state,
	//    meaning that its service has effectively stopped but its Context is still open.
	// Note this could any amount of time (e.g. until all open requests are closed)
	// Typically, GracefulStop() is called (blocking) and then Context.Close().
	// To stop immediately, Context.Close() is always available.
	GracefulStop()
}

// HostSession in an open session instance with a Host.
// Closing is initiated via Context.Close().
type HostSession interface {
	process.Context

	// Thread-safe
	TypeRegistry

	// Returns the running AssetPublisher instance for this session.
	AssetPublisher() AssetPublisher

	LoggedIn() User
}

// User + HostSession --> UserSession?
type User interface {
	Session() HostSession

	HomePlanet() Planet

	LoginInfo() LoginReq

	// Move to AppContext?
	PushMetaMsg(msg *Msg) error

	// Gets the currently active AppContext for an AppID.
	// If does not exist and autoCreate is set, a new AppContext is created, started, and returned.
	GetAppContext(appID UID, autoCreate bool) (AppContext, error)
}

// ServerStream wraps a Msg transport abstraction, allowing a Host to connect over any data transport layer.
// This is intended to be implemented by a grpc and other transport layers.
type ServerStream interface {

	// Describes this stream
	Desc() string

	// Called when this stream to be closed because the associated parent host session is closing or has closed.
	Close()

	// SendMsg sends a Msg to the remote client.
	// ErrStreamClosed is used to denote normal stream close.
	// Like grpc.ServerStream.SendMsg(), on exit, the Msg has been copied and so can be reused.
	SendMsg(m *Msg) error

	// RecvMsg blocks until it receives a Msg or the stream is done.
	// ErrStreamClosed is used to denote normal stream close.
	RecvMsg() (*Msg, error)
}

// CellSub is a subscriber to state update msgs for a pinned cell.
type CellSub interface {

	// Sets msg.ReqID and pushes the given msg to client, blocking until "complete" (queued) or canceled.
	// This msg is reclaimed after it is sent, so it should be accessed following this call.
	PushMsg(msg *Msg) error
}

// Planet is content and governance enclosure.
// A Planet is 1:1 with a KV database model, which works out well for efficiency and performance.
type Planet interface {

	// A Planet instance is a child process of a host
	process.Context

	PlanetID() uint64

	// A planet offers a persistent symbol table, allowing efficient compression of byte symbols into uint64s
	GetSymbolID(value []byte, autoIssue bool) (ID uint64)
	GetSymbol(ID uint64, io []byte) []byte

	// WIP
	PushTx(tx *MsgBatch) error
	ReadCell(cellKey []byte, schema *AttrSchema, msgs func(msg *Msg)) error

	//GetCell(ID CellID) (CellInstance, error)

	// BlobStore offers access to this planet's blob store (referenced via ValueType_BlobID).
	//blob.Store
}

type CellContext interface {

	// PinCell pins a requested cell, typically specified by req.PinCell.
	// req.KwArgs and ChildSchemas can also be used to specify the cell to pin.
	PinCell(req *CellReq) (AppCell, error)
}

type AppContext interface {
	process.Context // Each app instance has a process.Context
	AssetPublisher  // Allows an app to publish assets for client consumption
	User() User     // Access to user operations and io
	CellContext     // How to pin root cells

	// Atomically issues a new and unique ID that will remain globally unique for the duration of this session.
	// An ID may still expire, go out of scope, or otherwise become meaningless.
	IssueCellID() CellID

	// Unique state scope ID for this app instance -- defaults to the app's UID.
	StateScope() []byte

	// Uses reflection to build and register (as necessary) an AttrSchema for a given a ptr to a struct.
	GetSchemaForType(typ reflect.Type) (*AttrSchema, error)

	// Starts a child process
	// StartChild(task *process.Task) (process.Context, error)

	// Loads the data stored at the given key, appends it to the given buffer, and returns the result (or an error).
	// The given subKey is scoped by both the app and the user so key collision with other users or apps is not possible.
	// Typically used by apps for holding high-level state or settings.
	GetAppValue(subKey string) (val []byte, err error)

	// Write analog for GetAppValue()
	PutAppValue(subKey string, val []byte) error
}

// PushCellOpts specifies how an AppCell should be pushed to the client
type PushCellOpts uint32

const (
	PushAsParent PushCellOpts = 1 << iota
	PushAsChild
)

func (opts PushCellOpts) PushAsParent() bool { return opts&PushAsParent != 0 }
func (opts PushCellOpts) PushAsChild() bool  { return opts&PushAsChild != 0 }

// type CellInfo struct {
// 	CellID
// 	CellDataModel string
// 	Label         string
// }

// PinnedCell?
// AppCell is how an App offers a cell instance to the planet runtime.
type AppCell interface {
	//process.Context    // Started as sub of the app's AppContext

	CellContext

	//Info() CellInfo

	// Returns the CellID of this cell
	ID() CellID

	// Names the data model that this cell implements.
	CellDataModel() string

	// Called when a cell is pinned and should push its state (in accordance with req.ContentSchema & req.ChildSchemas supplied by the client).
	// The implementation uses req.CellSub.PushMsg(...) to push attributes and child cells to the client.
	// Called on the goroutine owned by the the target CellID.
	PushCellState(req *CellReq, opts PushCellOpts) error
}
