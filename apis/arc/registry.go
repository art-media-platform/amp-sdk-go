package arc

import (
	"strings"
	"sync"
)

func newRegistry() Registry {
	return &registry{
		appsByUID:    make(map[UID]*App),
		appsByInvoke: make(map[string]*App),
	}
}

// Implements arc.Registry
type registry struct {
	mu           sync.RWMutex
	appsByUID    map[UID]*App
	appsByInvoke map[string]*App
	elemDefs     map[AttrUID]AttrElemVal
	attrDefs     map[AttrUID]AttrElemVal
}

func (reg *registry) RegisterElemType(prototype AttrElemVal) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	attrID := FormAttrID(prototype.ElemTypeName())
	reg.elemDefs[attrID] = prototype
}

func (reg *registry) NewAttrElem(attrID AttrUID) (AttrElemVal, error) {

	// Often, an attrID will be a unnamed scalar ("degenerate") attr, which means we can get the elemDef directly.
	// This is also essential during bootstrapping when the client sends a RegisterDefs is not registered yet.
	elemDef, exists := reg.elemDefs[attrID]
	if !exists {
		attrDef, exists := reg.attrDefs[attrID]
		if !exists {
			return nil, arc.ErrCode_DefNotFound.Errorf("NewAttrElem: attr DefID %v not found", attrID)
		}
		elemDef, exists = reg.elemDefs[attrDef.Native.ElemType]
		if !exists {
			return nil, arc.ErrCode_DefNotFound.Errorf("NewAttrElem: elemTypeID %v not found", attrDef.Client.ElemType)
		}
	}
	return elemDef.prototype.New(), nil
}


// Implements arc.Registry
func (reg *registry) RegisterApp(app *App) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	if strings.ContainsRune(app.AppID, '/') ||
		strings.ContainsRune(app.AppID, ' ') ||
		strings.Count(app.AppID, ".") < 2 {

		// Reject if URI does not conform to standards for App.AppURI
		return ErrCode_BadSchema.Errorf("illegal app ID: %q", app.AppID)
	}

	reg.appsByUID[app.UID] = app

	for _, invok := range app.Invocations {
		if invok != "" {
			reg.appsByInvoke[invok] = app
		}
	}

	// invoke by full app ID
	reg.appsByInvoke[app.AppID] = app

	// invoke by first component of app ID
	appPos := strings.Index(app.AppID, ".")
	appName := app.AppID[0:appPos]
	reg.appsByInvoke[appName] = app

	return nil
}

// Implements arc.Registry
func (reg *registry) GetAppByUID(appUID UID) (*App, error) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()

	app := reg.appsByUID[appUID]
	if app == nil {
		return nil, ErrCode_AppNotFound.Errorf("app not found: %s", appUID)
	} else {
		return app, nil
	}
}

// Implements arc.Registry
func (reg *registry) GetAppForInvocation(invocation string) (*App, error) {
	if invocation == "" {
		return nil, ErrCode_AppNotFound.Errorf("missing app invocation")
	}

	reg.mu.RLock()
	defer reg.mu.RUnlock()

	app := reg.appsByInvoke[invocation]
	if app == nil {
		return nil, ErrCode_AppNotFound.Errorf("app not found for invocation %q", invocation)
	}
	return app, nil
}
