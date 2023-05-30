package arc

import (
	"sync"

	"github.com/arcspace/go-arc-sdk/stdlib/symbol"
)

type schemaDef struct {
	Schema *AttrSchema
	// TypeName    string
	// TypeInst    NodeType
	// TypeCreator func() NodeType
}

type typeRegistry struct {
	mu    sync.Mutex
	table symbol.Table
	defs  map[int32]schemaDef
	//nameLookup map[string]uint64
}

func NewTypeRegistry(table symbol.Table) TypeRegistry {
	reg := &typeRegistry{
		table: table,
		defs:  make(map[int32]schemaDef),
	}
	// if table == nil {
	// 	reg.nameLookup = make(map[string]uint64)
	// }

	return reg
}

// func (reg *typeRegistry) GetResolvedSpecByName(typeName string) *NodeSpec {
// 	var typeID uint64
// 	if reg.nameLookup != nil {
// 		typeID = reg.nameLookup[typeName]
// 	} else {
// 		typeID = uint64(reg.table.GetSymbolID([]byte(typeName), false))
// 	}
// 	if typeID == 0 {
// 		return nil
// 	}

// 	def := reg.defs[typeID]
// 	return def.Spec
// }

func (reg *typeRegistry) GetSchemaByID(schemaID int32) (*AttrSchema, error) {
	def := reg.defs[schemaID]
	if def.Schema == nil {
		return nil, ErrCode_TypeNotFound.Errorf("Schema %v not found", schemaID)
	}

	return def.Schema, nil
}

func (reg *typeRegistry) ResolveAndRegister(defs *Defs) error {
	var err error

	reg.mu.Lock()
	for _, sym := range defs.Symbols {
		if sym.ID == 0 {
			if len(sym.Value) > 0 {
				sym.ID = uint64(reg.table.GetSymbolID(sym.Value, true))
			}
		} else if len(sym.Value) == 0 {
			sym.Value = reg.table.GetSymbol(symbol.ID(sym.ID), sym.Value[:0])
		}
	}

	for _, schema := range defs.Schemas {
		err = reg.resolveSchema(schema)
		if err != nil {
			break
		}
		if def, exists := reg.defs[schema.SchemaID]; !exists {
			def.Schema = schema
			reg.defs[schema.SchemaID] = def
		}
	}
	reg.mu.Unlock()

	return err
}

func cleanURI(uri *string) bool {
	u := *uri
	N := len(u)
	if N <= 1 {
		return false
	}
	if u[N-1] == '/' {
		*uri = u[:N-1]
	}
	return true
}

func (reg *typeRegistry) resolveSchema(schema *AttrSchema) error {

	if !cleanURI(&schema.CellDataModel) {
		return ErrCode_BadSchema.Error("CellDataModel missing")
	}

	if !cleanURI(&schema.SchemaName) {
		return ErrCode_BadSchema.Error("SchemaName missing")
	}

	if schema.SchemaID == 0 {
		return ErrCode_BadSchema.Error("SchemaID missing")
	}

	for i, attr := range schema.Attrs {

		if !cleanURI(&attr.AttrURI) {
			return ErrCode_BadSchema.Errorf("missing Attrs[%d].AttrURI in schema %s", i, schema.SchemaDesc())
		}

		if attr.AttrID == 0 {
			return ErrCode_BadSchema.Errorf("missing Attrs[%d].AttrID in schema %s for attr %s", i, schema.SchemaDesc(), attr.AttrURI)
		}

		if attr.SeriesType != SeriesType_Fixed && attr.BoundSI != 0 {
			return ErrCode_BadSchema.Errorf("Attrs[%d].Fixed_SI is set but is ignored in schema %s for attr %s", i, schema.SchemaDesc(), attr.AttrURI)
		}

		// if !cleanURI(&attr.ValTypeURI) {
		// 	return ErrCode_BadSchema.Errorf("missing Attrs[%d].ValTypeURI in schema %s for attr %s", i, schema.SchemaDesc(), attr.AttrURI)
		// }

		// if attr.ValTypeID == 0 {
		// 	attr.ValTypeID = uint64(reg.table.GetSymbolID([]byte(attr.ValTypeURI), true))
		// }

	}

	return nil
	// // Reorder attrs by ascending AttrID for canonic (and efficient) db access
	// // NOTE: This is for a db symbol lookup table for the schema, not for the client-level declaration
	// sort.Slice(schema.Attrs, func(i, j int) bool {
	// 	return schema.Attrs[i].AttrID < schema.Attrs[j].AttrID
	// })
}

/*


    func (reg *typeRegistry) GetNodeType(typeID uint64) NodeType {
    def := reg.defs[typeID]
    if def.TypeInst == nil {
        if def.TypeCreator == nil || !def.Spec.Resolved {
            return nil
        }

        nt := def.TypeCreator()
        err := nt.Init(def.Spec, reg.parent)
        if err != nil {
            return nil
        }

        def.TypeInst = nt
        reg.defs[typeID] = def
    }

    return def.TypeInst
}



    if reg.table == nil {
        err := reg.parent.ResolveAndRegister(defs)
        if err != nil {
            return err
        }

        // Propigate new defs from parent
        for _, spec := range defs.Specs {
            reg.nameLookup[spec.NodeTypeName] = spec.NodeTypeID
            typeDef := reg.defs[spec.NodeTypeID]
            typeDef.Spec = spec
            typeDef.TypeName = spec.NodeTypeName
            reg.defs[spec.NodeTypeID] = typeDef
        }

    } else {


    }



func (reg *typeRegistry) RegisterNodeTypes(defs []CellDef) error {
    var err error


    if reg.table == nil {
        total := len(defs)

        tmp := Defs{
            Nodes: make([]*NodeSpec, total),
        }

        for i, typeDef := range defs {
            tmp.Nodes[i] = typeDef.GenericDef
        }

        err = reg.ResolveAndRegister(&tmp)
        if err != nil {
            return err
        }

        for i, typeDef := range defs {
            typeDef.defsDef = tmp.Nodes[i]
            reg.defs[typeDef.defsDef.NodeTypeID] = typeDef
        }

    } else {
        err = reg.registerWithTable(defs)
    }

    return err
}





func (reg *typeRegistry) tryResolveDefs(defs []CellDef) error {

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




    if def.defsDef == nil {
        def.defsDef = reg.tryResolve(def.GenericDef)
    }



    for i, _ := range defs {


    }

    reg.defs[def.defsDef.NodeTypeID] = def



    if def.defsDef == nil {
        if def.GenericDef.Status == DefStatus_Resolved {
            def.defsDef = def.GenericDef
        }
    }

    typeName := def.GenericDef.NodeTypeName

    if def.defsDef == nil {
        if reg.unresolved == nil {
            reg.unresolved = make(map[string]CellDef)
        }

        if _, exists := reg.unresolved[typeName]; exists {
            return ErrCode_NodeTypeNotRegistered.ErrWithMsgf("node type name %q already registered", typeName)
        }

        reg.unresolved[typeName] = def
    } else {
        reg.defs[def.defsDef.NodeTypeID] = def
        if reg.unresolved != nil {
            delete(reg.unresolved, typeName)
        }
    }


    return nil
}*/

/*
    progress := -1
    numResolved := 0

    // Remove defs as they able to be registered
    for progress != 0 {
        progress = 0

        for i := 0; i < total; i++ {
            if defs[i].GenericDef != nil {
                continue
            }

            gdef := in.Nodes[i]
            rdef := reg.tryResolve(gdef)
            if rdef == nil {
                continue
            }

            defs[i] = CellDef{
                GenericDef:  gdef,
                ResolvedDef: rdef,
            }

            // Move the newly resolved def to next dest slot
            progress++
            numResolved++
        }
    }

    if numResolved < total {
        return nil, ErrCode_NodeTypeNotRegistered.ErrWithMsgf("failed to resolve NodeSpec %q", in.Nodes[numResolved].NodeTypeName)
    }

    // TODO check if def already exists (and differs)?  For now, just replace
    for _, def := range defs {
        reg.RegisterTypeDef(def)
        reg.defs[def.defsDef.NodeTypeID] = def
    }

    out := &Defs{
        Nodes: make([]*NodeSpec, total),
    }

    return out, nil

}
*/
