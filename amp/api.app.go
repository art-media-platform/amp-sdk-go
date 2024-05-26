package amp

import (
	"github.com/amp-3d/amp-sdk-go/stdlib/media"
	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
	"github.com/amp-3d/amp-sdk-go/stdlib/task"
)

// App is how an app module is registered with an amp.Host so it can be invoked.~
//
// An App is invoked by a client or other app via the app's Tag or URI.
type App struct {

	// tag.Spec identifies this app with form "amp.app.{PublisherID}.{FamilyID}.{AppNameID}" -- e.g. "amp.app.os.filesys.posix"
	//   - PublisherID: typically the domain name of the publisher of this app -- e.g. "arcspace.systems"
	//   - FamilyID:    encompassing namespace ID used to group related apps (no spaces or punctuation)
	//   - AppNameID:   identifies this app within its parent family and domain (no spaces or punctuation)
	//
	AppSpec      tag.Spec // Universally unique and persistent ID for this module (and the module's "home" planet if present)
	Desc         string   // Human-readable description of this app
	Version      string   // "v{MajorVers}.{MinorID}.{RevID}"
	Dependencies []tag.ID // Module Tags this app may access
	Invocations  []string // Additional aliases that invoke this app
	AttrDecl     []string // Attrs to be resolved and registered with a HostSession

	// NewAppInstance is the entry point for an App.
	// Called when an App is first invoked on an active User session and is not yet running.
	// Blocks minimally and returns quickly.
	NewAppInstance func(ctx AppContext) (AppInstance, error)
}

// AppContext is provided by the amp runtime to an AppInstance for support and context.
type AppContext interface {
	task.Context          // Allows select{} for graceful handling of app shutdown
	media.Publisher       // Allows an app to publish assets for client consumption
	Session() HostSession // Access to underlying Session

	// Returns the absolute file system path of the app's local read-write directory.
	// This directory is scoped by the app's Tag.
	LocalDataPath() string

	// Gets the named attribute from the user's home planet -- used high-level app settings.
	// The attr is scoped by both the app Tag so key collision with other users or apps is not possible.
	// This is how an app can store and retrieve its settings for the current user.
	GetAppAttr(attrSpec tag.ID, dst ElemVal) error

	// Write analog for GetAppAttr()
	PutAppAttr(attrSpec tag.ID, src ElemVal) error
}

// Pinner is characterized by the ability to emit Pins.
type Pinner interface {

	// Creates or finds Pin for the given request.
	ServeRequest(req Requester) (Pin, error)
}

// AppInstance is implemented by an App and invoked by amp.Host responding to a client pin request.
type AppInstance interface {
	AppContext
	Pinner

	// Validates a request and performs any needed setup.
	// This åis a chance for an app to perform operations such refreshing an auth token.
	// Following this call, ServeRequest() is called.
	MakeReady(req Requester) error

	// Called exactly once when this App closes
	OnClosing()
}

// Pin is a attribute state connection to an amp.App.
// The handling App is responsible for updating the Requester with state changes as requested.
type Pin interface {
	Pinner

	// Apps start a Pin as a child Context of amp.AppContext.Context or as a child of another Pin.
	// This means an AppContext contains all its Pins and thus Close() will close all Pins (and child requests).
	// This is used to know if a request is still being served and to close it if needed.
	Context() task.Context

	// Pushes state until all requested attrs are synced.
	//
	// Exits when:
	//   - to.Closing() is signaled, or
	//   - state has been pushed to the client and no more updates are possible, or
	//   - state has been pushed initially but PinFlags_CloseOnSync is set, or
	//   - an error occurs
	//ServeRequest(to Requester) error

	// // Asks this Pin to handle a client request.
	// //
	// // Usually a Pin starts a child context to perform blocking work able to serve the given request,
	// // and returns it wrapped in a RequestHandler.
	// //
	// // If (nil, nil) is returned, the app handled the request immediately.
	// HandleRequest(op Request) (RequestHandler, error)

	// Processes, queues, or otherwise handles a changeset sent to this Pin.
	// Concurrent friendly -- each client request having a tx to submit calls this method from its own goroutine.
	// tx is READ ONLY.
	// CommitTx(tx *TxMsg) error
}

// // Wraps the task an App issues in response to Pin.HandleRequest()
// type RequestHandler interface {

// 	// The owning Pin that spawned this RequestHandler.
// 	ParentPin() Pin

// 	// This RequestHandler's context -- a child context of the associated Pin.
// 	// Used to know if a request is still being served and to close it if needed.
// 	Context() task.Context

// }

// Serialization abstraction
type PbValue interface {
	Size() int
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Unmarshal(dAtA []byte) error
}

// ElemVal wraps cell attribute element name and serialization.
type ElemVal interface {

	// Returns the element type name (a scalar tag.Spec).
	ElemTypeName() string // TODO: use generics for default name

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
	refCount  int32  // see AddRef() / ReleaseRef()
	Ops       []TxOp // ordered operations to perform on the target
	DataStore []byte // marshalled data store for Ops serialized data
}

// TxOp is an atomic operation on a target cell and is a unit of change (or message) for any target.
// Values are typically LSM sorted, so use low order bytes before high order bytes.
// Note that x0 is the most significant and x2 is least significant bytes.
type TxOp struct {
	OpCode   TxOpCode
	FromID   tag.ID // Directed link -- FromID to TargetID
	TargetID tag.ID // Target to operate on
	AttrID   tag.ID // Attribute to operate on
	SI       tag.ID // Index of the data being mutated
	Height   uint64 // parent tx height -- see "height" in https://peerlinks.io/protocol.html
	Hash     uint64 // parent tx hash (conflict resolution)
	DataLen  uint64 // Length of data in TxMsg.DataStore
	DataOfs  uint64 // Offset into TxMsg.DataStore
}

type AttrDef struct {
	tag.Spec
	Prototype ElemVal
}

// TagSpecID is a tag for the canonic string representation of an tag.Spec
// Leading bits are reserved to express pin detail level or layer.
type TagSpecID tag.ID
