package amp

import (
	"net/url"
)

// TagSpecs used universally
var (
	CellHeaderID = MustFormAttrID("CellHeader")
	CellLinkID   = MustFormAttrID("[Link.UID]CellHeader")

	//ErrAttrID        = FormAttrID(((*Err)(nil)).ElemTypeName())
	//LinksAttr     = FormAttrID("[Name.UID]Links")
	//PositionAttr   = FormAttrID("[CoordinateScheme.UID]Position")
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
