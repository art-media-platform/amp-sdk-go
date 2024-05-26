package amp

import (
	"time"

	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
)

const (
	// URL prefix for a glyph and is typically followed by a media (mime) type.
	GenericGlyphURL = "amp:glyph/"

	GenericImageType = "image/*"
	GenericAudioType = "audio/*"
	GenericVideoType = "video/*"
)

// Common universal glyphs
var (
	GenericFolderGlyph = &Tag{
		Use: TagUse_Glyph,
		URL: GenericGlyphURL + "application/x-directory",
	}
)

func FormPinnableTag(attrSpec tag.Spec) *Tag {
	//url := "amp:attr/" + attrSpec.Canonic
	tag := &Tag{
		Use:     TagUse_Pinnable,
		TagID_0: int64(attrSpec.ID[0]),
		TagID_1: attrSpec.ID[1],
		TagID_2: attrSpec.ID[2],
	}
	return tag
}

type PinnableAttr struct {
	Spec tag.Spec
}

// A tag.Spec has two available orderings, with the more stricter being the default interpretation:
// .   - Most Common
//   - Exact UTF8 literal match:     "only.One.path" != ""only.one.path"
//
// //   - Exact UTF8 swizzle
// //   - Exact UTF8 sum
// xor match:         "only.One.path" != ""only.one.path"
// //         amp.app.attr.catalog.tagTab" == "({CanonicalSepChar}[i])+"*
//   - Order-independent match: "attr1.attr2.attr3" == "attr1.attr3.attr2" == "({CanonicalSepChar}[i])+"*
var (
	AppSpec              = tag.FormSpec(tag.Spec{}, "amp.app")
	AttrSpec             = tag.FormSpec(tag.Spec{}, "amp.attr")
	HeaderChSpec         = tag.FormSpec(tag.Spec{}, "amp.app.ch.links.TagTab")
	TabCatalogSpec       = tag.FormSpec(tag.Spec{}, "amp.app.attr.catalog.TagTab")
	ContentSpec          = tag.FormSpec(tag.Spec{}, "amp.app.attr.content.Tag")
	PinnableTabIndexSpec = tag.FormSpec(tag.Spec{}, "amp.app.attr.index.TagTab")
	MetaAttrSpec         = tag.FormSpec(AttrSpec, "meta")

	PinnedTabSpec = tag.FormSpec(AttrSpec, "pinned.TagTab")
	ChildTabSpec  = tag.FormSpec(AttrSpec, "TagTab")

	//PinnableContent    = FormPinnableTag(ContentSpec)
	PinnableCatalog = FormPinnableTag(TabCatalogSpec)
)

func RegisterBuiltinTypes(reg Registry) error {

	prototypes := []ElemVal{
		&Err{},
		&Tag{},
		&LaunchURL{},
		&Login{},
		&LoginChallenge{},
		&LoginResponse{},
		&AuthCheckpoint{},
		&PinRequest{},
	}

	for _, pi := range prototypes {
		reg.RegisterPrototype(AttrSpec, pi, "")
	}

	reg.RegisterPrototype(tag.FormSpec(AttrSpec, "genesis"), &Tag{}, "")
	return nil
}

func MarshalPbToStore(src PbValue, dst []byte) ([]byte, error) {
	oldLen := len(dst)
	newLen := oldLen + src.Size()
	if cap(dst) < newLen {
		old := dst
		dst = make([]byte, (newLen+0x400)&^0x3FF)
		copy(dst, old)
	}
	dst = dst[:newLen]
	_, err := src.MarshalToSizedBuffer(dst[oldLen:])
	return dst, err
}

func ErrorToValue(v error) ElemVal {
	if v == nil {
		return nil
	}
	arcErr, _ := v.(*Err)
	if arcErr == nil {
		wrapped := ErrCode_UnnamedErr.Wrap(v)
		arcErr = wrapped.(*Err)
	}
	return arcErr
}

func (v *Tag) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *Tag) ElemTypeName() string {
	return "Tag"
}

func (v *Tag) New() ElemVal {
	return &Tag{}
}

// func (v *Tag) SetAttrID(attrSpec tag.ID) {
// 	v.AttrID_0 = int64(attrSpec[0])
// 	v.AttrID_1 = attrSpec[1]
// 	v.AttrID_2 = attrSpec[2]
// }

// func (v *Tag) AttrID() tag.ID {
// 	return [3]uint64{
// 		uint64(v.AttrID_0),
// 		v.AttrID_1,
// 		v.AttrID_2,
// 	}
// }

func (v *Tag) TagID() tag.ID {
	return [3]uint64{
		uint64(v.TagID_0),
		v.TagID_1,
		v.TagID_2,
	}
}

func (v *Tag) SetTagID(tagID tag.ID) {
	v.TagID_0 = int64(tagID[0])
	v.TagID_1 = tagID[1]
	v.TagID_2 = tagID[2]
}

func (v *TagTab) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *TagTab) ElemTypeName() string {
	return "TagTab"
}

func (v *TagTab) New() ElemVal {
	return &TagTab{}
}

func (v *TagTab) SetCreatedAt(t time.Time) {
	tag := tag.FromTime(t, false)
	v.CreatedAt = int64(tag[0])
}

func (v *TagTab) SetModifiedAt(t time.Time) {
	tag := tag.FromTime(t, false)
	v.ModifiedAt = int64(tag[0])
}

func (v *TagTab) AddPinnableContentTag(contentType string) {
	tag := &Tag{
		Use:         TagUse_Pinnable,
		ContentType: contentType,
	}
	tag.SetTagID(ContentSpec.ID)
	v.Tags = append(v.Tags, tag)
}

func (v *TagTab) AddTag(tag *Tag) {
	v.Tags = append(v.Tags, tag)
}

func (v *Err) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *Err) ElemTypeName() string {
	return "Err"
}

func (v *Err) New() ElemVal {
	return &Err{}
}

func (v *LaunchURL) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *LaunchURL) ElemTypeName() string {
	return "LaunchURL"
}

func (v *LaunchURL) New() ElemVal {
	return &LaunchURL{}
}

func (v *Login) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *Login) ElemTypeName() string {
	return "Login"
}

func (v *Login) New() ElemVal {
	return &Login{}
}

func (v *LoginChallenge) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *LoginChallenge) ElemTypeName() string {
	return "LoginChallenge"
}

func (v *LoginChallenge) New() ElemVal {
	return &LoginChallenge{}
}

func (v *LoginResponse) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *LoginResponse) ElemTypeName() string {
	return "LoginResponse"
}

func (v *LoginResponse) New() ElemVal {
	return &LoginResponse{}
}

func (v *AuthCheckpoint) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *AuthCheckpoint) ElemTypeName() string {
	return "AuthCheckpoint"
}

func (v *AuthCheckpoint) New() ElemVal {
	return &AuthCheckpoint{}
}

func (v *Position) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *Position) ElemTypeName() string {
	return "Position"
}

func (v *Position) New() ElemVal {
	return &Position{}
}

func (v *AuthToken) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *AuthToken) ElemTypeName() string {
	return "AuthToken"
}

func (v *AuthToken) New() ElemVal {
	return &AuthToken{}
}

func (v *PinRequest) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *PinRequest) ElemTypeName() string {
	return "PinRequest"
}

func (v *PinRequest) New() ElemVal {
	return &PinRequest{}
}

func (v *PinRequest) SetTargetID(id tag.ID) {
	if v.PinTarget == nil {
		v.PinTarget = &Tag{}
	}
	v.PinTarget.TagID_0 = int64(id[0])
	v.PinTarget.TagID_1 = id[1]
	v.PinTarget.TagID_2 = id[2]
}

func (v *PinRequest) TargetID() tag.ID {
	target := v.PinTarget
	if target == nil {
		return tag.ID{}
	}
	return [3]uint64{
		uint64(target.TagID_0),
		target.TagID_1,
		target.TagID_2,
	}
}

/*
func (v *Request) AttrsToPin() map[tag.ID]struct{} {
	pinAttrs := make(map[tag.ID]struct{}, len(v.PinAttrs))
	for _, attr := range v.PinAttrs {
		attrID := attr.AttrID()
		if attrID.IsNil() && attr.URL != "" {
			attrID = tag.FormSpec(tag.Spec{}, attr.URL).ID
		}
		if !attrID.IsNil() {
			pinAttrs[attrID] = struct{}{}
		}
	}
	return pinAttrs
}
*/
