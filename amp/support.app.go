package amp

import (
	"net/url"
)

// TagSpecs used universally
var (
	CellHeaderAttrID = MustFormAttrSpec("amp.tag.CellHeader") // CellLink??
	
	//CellLinkID   = MustFormAttrSpec("amp.tag.[Link.TagID]CellHeader")
	//ErrTagSpecID        = FormTagSpecID(((*Err)(nil)).ElemTypeName())
	//LinksAttr     = FormTagSpecID("[Name.TagID]Links")
	//PositionAttr   = FormTagSpecID("[CoordinateScheme.TagID]Position")
)

// This file contains types and interfaces intended to ease an arc app development.
// These types are not required to be used, but are provided as a convenience.

// AppBase is a helper for implementing AppInstance.
// It is typically extended by embedding it into a struct that builds on top of it.
type AppBase struct {
	AppContext
}

func (app *AppBase) OnNew(ctx AppContext) error {
	app.AppContext = ctx
	return nil
}

func (app *AppBase) HandleURL(*url.URL) error {
	return ErrUnimplemented
}

func (app *AppBase) OnClosing() {
}

func (app *AppBase) RegisterPrototype(registerAs string, prototype ElemVal) error {
	_, err := app.AppContext.Session().RegisterPrototype(registerAs, prototype)
	if err != nil {
		return err
	}
	return nil
}
