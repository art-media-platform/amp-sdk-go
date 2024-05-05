package amp

import (
	"time"

	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
)

// URI form of a glyph typically followed by a media (mime) type.
const GenericGlyphURL = "amp:glyph/"

// Describes an asset to be an image stream but not specify format / codec
const (
	GenericImageType = "image/*"
	GenericAudioType = "audio/*"
	GenericVideoType = "video/*"
)

func RegisterBuiltinTypes(reg Registry) error {

	prototypes := []ElemVal{
		&Err{},
		&FeedGenesis{},
		&TagAttr{},
		&HandleURL{},
		&Login{},
		&LoginChallenge{},
		&LoginResponse{},
		&AuthCheckpoint{},
		&PinRequest{},
	}

	for _, val := range prototypes {
		reg.RegisterPrototype(AttrSpec, val)
	}

	reg.RegisterPrototype(tag.FormSpec(AttrSpec, "genesis"), &TagAttr{})
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

func (v *TagAttr) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *TagAttr) ElemTypeName() string {
	return "TagAttr"
}

func (v *TagAttr) New() ElemVal {
	return &TagAttr{}
}

func (v *TagAttr) SetAttrSpec(attrSpec tag.ID) {
	v.AttrSpec_0 = int64(attrSpec[0])
	v.AttrSpec_1 = attrSpec[1]
	v.AttrSpec_2 = attrSpec[2]
}

func (v *TagAttr) AttrSpec() tag.ID {
	return [3]uint64{
		uint64(v.AttrSpec_0),
		v.AttrSpec_1,
		v.AttrSpec_2,
	}
}

func (v *TagAttr) TagID() tag.ID {
	return [3]uint64{
		uint64(v.Tag_0),
		v.Tag_1,
		v.Tag_2,
	}
}

func (v *TagAttr) SetTag(tagID tag.ID) {
	v.Tag_0 = int64(tagID[0])
	v.Tag_1 = tagID[1]
	v.Tag_2 = tagID[2]
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

func (v *FeedGenesis) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *FeedGenesis) ElemTypeName() string {
	return "FeedGenesis"
}

func (v *FeedGenesis) New() ElemVal {
	return &FeedGenesis{}
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

func (v *HandleURL) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *HandleURL) ElemTypeName() string {
	return "HandleURL"
}

func (v *HandleURL) New() ElemVal {
	return &HandleURL{}
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

/*
func (v *PinRequest) SetTargetID(id tag.ID) {
	if v.PinTarget == nil {
		v.PinTarget = &TagAttr{}
	}
	v.PinTarget.Tag_0 = int64(id[0])
	v.PinTarget.Tag_1 = id[1]
	v.PinTarget.Tag_2 = id[2]
}

func (v *PinRequest) TargetID() tag.ID {
	return [3]uint64{
		uint64(v.PinTarget.Tag_0),
		v.PinTarget.Tag_1,
		v.PinTarget.Tag_2,
	}
}
*/
