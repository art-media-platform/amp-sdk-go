package arc

import (
	"strings"
	"sync"
)

func newRegistry() Registry {
	return &registry{
		appsByUID:    make(map[UID]*App),
		appsByInvoke: make(map[string]*App),
		attrDefs:     make(map[AttrUID]AttrDef),
	}
}

// Implements arc.Registry
type registry struct {
	mu           sync.RWMutex
	appsByUID    map[UID]*App
	appsByInvoke map[string]*App
	attrDefs     map[AttrUID]AttrDef
}

type AttrDef struct {
	AttrSpec
	Prototype AttrElemVal
}

func (attr *AttrDef) AttrID() AttrUID {
	return AttrUID(FormUID(attr.AttrUIDx0, attr.AttrUIDx1))
}


func (reg *registry) RegisterElemType(prototype AttrElemVal) {
	attrSpec := prototype.ElemTypeName()
	elemTypeUID := FormAttrID(attrSpec)

	spec := AttrSpec{
		AsCanonicString: attrSpec,
		ElemTypeUIDx0:   elemTypeUID[0],
		ElemTypeUIDx1:   elemTypeUID[1],
		AttrUIDx0:       elemTypeUID[0],
		AttrUIDx1:       elemTypeUID[1],
	}
	
	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.attrDefs[elemTypeUID] = AttrDef{
		AttrSpec: spec,
		Prototype: prototype,
	}
}


func (reg *registry) RegisterAttr(attrDef *AttrDef) error {
	spec, err := FormAttrSpec(attrDef.AsCanonicString)
	if err != nil {
		return err
	}
	
	var attrID AttrUID
	attrID[0] = spec.AttrUIDx0
	attrID[1] = spec.AttrUIDx1

	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.attrDefs[spec.AttrID()] = AttrDef{
		AttrSpec: spec,
		Prototype: attrDef.Prototype,
	}
	return spec, nil
}


func (reg *registry) ExportTo(dst Registry) error {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	defs := make([]AttrDef, 0, len(reg.attrDefs))
	for _, attrDef := range reg.attrDefs {
		defs = append(defs, attrDef)
	}
	
	if err := dst.RegisterAttrDefs(defs); err != nil {
		return err
	}
	return nil
}

func (reg *registry) RegisterAttrDefs(defs []AttrDef) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	for _, def := range defs {
		attrID := def.AttrID()
		if attrID != AttrUID(NilUID) {
			reg.attrDefs[attrID] = def
		}
	}
	return nil
}


func (reg *registry) NewAttrElem(attrID AttrUID) (AttrElemVal, error) {

	// Often, an attrID will be a unnamed scalar ("degenerate") attr, which means we can get the elemDef directly.
	// This is also essential during bootstrapping when the client sends a RegisterDefs is not registered yet.
	def, exists := reg.attrDefs[attrID]
	if !exists {
		return nil, ErrCode_DefNotFound.Errorf("NewAttrElem: attr DefID %v not found", attrID)
	}
	if (def.Prototype == nil) {
		var elemID AttrUID
		elemID[0] = def.ElemTypeUIDx0
		elemID[1] = def.ElemTypeUIDx1
		def, exists = reg.attrDefs[elemID]
	}
	if (def.Prototype == nil) {
		return nil, ErrCode_DefNotFound.Errorf("NewAttrElem: no prototype for attr DefID %v", attrID)
	}
	return def.Prototype.New(), nil
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



func (reg *registry) RegisterDefs(defs *RegisterDefs) error {

	//
	//
	// AttrSpecs
	for _, specIn := range defs.Attrs {
		reg.RegisterAttr(specIn)
		def := AttrDef{
			AttrSpec: *specIn,
		}

		switch {
		case def.Client.AttrName == 0 && def.Native.AttrName != 0:
			return arc.ErrCode_BadSchema.Errorf("RegisterDefs: AttrSpec %v failed to resolve AttrName", def.Client.DefID)
		case def.Client.ElemType == 0 && def.Native.ElemType != 0:
			return arc.ErrCode_BadSchema.Errorf("RegisterDefs: AttrSpec %v failed to resolve ElemType", def.Client.DefID)
		case def.Client.SeriesSpec == 0 && def.Native.SeriesSpec != 0:
			return arc.ErrCode_BadSchema.Errorf("RegisterDefs: AttrSpec %v failed to resolve SeriesSpec", def.Client.DefID)
		case def.Native.DefID == 0:
			return arc.ErrCode_BadSchema.Errorf("RegisterDefs: AttrSpec %v failed to resolve DefID", def.Client.DefID)
		}

		// In the case an attr was already registered natively, we want to still overwrite since the client IDs will now be available.
		reg.attrDefs[def.Native.DefID] = def
	}

	//
	//
	// CellSpecs
	// for _, cellSpec := range defs.Cells {
	// 	def := arc.CellDef{
	// 		ClientDefID: cellSpec.DefID,
	// 		NativeDefID: reg.clientToNativeID[cellSpec.DefID],
	// 	}
	// 	reg.cellDefs[def.NativeDefID] = def
	// }

	return nil
}
