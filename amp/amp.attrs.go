package amp

import (
	"time"

	"github.com/art-media-platform/amp-sdk-go/stdlib/tag"
)

var (
	// CellID hard-wired to denote the root c
	MetaNodeID = tag.ID{0, 0, 2701}

	TagRoot  = tag.Spec{}.With("amp")
	AttrSpec = TagRoot.With("attr")
	AppSpec  = TagRoot.With("app")

	CellProperties = AttrSpec.With("cell-properties")
)

func RegisterBuiltinTypes(reg Registry) error {

	prototypes := []tag.Value{
		&Err{},
		&Tag{},
		&LaunchURL{},
		&Login{},
		&LoginChallenge{},
		&LoginResponse{},
		&LoginCheckpoint{},
		&PinRequest{},
	}

	for _, pi := range prototypes {
		reg.RegisterPrototype(AttrSpec, pi, "")
	}

	return nil
}

func MarshalPbToStore(src tag.ValuePb, dst []byte) ([]byte, error) {
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

func ErrorToValue(v error) tag.Value {
	if v == nil {
		return nil
	}
	artErr, _ := v.(*Err)
	if artErr == nil {
		wrapped := ErrCode_UnnamedErr.Wrap(v)
		artErr = wrapped.(*Err)
	}
	return artErr
}

func (v *TagID) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *TagID) TagSpec() tag.Spec {
	return AttrSpec.With("TagID")
}

func (v *TagID) New() tag.Value {
	return &TagID{}
}

func (v *TagID) SetFromTime(t time.Time) {
	tag := tag.FromTime(t, false)
	v.ID_0 = int64(tag[0])
	v.ID_1 = tag[1]
	v.ID_2 = tag[2]
}

func (v *Tag) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *Tag) TagSpec() tag.Spec {
	return AttrSpec.With("Tag")
}

func (v *Tag) New() tag.Value {
	return &Tag{}
}

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

// func (v *Tag) SetTimesatamp(t time.Time) {
// 	tag := tag.FromTime(t, false)
// 	v.Timestamp = int64(tag[0])
// }

func (v *Err) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *Err) TagSpec() tag.Spec {
	return AttrSpec.With("Err")
}

func (v *Err) New() tag.Value {
	return &Err{}
}

func (v *LaunchURL) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *LaunchURL) TagSpec() tag.Spec {
	return AttrSpec.With("LaunchURL")
}

func (v *LaunchURL) New() tag.Value {
	return &LaunchURL{}
}

func (v *Login) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *Login) TagSpec() tag.Spec {
	return AttrSpec.With("Login")
}

func (v *Login) New() tag.Value {
	return &Login{}
}

func (v *LoginChallenge) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *LoginChallenge) TagSpec() tag.Spec {
	return AttrSpec.With("LoginChallenge")
}

func (v *LoginChallenge) New() tag.Value {
	return &LoginChallenge{}
}

func (v *LoginResponse) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *LoginResponse) TagSpec() tag.Spec {
	return AttrSpec.With("LoginResponse")
}

func (v *LoginResponse) New() tag.Value {
	return &LoginResponse{}
}

func (v *LoginCheckpoint) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *LoginCheckpoint) TagSpec() tag.Spec {
	return AttrSpec.With("LoginCheckpoint")
}

func (v *LoginCheckpoint) New() tag.Value {
	return &LoginCheckpoint{}
}

func (v *PinRequest) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *PinRequest) TagSpec() tag.Spec {
	return AttrSpec.With("PinRequest")
}

func (v *PinRequest) New() tag.Value {
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
