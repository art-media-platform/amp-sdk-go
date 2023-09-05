package arc

import "net/url"

// AttrSpecs used universally
const (
	CellHeaderAttrSpec = "CellHeader"
	CellTextAttrSpec   = "[Locale.Name]CellText"
	CellPosAttrSpec    = "[Surface.Name]Position"
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

func (app *AppBase) RegisterElemType(prototype ElemVal) error {
	err := app.AppContext.Session().RegisterElemType(prototype)
	if err != nil {
		return err
	}
	return nil
}
