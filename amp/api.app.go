package amp

import (
	"net/url"

	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
	"github.com/amp-3d/amp-sdk-go/stdlib/task"
)

var (
	AppSpec  = tag.FormSpec(tag.Spec{}, "amp.app")
	AttrSpec = tag.FormSpec(AppSpec, "attr")
	//MetaAttrSpec = FormSpec(AttrSpec, "meta")
)

// App is how an app module is registered with an amp.Host so it can be invoked.~
//
// An App is invoked by a client or other app via the app's Tag or URI.
type App struct {

	// TagSpec identifies this app with form "{AppNameID}.{FamilyID}.{PublisherID}" -- e.g. "filesys.hosting.arcspace.systems"
	//   - PublisherID: typically the domain name of the publisher of this app -- e.g. "arcspace.systems"
	//   - FamilyID:    encompassing namespace ID used to group related apps (no spaces or punctuation)
	//   - AppNameID:   identifies this app within its parent family and domain (no spaces or punctuation)
	//
	AppSpec      tag.Spec  // Universally unique and persistent ID for this module (and the module's "home" planet if present)
	Desc         string    // Human-readable description of this app
	Version      string    // "v{MajorVers}.{MinorID}.{RevID}"
	Dependencies []tag.ID // Module Tags this app may access
	Invocations  []string  // Additional aliases that invoke this app
	AttrDecl     []string  // Attrs to be resolved and registered with a HostSession

	// NewAppInstance is the entry point for an App.
	// Called when an App is invoked on an active User session and is not yet running.
	NewAppInstance func() AppInstance
}

// AppContext is provided by the amp runtime to an AppInstance for support and context.
type AppContext interface {
	task.Context          // Allows select{} for graceful handling of app shutdown
	AssetPublisher        // Allows an app to publish assets for client consumption
	Session() HostSession // Access to underlying Session

	// Returns the absolute file system path of the app's local read-write directory.
	// This directory is scoped by the app's Tag.
	LocalDataPath() string

	// Gets the named attribute from the user's home planet -- used high-level app settings.
	// The attr is scoped by both the app Tag so key collision with other users or apps is not possible.
	// This is how an app can store and retrieve settings.
	GetAppAttr(attrSpec tag.Spec, dst ElemVal) error

	// Write analog for GetAppAttr()
	PutAppAttr(attrSpec tag.Spec, src ElemVal) error
}

// AppInstance is implemented by an App and invoked by amp.Host responding to a client pin request.
type AppInstance interface {
	AppContext // amp's app runtime support exposed

	// Instantiation callback made immediately after App.NewAppInstance() -- typically resolves app-specific type specs.
	OnNew(this AppContext) error
	
	// aka CreateNewCell with feed template, 
	// Installs or "mints" the attributes onto a target
	MintNewFeed(managedTarget Pin, template FeedGenesis) error

	// Pins the requested cell or URL
	// If from != nil, it is the invoking context of the request.
	// If from == nil, there is no parent context and the request is typically a URL.
	NewPin(from Pin, req PinOp) (Pin, error)

	// Handles a meta message sent to this app, which could be any attr type.
	HandleURL(*url.URL) error

	// Called exactly once when an app is signaled to close.
	OnClosing()
}

// Pin is how your app encapsulates a pinned URI to the host runtime and thus clients.
type Pin interface {

	// Apps spawn a Pin as a child task.Context of amp.AppContext.Context or as a child of another Pin.
	// This means an AppContext contains all its Pins and thus Close() will close all Pins.
	Context() task.Context

	// Pins the requested cell from the context of this Pin (typically a child cell).
	// Called from within AppInstance.NewPin() since an app may require preparation, such a renewing a session.
	PinSub(req PinOp) (Pin, error)

	// Pushes this cell and child cells to the client.
	// Exits when:
	//   - ctx.Closing() is signaled, or
	//   - state has been pushed to the client and no more updates are possible, or
	//   - state has been pushed initially but PinFlags_CloseOnSync is set, or
	//   - an error occurs
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

	// Returns the element type name (a scalar TagSpec).
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
	refCount  int32  // see AddRef() / ReleaseRef()
	Ops       []TxOp // ordered operations to perform on the target
	DataStore []byte // marshalled data store for Ops serialized data
}

// TxOp is an atomic operation on a target cell and is a unit of change (or message) for any target.
// Values are typically LSM sorted, so use low order bytes before high order bytes.
// Note that x0 is the most significant and x2 is least significant bytes.
type TxOp struct {
	OpCode   TxOpCode
	TargetID tag.ID // Target to operate on
	AttrID   tag.ID // Attribute to operate on
	SI       tag.ID // Index of the data being mutated
	Height   uint64  // Content revision "height"
	Hash     uint64  // hash of genesis tag from of the Tx containing this op
	DataLen  uint64  // Length of data in TxMsg.DataStore
	DataOfs  uint64  // Offset into TxMsg.DataStore
}

type AttrDef struct {
	tag.Spec
	Prototype ElemVal
}

// TagSpecID is a tag for the canonic string representation of an TagSpec
// Leading bits are reserved to express pin detail level or layer.
type TagSpecID tag.ID
