package arc

import (
	"fmt"
	"strings"
	"sync"
)

func newRegistry() Registry {
	return &registry{
		appsByUID:   make(map[UID]*AppModule),
		appsByModel: make(map[string]*AppModule),
	}
}

// Implements arc.Registry
type registry struct {
	mu          sync.RWMutex
	appsByUID   map[UID]*AppModule
	appsByModel map[string]*AppModule // TODO: index by symbol.ID and move into TypeRegistry?
}

// Implements arc.Registry
func (reg *registry) RegisterApp(app *AppModule) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	// Reject if URI does not conform to standards for AppModule.AppURI
	if len(strings.Split(app.URI, "/")) != 4 {
		return ErrInvalidAppURI
	}
	reg.appsByUID[app.UID] = app

	for ID := range app.DataModels.ModelsByID {
		if ID != "" {
			reg.appsByModel[ID] = app
		}
	}

	return nil
}

// Implements arc.Registry
func (reg *registry) GetAppByUID(appUID UID) (*AppModule, error) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()

	app := reg.appsByUID[appUID]

	fmt.Printf("app: %v\n", app)
	if app == nil {
		return nil, ErrCode_AppNotFound.Errorf("app not found: %v", appUID)
	} else {
		return app, nil
	}
}

// Implements arc.Registry
func (reg *registry) GetAppByURI(appURI string) (*AppModule, error) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()

	for _, app := range reg.appsByUID {
		if app.URI == appURI {
			return app, nil
		}
	}
	return nil, ErrCode_AppNotFound.Errorf("app not found: %q", appURI)
}

// Implements arc.Registry
func (reg *registry) GetAppForSchema(schema *AttrSchema) (*AppModule, error) {
	if schema == nil {
		return nil, ErrCode_AppNotFound.Errorf("missing schema")
	}

	reg.mu.RLock()
	defer reg.mu.RUnlock()

	app := reg.appsByModel[schema.CellDataModel]
	if app == nil {
		return nil, ErrCode_AppNotFound.Errorf("app not found for schema: %s", schema.SchemaDesc())
	}

	return app, nil
}
