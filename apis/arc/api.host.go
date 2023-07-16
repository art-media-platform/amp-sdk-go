package arc

import (
	"net/url"

	"github.com/arcspace/go-arc-sdk/stdlib/symbol"
	"github.com/arcspace/go-arc-sdk/stdlib/task"
)

// Host is the highest level controller.
// Child processes attach to it and start new host sessions as needed.
type Host interface {
	task.Context

	Registry() Registry

	// StartNewSession creates a new HostSession and binds its Msg transport to a stream.
	StartNewSession(parent HostService, via Transport) (HostSession, error)
}

// Transport wraps a Msg transport abstraction, allowing a Host to connect over any data transport layer.
// This is intended to be implemented by a grpc and other transport layers.
type Transport interface {

	// Describes this stream
	Desc() string

	// Called when this stream to be closed because the associated parent host session is closing or has closed.
	Close()

	// SendMsg sends a Msg to the remote client.
	// ErrStreamClosed is used to denote normal stream close.
	// Like grpc.Transport.SendMsg(), on exit, the Msg has been copied and so can be reused.
	SendMsg(m *Msg) error

	// RecvMsg blocks until it receives a Msg or the stream is done.
	// ErrStreamClosed is used to denote normal stream close.
	RecvMsg() (*Msg, error)
}

// HostService attaches to a arc.Host as a child task, extending host functionality.
// For example. it wraps a Grpc-based Msg transport as well as a dll-based Msg transport implementation.
type HostService interface {
	task.Context

	// Returns short string identifying this service
	ServiceURI() string

	// Returns the parent Host this extension is attached to.
	Host() Host

	// StartService attaches a child task to a Host and starts this HostService.
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
	task.Context    // Underlying task context
	SessionRegistry // How an AppInstance resolves symbols and types

	// Called when this session is newly opened to set up the SessionRegistry
	InitSessionRegistry(symTable symbol.Table)

	// Returns the running AssetPublisher instance for this session.
	AssetPublisher() AssetPublisher

	// Returns info about this user and session
	LoginInfo() Login

	// Sends an unnamed attr to the client's session controller.
	// If the reqID == 0, the attr is sent to the client's session controller.
	PushMetaAttr(val ElemVal, reqID uint64) error

	// Gets the currently running AppInstance for an AppID.
	// If the requested app is not running and autoCreate is set, a new instance is created and started.
	GetAppInstance(appID UID, autoCreate bool) (AppInstance, error)
}

// SessionRegistry manages a HostSession's symbol and type definitions.
type SessionRegistry interface {
	ClientSymbols() symbol.Table
	
	// Translates a native symbol ID to a client symbol ID, returning false if not found.
	NativeToClientID(nativeID uint32) (clientID uint32, found bool)

	// Registers an ElemVal as a prototype under its element type name..
	// This and ResolveAttrSpec() allow NewElemVal() to work.
	RegisterElemType(prototype ElemVal) error

	// Registers a block of symbol, attr, cell, and selector definitions for a client.
	RegisterDefs(defs *RegisterDefs) error

	// Resolves an AttrSpec into useful symbols, auto-registering the AttrSpec as needed.
	// Typically used during AppInstance.OnNew() to get the AttrIDs that correspond to the AttrSpecs it will send later.
	//
	// If native === true, the spec is resolved with native symbols (vs client symbols).
	//
	// See AttrSpec docs.
	ResolveAttrSpec(attrSpec string, native bool) (AttrSpec, error)

	// Resolves a CellSpec (a cell attr schema) into symbol IDs, auto-registering each as needed.
	// Called by apps to resolve cell types it supports, getting a CellSpec ID to stamp cells it pushes to clients.
	//
	// See CellSpec docs.
	ResolveCellSpec(cellSpec string) (CellDef, error)

	// Instantiates an attr element value for an AttrID -- typically followed by ElemVal.Unmarshal()
	NewAttrElem(attrDefID uint32, native bool) (ElemVal, error)
}

// Registry maps an app ID to an AppModule.    It is safe to access from multiple goroutines.
type Registry interface {

	// Registers an ElemVal as a prototype under its AttrElemType (also a valid AttrSpec type expression).
	// If an entry already exists (common for a type used by multiple apps), an error is returned and is a no-op.
	// This and ResolveAttrSpec() allow NewElemVal() to work.
	RegisterElemType(prototype ElemVal)

	// When a HostSession creates a new SessionRegistry(), this populates it with its registered ElemTypes.
	ExportTo(dst SessionRegistry) error

	// Registers an app by its UUID, URI, and schemas it supports.
	RegisterApp(app *AppModule) error

	// Looks-up an app by UUID
	GetAppByUID(appUID UID) (*AppModule, error)

	// Selects the app that best matches an invocation string.
	GetAppForInvocation(invocation string) (*AppModule, error)
}

// NewRegistry returns a new Registry
func NewRegistry() Registry {
	return newRegistry()
}

type SIRange struct {
	Lo uint64
	Hi uint64
}

type CellReq interface {
	PinID() CellID
	URL() *url.URL
	URLPath() []string
	String() string
}

// FUTURE: type CellTID [2]uint64
type CellID uint64

type PbValue interface {
	Size() int
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Unmarshal(dAtA []byte) error
}

type ElemVal interface {

	// Returns the element type name (a degenerate AttrSpec).
	TypeName() string

	// Marshals this ElemVal to a buffer, reallocating if needed.
	MarshalToBuf(dst *[]byte) error

	// Unmarshals and merges value state from a buffer.
	Unmarshal(src []byte) error

	// Creates a default instance of this same ElemVal type
	New() ElemVal
}

type AttrElem struct {
	Val    ElemVal
	SI     int64
	AttrID uint32
}

type AttrDef struct {
	Client AttrSpec
	Native AttrSpec
}

type CellDef struct {
	ClientDefID uint32    // READ-ONLY
	NativeDefID uint32    // READ-ONLY
	CommonAttrs []AttrDef // READ-ONLY
	PinnedAttrs []AttrDef // READ-ONLY
}

// AttrTx?
// FUTURE: write custom MarshalToBuf / Unmarshal that write multiple AttrElems to a single Msg
type AttrBatch struct {
	Target CellID
	Attrs  []AttrElem
}
