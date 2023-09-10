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
	elemTypes    []AttrElemVal
}

func (reg *registry) RegisterElemType(prototype AttrElemVal) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.elemTypes = append(reg.elemTypes, prototype)
}

func (reg *registry) ExportTo(dst SessionRegistry) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	for _, elemType := range reg.elemTypes {
		if err := dst.RegisterElemType(elemType); err != nil {
			return err
		}
	}
	return nil
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
