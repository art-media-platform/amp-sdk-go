package amp

import (
	"strings"
	"sync"
)

func NewRegistry() Registry {
	reg := &registry{
		appsByInvoke: make(map[string]*App),
		appsByUID:    make(map[UID]*App),
		elemDefs:     make(map[AttrID]AttrDef),
		attrDefs:     make(map[AttrID]AttrDef),
	}
	return reg
}

// Implements Registry
type registry struct {
	mu           sync.RWMutex
	appsByInvoke map[string]*App
	appsByUID    map[UID]*App
	elemDefs     map[AttrID]AttrDef
	attrDefs     map[AttrID]AttrDef
}

func (reg *registry) RegisterPrototype(registerAs string, prototype ElemVal) (AttrID, error) {

	// by default, use the canonical name of the prototype
	if registerAs == "" {
		registerAs = prototype.ElemTypeName()
	}
	attrID, err := reg.RegisterAttr(registerAs, prototype)
	if err != nil {
		panic(err)
	}
	return attrID, err
}

func (reg *registry) RegisterAttr(tagSpecExpr string, prototype ElemVal) (AttrID, error) {
	spec, err := FormTagSpec(tagSpecExpr)
	if err != nil {
		return AttrID{}, err
	}
	attrID := spec.AttrID()

	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.attrDefs[attrID] = AttrDef{
		TagSpec:   spec,
		Prototype: prototype,
	}
	return attrID, nil
}

func (reg *registry) Import(other Registry) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	src := other.(*registry)
	src.mu.Lock()
	defer src.mu.Unlock()

	for _, def := range src.elemDefs {
		reg.elemDefs[def.AttrID()] = def
	}
	for _, def := range src.attrDefs {
		reg.attrDefs[def.AttrID()] = def
	}
	for _, app := range src.appsByUID {
		if err := reg.RegisterApp(app); err != nil {
			return err
		}
	}
	return nil
}

// Implements Registry
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

// Implements Registry
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

// Implements Registry
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

func (reg *registry) NewAttrElem(attrID AttrID) (ElemVal, error) {

	// Often, an attrID will be a unnamed scalar attr (which means we can get the elemDef directly.
	// This is also essential during bootstrapping when the client sends a RegisterDefs is not registered yet.
	def, exists := reg.elemDefs[attrID]
	if !exists {
		def, exists = reg.attrDefs[attrID]
		if !exists {
			return nil, ErrCode_DefNotFound.Errorf("NewAttrElem: attr %s not found", attrID.String())
		}
	}
	return def.Prototype.New(), nil
}

func (reg *registry) RegisterDefs(defs *RegisterDefs) error {

	for _, tagSpec := range defs.TagSpecs {
		def := AttrDef{
			TagSpec: *tagSpec,
		}
		reg.attrDefs[def.AttrID()] = def
	}

	return nil
}

/*
func (reg *registry) ResolveTagSpec(attrSpec string) (def TagSpec, err error) {
	expr, err := app_attr_parser.ParseAttrDef(attrSpec)
	if err != nil {
		return TagSpec{}, err
	}

	spec := TagSpec{
		DefID:           reg.resolveNative(attrSpec),
		AttrName:        reg.resolveNative(expr.AttrName),
		ElemType:        reg.resolveNative(expr.ElemType),
		SeriesSpec:      reg.resolveNative(expr.SeriesSpec),
		SeriesIndexType: GetSeriesIndexType(expr.SeriesSpec),
	}

	// If resolving a native-only attr spec, also register it since RegisterDefs() is for client defs only.
	if native {
		prev, exists := reg.attrDefs[spec.DefID]
		if exists {
			if prev.Native != spec {
				err = ErrCode_BadSchema.Errorf("native TagSpec %v already registered with different fields ", spec.DefID)
			}
		} else {
			// If the client also registers the this attr spec later, the client portion will be updated.
			reg.attrDefs[spec.DefID] = AttrDef{
				Native: spec,
			}
		}
	} else {
		clientSpec := TagSpec{
			DefID:           reg.nativeToClientID[spec.DefID],
			AttrName:        reg.nativeToClientID[spec.AttrName],
			ElemType:        reg.nativeToClientID[spec.ElemType],
			SeriesSpec:      reg.nativeToClientID[spec.SeriesSpec],
			SeriesIndexType: spec.SeriesIndexType,
		}

		switch {
		case clientSpec.AttrName == 0 && spec.AttrName != 0:
			err = ErrCode_BadSchema.Errorf("failed to resolve AttrName %q for TagSpec %q", expr.AttrName, attrSpec)
		case clientSpec.ElemType == 0 && spec.ElemType != 0:
			err = ErrCode_BadSchema.Errorf("failed to resolve ElemType %q for TagSpec %q", expr.ElemType, attrSpec)
		case clientSpec.SeriesSpec == 0 && spec.SeriesSpec != 0:
			err = ErrCode_BadSchema.Errorf("failed to resolve SeriesSpec %q for TagSpec %q", expr.SeriesSpec, attrSpec)
		case clientSpec.DefID == 0:
			err = ErrCode_BadSchema.Errorf("failed to resolve %q", attrSpec)
		}

		spec = clientSpec
	}

	return spec, err
}


func (reg *registry) ResolveCellSpec(cellSpec string) (CellDef, error) {
	// TODO: build parser for CellSpec -- for now just assume cellSpec is canonic and good to go

	nativeSpecID := reg.resolveNative(cellSpec)
	//def, exists := reg.cellDefs[nativeSpecID]

	def := CellDef{
		NativeDefID: nativeSpecID,
		ClientDefID: reg.nativeToClientID[nativeSpecID],
	}

	var err error
	if def.ClientDefID == 0 {
		err = ErrCode_BadSchema.Errorf("ResolveCellSpec: failed to resolve %q", cellSpec)
	}

	return def, err
}


func (reg *registry) resolveSymbol(sym *Symbol, autoIssue bool) error {
	reg.tokMu.RLock()
	sym.ID = reg.tokCache[string(sym.Name)]
	reg.tokMu.RUnlock()
	if sym.ID != 0 {
		return nil
	}
	sym.ID = uint32(reg.table.GetSymbolID(sym.Name, autoIssue))
	if sym.ID != 0 {
		reg.tokMu.Lock()
		reg.tokCache[string(sym.Name)] = sym.ID
		reg.tokMu.Unlock()
	}
	return nil
}


func (reg *registry) FormAttr(attrName string, val ElemVal) (AttrElem, error) {
	spec := TagSpec{
		AttrName: attrName,
		ElemType: val.TypeName(),
	}
	if err := reg.ResolveAttr(&spec, false); err != nil {
		return AttrElem{}, err
	}

	return AttrElem{
		Value:  val,
		AttrID: spec.AttrID,
	}, nil
}


func (reg *registry) NewAttrForID(attrID uint32) (AttrElem, error) {
	reg.typesMu.RLock()
	typ, found := reg.types[attrID]
	reg.typesMu.RUnlock()
	if found {
		return AttrElem{}, ErrCode_BadSchema.Errorf("unknown attr ID %v", attrID)
	}

	return AttrElem{
		Value:  typ.elemVal.New(),
		AttrID: attrID,
	}, nil
}

func (reg *registry) ResolveAttr(spec *TagSpec, autoIssue bool) error {
	var typedName string
	hasName := len(spec.AttrName) > 0
	if hasName {
		typedName = spec.AttrName + spec.ElemType
	}

	reg.tokMu.RLock()
	elemID := reg.tokCache[spec.ElemType]
	if hasName {
		spec.AttrID = reg.tokCache[typedName]
	} else {
		spec.AttrID = elemID
	}
	reg.tokMu.RUnlock()

	if elemID != 0 && spec.AttrID != 0 {
		spec.ElemTypeID = elemID
		return nil
	}

	// The above is the hot path and so if it's not found, retroactively check for bad syntax.
	if autoIssue {
		if strings.ContainsAny(spec.AttrName, "./ ") {
			return ErrCode_BadSchema.Errorf("illegal attr name: %q", spec.AttrName)
		}
		if strings.ContainsAny(spec.ElemType, "/ ") || len(spec.ElemType) <= 2 || spec.ElemType[0] != '.' {
			return ErrCode_BadSchema.Errorf("illegal type name: %q", spec.ElemType)
		}
	} else {
		if spec.ElemType == "" {
			return ErrCode_BadSchema.Error("missing TagSpec.ElemType")
		}
	}

	gotElem := false
	gotName := false
	if elemID == 0 {
		elemID = uint32(reg.table.GetSymbolID([]byte(spec.ElemType), autoIssue))
		gotElem = elemID != 0
	}
	if spec.AttrID == 0 {
		if hasName {
			spec.AttrID = uint32(reg.table.GetSymbolID([]byte(typedName), autoIssue))
			gotName = spec.AttrID != 0
		} else {
			spec.AttrID = elemID
		}
	}

	if gotName || gotElem {
		reg.tokMu.Lock()
		if gotElem {
			spec.ElemTypeID = elemID
			reg.tokCache[spec.ElemType] = elemID
		}
		if gotName {
			reg.tokCache[typedName] = spec.AttrID
		}
		reg.tokMu.Unlock()
	}

	if !gotName || !gotElem {
		return ErrCode_BadSchema.Errorf("failed to resolve TagSpec %v", spec)
	}

	return nil
}

func (reg *registry) RegisterAttrType(attrName string, prototype ElemVal) error {
	spec := TagSpec{
		AttrName: attrName,
		ElemType: prototype.TypeName(),
	}
	err := reg.ResolveAttr(&spec, true)
	if err != nil {
		return err
	}

	reg.typesMu.Lock()
	reg.types[spec.AttrID] = attrType{
		//attrName: attrName,
		elemVal: prototype,
	}
	reg.typesMu.Unlock()

	return nil
}

// func (attr *TagSpec) String() string {
//     var buf [128]byte
//     str := fmt.Appendf(buf[:0], "TagSpec{AttrID:%v, TypedName:%q, ValTypeID:%v, SymbolID:%v}", attr.AttrID, attr.TypedName, attr.ValTypeID, attr.SymbolID)
//     return string(str)
// }


func (reg *registry) registerAttr(attr *TagSpec) error {


		if !cleanURI(&attr.AttrName) {
			return ErrCode_BadSchema.Errorf("missing TagSpec.TypedName in attr %q", attr.String())
		}

		if attr.AttrID == 0 {
			return ErrCode_BadSchema.Errorf("missing TagSpec.AttrID in attr %q", attr.TypedName)
		}

		if attr.AttrSymID == 0 {
			attr.AttrSymID = reg.table.GetSymbolID([]byte(attr.TypedName), true).Ord()
		}

		if attr.SeriesType != SeriesType_Fixed && attr.BoundSI != 0 {
			return ErrCode_BadSchema.Errorf("TagSpec.BoundSI is set but is ignored in attr %q", attr.TypedName)
		}

		{
			extPos := strings.IndexByte(attr.TypedName, '.')
			if extPos < 0 {
				return ErrCode_BadSchema.Errorf("missing type suffix in %q", attr.TypedName)
			}
			typeName := attr.TypedName[extPos:]
			typeID := reg.table.GetSymbolID([]byte(typeName), true).Ord()
			if attr.ValTypeID == 0 {
				attr.ValTypeID = typeID
			} else {
				if attr.ValTypeID != typeID {
					return ErrCode_BadSchema.Errorf("TagSpec.ValTypeID (%v) for type %q does not match the registered type (%v)", attr.ValTypeID, typeID, typeID)
				}
			}
		}

		def := reg.attrsBySymbol[attr.AttrSymID]
		if def != nil {
			// TODO: greenlight multiple definitions of the same attr that are indentical
			return ErrCode_BadSchema.Errorf("duplicate AttrID %v", attr.AttrID)
		}
		reg.attrsBySymbol[attr.AttrSymID] = &attrDef{
			spec: *attr,
		}
		reg.attrsByName[attr.TypedName] = attrEntry{
			attrID: attr.AttrID,
			symID:  attr.AttrSymID,
		}

		// if !cleanURI(&attr.ValTypeURI) {
		// 	return ErrCode_BadSchema.Errorf("missing Attrs[%d].ValTypeURI in schema %s for attr %s", i, schema.SchemaDesc(), attr.AttrURI)
		// }

		// if attr.ValTypeID == 0 {
		// 	attr.ValTypeID = uint64(reg.table.GetSymbolID([]byte(attr.ValTypeURI), true))
		// }


	// // Reorder attrs by ascending AttrID for canonic (and efficient) db access
	// // NOTE: This is for a db symbol lookup table for the schema, not for the client-level declaration
	// sort.Slice(schema.Attrs, func(i, j int) bool {
	// 	return schema.Attrs[i].AttrID < schema.Attrs[j].AttrID
	// })
}



func extractTypeName(attr *TagSpec) (string, error) {
	extPos := strings.IndexByte(attr.TypedName, '.')
	if extPos < 0 {
		return "", ErrCode_BadSchema.Errorf("missing type suffix in %q", attr.TypedName)
	}
	typeName := attr.TypedName[:extPos]
}



func (reg *registry) tryResolveDefs(defs []CellDef) error {

    progress := -1
    var unresolved int

    // Remove defs as they able to be registered
    for progress != 0 {
        progress = 0
        unresolved = -1

        for i, def := range defs {
            if def.Spec == nil || def.Spec.Resolved {
                continue
            }

            spec := reg.tryResolve(def.Spec)
            if spec == nil {
                if unresolved < 0 {
                    unresolved = i
                }
                continue
            }

            // TODO -- the proper way to do do this is to:
            //   1) resolve all symbol names into IDs
            //   2) output a canonical text-based spec for def.Spec
            //   3) hash (2) into MD5 etc
            //   4) if (3) already exists, use the already-existing NodeSpec
            //      else, issue a new NodeSpec ID and associate with (3)
            //
            // Until the above is done, we just assume there are no issues and register as we go along.
            def.TypeName = spec.NodeTypeName
            def.Spec = spec
            defs[i] = def
            reg.defs[spec.NodeTypeID] = def
            if reg.nameLookup != nil {
                reg.nameLookup[def.TypeName] = def.Spec.NodeTypeID
            }

            progress++
        }
    }

    if unresolved >= 0 {
        return ErrCode_NodeTypeNotRegistered.ErrWithMsgf("failed to resolve NodeSpec %q", defs[unresolved].TypeName)
    }

    return nil
}




func (schema *AttrSchema) SchemaDesc() string {
	return path.Join(schema.CellDataModel, schema.SchemaName)
}

func (schema *AttrSchema) LookupAttr(typedName string) *TagSpec {
	for _, attr := range schema.Attrs {
		if attr.TypedName == typedName {
			return attr
		}
	}
	return nil
}

func MakeSchemaForType(valTyp reflect.Type) (*AttrSchema, error) {
	numFields := valTyp.NumField()

	schema := &AttrSchema{
		CellDataModel: valTyp.Name(),
		SchemaName:    "on-demand-reflect",
		Attrs:         make([]*TagSpec, 0, numFields),
	}

	for i := 0; i < numFields; i++ {

		// Importantly, AttrID is always set to the field index + 1, so we know what field to inspect when given an AttrID.
		field := valTyp.Field(i)
		if !field.IsExported() {
			continue
		}

		attr := &TagSpec{
			TypedName: field.Name,
			AttrID:  int32(i + 1),
		}

		attrType := field.Type
		attrKind := attrType.Kind()
		switch attrKind {
		case reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64:
			attr.ValTypeID = int32(ValType_int)
		case reflect.String:
			attr.ValTypeID = int32(ValType_string)
		case reflect.Slice:
			elementType := attrType.Elem().Kind()
			switch elementType {
			case reflect.Uint8, reflect.Int8:
				attr.ValTypeID = int32(ValType_bytes)
			}
		}

		if attr.ValTypeID == 0 {
			return nil, ErrCode_ExportErr.Errorf("unsupported type '%s.%s (%v)", schema.CellDataModel, attr.TypedName, attrKind)
		}

		schema.Attrs = append(schema.Attrs, attr)
	}
	return schema, nil
}

// ReadCell loads a cell with the given URI having the inferred schema (built from its fields using reflection).
// The URI is scoped into the user's home planet and AppID.
func ReadCell(ctx AppContext, subKey string, schema *AttrSchema, dstStruct any) error {

	dst := reflect.Indirect(reflect.ValueOf(dstStruct))
	switch dst.Kind() {
	case reflect.Pointer:
		dst = dst.Elem()
	case reflect.Struct:
	default:
		return ErrCode_ExportErr.Errorf("expected struct, got %v", dst.Kind())
	}

	var keyBuf [128]byte
	cellKey := append(append(keyBuf[:0], []byte(ctx.StateScope())...), []byte(subKey)...)

	msgs := make([]*Msg, 0, len(schema.Attrs))
	err := ctx.User().HomePlanet().ReadCell(cellKey, schema, func(msg *Msg) {
		switch msg.Op {
		case MsgOp_PushAttr:
			msgs = append(msgs, msg)
		}
	})
	if err != nil {
		return err
	}

	numFields := dst.NumField()
	valType := dst.Type()

	for fi := 0; fi < numFields; fi++ {
		field := valType.Field(fi)
		for _, ai := range schema.Attrs {
			if ai.TypedName == field.Name {
				for _, msg := range msgs {
					if msg.AttrID == ai.AttrID {
						msg.LoadVal(dst.Field(fi).Addr().Interface())
						goto nextField
					}
				}
			}
		}
	nextField:
	}
	return err
}

// WriteCell is the write analog of ReadCell.
func WriteCell(ctx AppContext, subKey string, schema *AttrSchema, srcStruct any) error {

	src := reflect.Indirect(reflect.ValueOf(srcStruct))
	switch src.Kind() {
	case reflect.Pointer:
		src = src.Elem()
	case reflect.Struct:
	default:
		return ErrCode_ExportErr.Errorf("expected struct, got %v", src.Kind())
	}

	{
		tx := NewMsgBatch()
		msg := tx.AddMsg()
		msg.Op = MsgOp_UpsertCell
		msg.ValType = ValType_SchemaID.Ord()
		msg.ValInt = int64(schema.SchemaID)
		msg.ValBuf = append(append(msg.ValBuf[:0], []byte(ctx.StateScope())...), []byte(subKey)...)

		numFields := src.NumField()
		valType := src.Type()

		for _, attr := range schema.Attrs {
			msg := tx.AddMsg()
			msg.Op = MsgOp_PushAttr
			msg.AttrID = attr.AttrID
			for i := 0; i < numFields; i++ {
				if valType.Field(i).Name == attr.TypedName {
					msg.setVal(src.Field(i).Interface())
					break
				}
			}
			if msg.ValType == ValType_nil.Ord() {
				panic("missing field")
			}
		}

		msg = tx.AddMsg()
		msg.Op = MsgOp_Commit

		if err := ctx.User().HomePlanet().PushTx(tx); err != nil {
			return err
		}
	}

	return nil
}


func (req *CellReq) GetKwArg(argKey string) (string, bool) {
	for _, arg := range req.Args {
		if arg.Key == argKey {
			if arg.Val != "" {
				return arg.Val, true
			}
			return string(arg.ValBuf), true
		}
	}
	return "", false
}

func (req *CellReq) GetChildSchema(modelURI string) *AttrSchema {
	for _, schema := range req.ChildSchemas {
		if schema.CellDataModel == modelURI {
			return schema
		}
	}
	return nil
}

func (req *CellReq) PushBeginPin(target CellID) {
	m := NewTxMsg()
	m.CellID = target.U64()
	m.Op = MsgOp_PinCell
	req.PushTx(m)
}

func (req *CellReq) PushInsertCell(target CellID, schema *AttrSchema) {
	if schema != nil {
		m := NewTxMsg()
		m.CellID = target.U64()
		m.Op = MsgOp_InsertChildCell
		m.ValType = int32(ValType_SchemaID)
		m.ValInt = int64(schema.SchemaID)
		req.PushTx(m)
	}
}

// Pushes the given attr to the client
func (req *CellReq) PushAttr(target CellID, schema *AttrSchema, attrURI string, val Value) {
	attr := schema.LookupAttr(attrURI)
	if attr == nil {
		return
	}

	m := NewTxMsg()
	m.CellID = target.U64()
	m.Op = MsgOp_PushAttr
	m.AttrID = attr.AttrID
	if attr.SeriesType == SeriesType_Fixed {
		m.SI = attr.BoundSI
	}
	val.MarshalToMsg(m)
	if attr.ValTypeID != 0 { // what is this for!?
		m.ValType = int32(attr.ValTypeID)
	}
	req.PushTx(m)
}

func (req *CellReq) PushCheckpoint(err error) {
	m := NewTxMsg()
	m.Op = MsgOp_Commit
	m.CellID = req.PinCell.U64()
	if err != nil {
		m.setVal(err)
	}
	req.PushTx(m)
}

*/

// type Serializable interface {
// 	MarshalToStore(dst *[]byte) error
// 	Unmarshal(src []byte) error
// }

// type Value2[T any] interface {
// 	Serializable
// }

// type ElemValType[V Serializable] interface {
// 	New() V // Returns a new default instance of this ElemVal type
// 	TypeName() string
// 	//MarshalToStore(src V, dst *[]byte) error
// 	//UnmarshalBuf(src []byte, dst V) error
// }
