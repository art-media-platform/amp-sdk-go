package arc

import (
	"net/url"

	"github.com/arcspace/go-arc-sdk/stdlib/task"
)

// AppModule declares a 3rd-party module this is registered with an archost.
//
// An app can be invoked by:
//   - a client pinning a cell with a data model that the app handles
//   - a client or other app invoking its UID or URI directly
type AppModule struct {

	// AppID identifies this app with form "{AppNameID}.{FamilyID}.{PublisherID}" -- e.g. "filesys.amp.arcspace.systems"
	//   - PublisherID: typically the domain name of the publisher of this app -- e.g. "arcspace.systems"
	//   - FamilyID:    encompassing namespace ID used to group related apps and content (no spaces or punctuation)
	//   - AppNameID:   uniquely identifies this app within its parent family and domain (no spaces or punctuation)
	//
	// AppID form is consistent of a URL domain name (and subdomains).
	AppID        string
	UID          UID      // Universally unique and persistent ID for this module (and the module's "home" planet if present)
	Desc         string   // Human-readable description of this app
	Version      string   // "v{MajorVers}.{MinorID}.{RevID}"
	Dependencies []UID    // Module UIDs this app may access
	Invocations  []string // Additional aliases that invoke this app
	AttrDecl     []string // Attrs to be resolved and registered with the HostSession -- get the registered

	// Called when an App is invoked on an active User session and is not yet running.
	NewAppInstance func() AppInstance
}

// AppContext hosts is provided by the arc runtime and hosts an AppInstance.
//
// An AppModule retains the AppContext it is given via NewAppInstance() for:
//   - archost operations (e.g. resolve type schemas, publish assets for client consumption -- see AppContext)
//   - having a context to select{} against (for graceful shutdown)
type AppContext interface {
	task.Context
	AssetPublisher        // Allows an app to publish assets for client consumption
	Session() HostSession // Access to underlying Session

	// Returns the absolute fs path of the app's local state directory.
	// This directory is scoped by the app's UID and is unique to this app instance.
	LocalDataPath() string

	// Atomically issues a new and unique ID that will remain globally unique for the duration of this session.
	// An ID may still expire, go out of scope, or otherwise become meaningless.
	IssueCellID() CellID

	// Allows an app resolve attrs by name, etc

	// Gets the named cell and attribute from the user's home planet -- used high-level app settings.
	// The attr is scoped by both the app UID so key collision with other users or apps is not possible.
	GetAppCellAttr(attrSpec string, dst ElemVal) error

	// Write analog for GetAppCellAttr()
	PutAppCellAttr(attrSpec string, src ElemVal) error
}

// AppInstance is implemented by an arc app (AppModule)
type AppInstance interface {
	AppContext // An app instance implies an underlying host AppContext

	// Callback made immediately after AppModule.NewAppInstance() -- typically resolves app-specific type specs.
	OnNew(ctx AppContext) error

	// Celled when the app is pin the cell IAW with the given request.
	// If parent != nil, this is the context of the request.
	// If parent == nil, this app was invoked without out a parent cell / context.
	PinCell(parent PinnedCell, req CellReq) (PinnedCell, error)

	// Handles a meta message sent to this app, which could be any attr type.
	HandleURL(*url.URL) error

	// Called exactly once if / when an app is signaled to close.
	OnClosing()
}

type CellTx struct {
	Attrs []AttrElem // Attrs to merge -- AttrIDs are NATIVE (not client) IDs
}

// PinnedCell is how your app encapsulates a pinned cell to the archost runtime and thus clients.
type PinnedCell interface {

	// Apps spawn a PinnedCell as a child task.Context of arc.AppContext.Context or as a child of another PinnedCell.
	// This means an AppContext contains all its PinnedCells and thus Close() will close all PinnedCells.
	Context() task.Context

	// Pins the requested cell (typically a child cell).
	PinCell(req CellReq) (PinnedCell, error)
	//PinChild(req CellReq) (PinnedCell, error)

	//App() AppInstance

	// A PinnedCell is also a closure / context for incoming ResolveCell requests.
	// ResolveCell resolves the given request to a PinnedCell, potentially pinning the cell as needed.
	// After returned PinnedCell will then have PushState() to:
	///    - push the cell's state to the client
	//     - push cell state updates as needed
	//     - have <-ctx.Closing() to use alongside blocking operations.
	//
	//ResolveRequest(req CellReq, parentApp AppInstance) (Cell, error)

	// Pushes this cell and child cells to the client state is called.
	// Exits when any of the following occur:
	//   - ctx.Closing() is signaled,
	//   - a fatal error is encountered, or
	//   - state has been pushed to the client AND ctx.MaintainSync() == false
	PushState(ctx PinContext) error

	// Merges a set of changes into this cell.
	MergeTx(tx CellTx) error
}

// PinContext wraps a client request to receive a cell's state / updates.
type PinContext interface {
	task.Context // Started as a CHILD of the arc.PinnedCell returned by AppInstance.PinCell()

	// If true, symbol IDs are native (not client).
	UsingNativeSymbols() bool

	// If true:   PinnedCell.PushState() should block until PinContext.Closing() is signaled.
	// If false:  PinnedCell.PushState() should exit once state has been pushed.
	MaintainSync() bool

	// Low-level push of a Msg to the client, returning true if the msg was sent (false if the client has been closed).
	PushMsg(msg *Msg) bool

	// Parent app of the cell associated with this context
	App() AppContext
}
