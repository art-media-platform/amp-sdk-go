package arc

// AppModule declares a 3rd-party module this is registered with an archost.
// An app is invoked by its AppID directly or a client requesting a data model this app declares to support.
type AppModule struct {
	AppID        string       // "{domain_name}/{module_identifier}"
	Version      string       // "v{MAJOR}.{MINOR}.{REV}"
	Dependencies []string     // App URIs this app may need to invoke
	DataModels   DataModelMap // Data models that this app defines and handles.

	// Called when an App is invoked on an active User session and is not yet running.
	// Msg processing is blocked until this returns -- only AppRuntime calls should block.
	NewAppInstance func(ctx AppContext) (AppRuntime, error)
}

// AppRuntime is a runtime-furnished container context for an AppModule instance.
type AppRuntime interface {
	CellContext

	// Pre: msg.Op == MsgOp_MetaMsg
	HandleMetaMsg(msg *Msg) (handled bool, err error)

	// Called exactly once when an app is signaled to close.
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

// PinReq?
// See api.support.go for CellReq helper methods such as PushMsg.
type CellReq struct {
	CellSub

	Args          []*KwArg      // Client-set args (typically used when pinning a root where CellID is not known)
	PinCell       CellID        // Client-set cell ID to pin (or 0 if Args sufficient).  Use req.Cell.ID() for the resolved CellID.
	ContentSchema *AttrSchema   // Client-set schema specifying the cell attr model for the cell being pinned.
	ChildSchemas  []*AttrSchema // Client-set schema(s) specifying which child cells (and attrs) should be pushed to the client.
}

// See AttrSchema.ScopeID in arc.proto
const ImpliedScopeForDataModel = "."

// type DataModel map[string]*Attr
type DataModel struct {
}

type DataModelMap struct {
	ModelsByID map[string]DataModel // Maps a data model ID to a data model definition
}


// A Registry is a mapping of IDs to app modules.  It is safe to access from multiple goroutines.
type Registry interface {

	// Registers an app for invocation
	RegisterApp(app *AppModule) error

	// Looks-up an app by an ID / URI
	GetAppByID(appID string) (*AppModule, error)

	// Selects the app that best handles the given schema
	SelectAppForSchema(schema *AttrSchema) (*AppModule, error)
}

// NewRegistry returns a new Registry
func NewRegistry() Registry {
	return &registry{
		appsByID:    make(map[string]*AppModule),
		appsByModel: make(map[string]*AppModule),
	}
}
