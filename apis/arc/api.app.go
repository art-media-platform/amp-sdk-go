package arc

import (
	"reflect"

	"github.com/arcspace/go-arc-sdk/stdlib/process"
)

// AppModule declares a 3rd-party module this is registered with an archost.
//
// An app can be invoked by:
//   - a client pinning a cell with a data model that the app handles
//   - a client or other app invoking its UID or URI directly
type AppModule struct {

	// URI identifies this app using the form "{PublisherID}/{FamilyID}/{AppNameID}/v{MajorVers}" -- e.g. "arcspace.systems/amp/filesys/v1"
	//   - PublisherID: typically a publicly registered domain name of the publisher of this app
	//   - FamilyID:    encompassing namespace ID used to group related apps and content
	//   - AppNameID:   uniquely identifies this app within its parent family and domain.
	//   - MajorVers:   an integer starting with 1 that is incremented when a breaking change is made to the app's API.
	URI          string
	UID          UID          // Universally unique and persistent ID for this module
	Desc         string       // Human-readable description of this app
	Version      string       // "v{MajorVers}.{MinorID}.{RevID}"
	Dependencies []UID        // Module UIDs this app may access via GetAppContext()
	DataModels   DataModelMap // Data models that this app defines and handles.

	// Called when an App is invoked on an active User session and is not yet running.
	NewAppInstance func(ctx AppContext) (AppRuntime, error)
}

// AppContext encapsulates execution of an AppRuntime.
//
// An AppModule retains the AppContext it is given via NewAppInstance() for:
//   - archost operations (e.g. resolve type schemas, publish assets for client consumption -- see AppContext)
//   - having a context to select{} against (for graceful shutdown)
type AppContext interface {
	process.Context // Each app instance has a process.Context
	AssetPublisher  // Allows an app to publish assets for client consumption
	User() User     // Access to user operations and io
	CellPinner      // How to pin root cells

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

// AppRuntime is a runtime-furnished container context for an AppModule instance.
type AppRuntime interface {
	CellPinner

	// Pre: msg.Op == MsgOp_MetaMsg
	HandleMetaMsg(msg *Msg) (handled bool, err error)

	// Called exactly once if / when an app is signaled to close.
	OnClosing()
}

type TypeRegistry interface {

	// Resolves and then registers each given def, returning the resolved defs in-place if successful.
	//
	// Resolving a AttrSchema means:
	//    1) all name identifiers have been resolved to their corresponding host-dependent symbol IDs.
	//    2) all "InheritsFrom" types and fields have been "flattened" into the form
	//
	// See MsgOp_ResolveAndRegister
	ResolveAndRegister(defs *Defs) error

	// Returns the resolved AttrSchema for the given cell type ID.
	GetSchemaByID(schemaID int32) (*AttrSchema, error)
}

// CellContext?
// CellInvoker? 
// CellInvoker is a runtime-furnished container context for a pinned Cell.
type CellClient interface {
	process.Context
	
	// Returns params given by the client for the request bei
	Params() CellReq 
	
	// Fetches the args given by the client when pinning this cell. 
	KwArgs(name string) []*KwArg
	
	// Sets msg.ReqID and pushes the given msg to client, blocking until "complete" (queued) or canceled.
	// This msg is reclaimed downstream from here, so it should be considered released / inaccessible after this call.
	PushMsg(msg *Msg) error
}

// CellContext?
// PinReq?
// See api.support.go for CellReq helper methods such as PushMsg.
type CellReq struct {
	CellSub

	Args          []*KwArg      // Client-set args (typically used when pinning a root where CellID is not known)
	PinCell       CellID        // Client-set cell ID to pin (or 0 if Args sufficient).  Use req.Cell.ID() for the resolved CellID.
	ContentSchema *AttrSchema   // Client-set schema specifying the cell attr model for the cell being pinned.
	ChildSchemas  []*AttrSchema // Client-set schema(s) specifying which child cells (and attrs) should be pushed to the client.
}

// type DataModel map[string]*Attr
type DataModel struct {
	// TODO
}

type DataModelMap struct {
	ModelsByID map[string]DataModel // Maps a data model ID to a data model definition
}

// Registry maps an app ID to an AppModule.    It is safe to access from multiple goroutines.
type Registry interface {

	// Registers an app by its UUID, URI, and schemas it supports.
	RegisterApp(app *AppModule) error

	// Looks-up an app by UUID
	GetAppByUID(appUID UID) (*AppModule, error)

	// Looks-up an app by URI
	GetAppByURI(appURI string) (*AppModule, error)

	// Selects the app that best handles the given schema
	GetAppForSchema(schema *AttrSchema) (*AppModule, error)
}

// NewRegistry returns a new Registry
func NewRegistry() Registry {
	return newRegistry()
}
