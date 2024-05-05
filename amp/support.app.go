package amp

import (
	"net/url"

	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
)

// TagSpecs used universally
var (
	TagHeaderSpec = tag.FormSpec(AttrSpec, "TagTab")

	//CellLinkID   = MustFormAttrSpec("amp.tag.[Link.Tag]Link")
	//ErrTagSpecID        = FormTagSpecID(((*Err)(nil)).ElemTypeName())
	//LinksAttr     = FormTagSpecID("[Name.Tag]Links")
	//PositionAttr   = FormTagSpecID("[CoordinateScheme.Tag]Position")
)

// This file contains types and interfaces intended to ease an arc app development.
// These types are not required to be used, but are provided as a convenience.

// AppBasic is a helper for implementing AppInstance.
// It is typically extended by embedding it into a struct that builds on top of it.
type AppBasic struct {
	AppContext
}

func (app *AppBasic) OnNew(ctx AppContext) error {
	app.AppContext = ctx
	return nil
}

func (app *AppBasic) HandleURL(*url.URL) error {
	return ErrUnimplemented
}

func (app *AppBasic) OnClosing() {
}
/*
func InsertChannel(dst TxMsg) {
	def := &TagFeed{
		//App: app TODO
	}
	epoch := TagEpoch{
		InsertFeeds: []*TagFeed{
			def,
		},
	}
	
	dst.MarshalUpsert(tag.FromString("insert"), &epoch)

	
}

func (app *AppBasic) CreateFeed() error {

	err = app.AppBasic.OnNew(ctx)
	if err != nil {
		return
	}
	return nil
}

func (app *AppBasic) CreateFeed(target *CreateFeed) error {

}

*/
