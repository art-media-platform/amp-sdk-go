package arc

import (
	"github.com/arcspace/go-arc-sdk/stdlib/task"
)

// AppModule declares a 3rd-party module this is registered with an archost.
//
// An app can be invoked by:
//   - a client pinning a cell with a data model that the app handles
//   - a client or other app invoking its UID or URI directly
type AppModule struct {

	// AppID identifies this app with form "v{MajorVers}.{AppNameID}.{FamilyID}.{PublisherID}" -- e.g. "v1.filesys.amp.arcspace.systems"
	//   - PublisherID: typically the domain name of the publisher of this app -- e.g. "arcspace.systems"
	//   - FamilyID:    encompassing namespace ID used to group related apps and content (no spaces or punctuation)
	//   - AppNameID:   uniquely identifies this app within its parent family and domain (no spaces or punctuation)
	//   - MajorVers:   an integer starting with 1 that is incremented when a breaking change is made to the app's API.
	//
	// AppID form is consistent of a URL domain name (and subdomains).
	AppID        string
	UID          UID      // Universally unique and persistent ID for this module (and the module's "home" planet if present)
	Desc         string   // Human-readable description of this app
	Version      string   // "v{MajorVers}.{MinorID}.{RevID}"
	Dependencies []UID    // Module UIDs this app may access via GetAppContext()
	Invocations  []string // Additional aliases that invoke this app
	AttrDecl     []string // Attrs to be resolved and registered with the HostSession -- get the registered

	// Called when an App is invoked on an active User session and is not yet running.
	NewAppInstance func() AppInstance
}

// AppContext encapsulates execution of an AppInstance and is provided by the archost runtime.
//
// An AppModule retains the AppContext it is given via NewAppInstance() for:
//   - archost operations (e.g. resolve type schemas, publish assets for client consumption -- see AppContext)
//   - having a context to select{} against (for graceful shutdown)
type AppContext interface {
	task.Context          // Each app instance has a task.Context and is a child of the user's session
	AssetPublisher        // Allows an app to publish assets for client consumption
	CellResolver          // How to resolve App cells by URI
	SessionRegistry       // How to resolve Attrs by name and type
	Session() HostSession // Access to host & user ops

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
	AppContext

	// Callback made immediately after AppModule.NewAppInstance() -- typically resolves app-specific type specs.
	OnNew(ctx AppContext) error

	// Handles a meta message sent to this app, which could be any attr type.
	HandleMetaAttr(attr AttrElem) (handled bool)

	// Called exactly once if / when an app is signaled to close.
	OnClosing()
}

// PinnedCell is how your app encapsulates a pinned cell to the archost runtime and thus clients.
type PinnedCell interface {

	// Apps spawn a PinnedCell as a child task.Context of arc.AppContext.Context or as a child of another PinnedCell.
	// This means an AppContext contains all its PinnedCells and thus Close() will close all PinnedCells.
	Context() task.Context

	// A PinnedCell is a closure / context for incoming ResolveCell requests.
	CellResolver

	// Pushes this cell and child cells to the client state is called.
	// Exits when any of the following occur:
	//   - ctx.Closing() is signaled,
	//   - a fatal error is encountered, or
	//   - state has been pushed to the client AND ctx.MaintainSync() == false
	PushState(ctx PinContext) error
}

// PinContext wraps a client request to receive a cell's state / updates.
type PinContext interface {
	task.Context // Started as a CHILD of the arc.PinnedCell returned by App.PinCell()

	// If true:   PinnedCell.PushState() should block until PinContext.Closing() is signaled.
	// If false:  PinnedCell.PushState() should exit once state has been pushed.
	MaintainSync() bool

	// Low-level push of a Msg to the client, returning true if the msg was sent (false if the client has been closed).
	PushMsg(msg *Msg) bool

	// Parent app of the cell associated with this context
	App() AppContext
}
