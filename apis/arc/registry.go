package arc

import "sync"

// Implements arc.Registry
type registry struct {
	mu          sync.RWMutex
	appsByID    map[string]*AppModule
	appsByModel map[string]*AppModule // TODO: index by symbol.ID and move into TypeRegistry?
}

// Implements arc.Registry
func (reg *registry) RegisterApp(app *AppModule) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	if app.AppID == "" {
		return ErrInvalidAppID
	}
	reg.appsByID[app.AppID] = app

	for ID := range app.DataModels.ModelsByID {
		if ID != "" {
			reg.appsByModel[ID] = app
		}
	}

	return nil
}

// Implements arc.Registry
func (reg *registry) GetAppByID(appID string) (*AppModule, error) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	app := reg.appsByID[appID]
	if app == nil {
		return nil, ErrCode_AppNotFound.Errorf("App not found: %s", appID)
	} else {
		return app, nil
	}
}

// Implements arc.Registry
func (reg *registry) SelectAppForSchema(schema *AttrSchema) (*AppModule, error) {
	if schema == nil {
		return nil, ErrCode_AppNotFound.Errorf("missing schema")
	}

	reg.mu.RLock()
	defer reg.mu.RUnlock()

	if schema.ScopeID != ImpliedScopeForDataModel {
		app := reg.appsByID[schema.ScopeID]
		if app != nil {
			return app, nil
		}
	}

	app := reg.appsByModel[schema.CellDataModel]
	if app == nil {
		return nil, ErrCode_AppNotFound.Errorf("App not found for schema: %s", schema.SchemaDesc())
	}

	return app, nil
}
