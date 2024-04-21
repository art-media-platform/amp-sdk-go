package amp

import (
	fmt "fmt"
	"time"
)

func RegisterBuiltinTypes(reg Registry) error {

	prototypes := []ElemVal{
		&Err{},
		&RegisterDefs{},
		&HandleURI{},
		&Login{},
		&LoginChallenge{},
		&LoginResponse{},

		&PinRequest{},
		&CellHeader{},
		&AssetTag{},
		&AuthToken{},
		&Position{},
		//&TRS{},
	}

	for _, val := range prototypes {
		reg.RegisterPrototype("", val)
	}
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

func (v *AssetTag) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *AssetTag) ElemTypeName() string {
	return "AssetTag"
}

func (v *AssetTag) New() ElemVal {
	return &AssetTag{}
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

func (v *HandleURI) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *HandleURI) ElemTypeName() string {
	return "HandleURI"
}

func (v *HandleURI) New() ElemVal {
	return &HandleURI{}
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

func (v *CellHeader) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *CellHeader) ElemTypeName() string {
	return "CellHeader"
}

func (v *CellHeader) New() ElemVal {
	return &CellHeader{}
}

func (v *CellHeader) SetCreatedAt(t time.Time) {
	tid := ConvertToTimeID(t, false)
	v.CreatedAt_0 = int64(tid[0])
	v.CreatedAt_1 = tid[1]
}

func (v *CellHeader) SetModifiedAt(t time.Time) {
	tid := ConvertToTimeID(t, false)
	v.ModifiedAt_0 = int64(tid[0])
	v.ModifiedAt_1 = tid[1]
}

func (v *RegisterDefs) MarshalToStore(in []byte) (out []byte, err error) {
	return MarshalPbToStore(v, in)
}

func (v *RegisterDefs) ElemTypeName() string {
	return "RegisterDefs"
}

func (v *RegisterDefs) New() ElemVal {
	return &RegisterDefs{}
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

func (v *PinRequest) PinCell() CellID {
	return [3]uint64{
		v.PinCellIDx0,
		v.PinCellIDx1,
		v.PinCellIDx2,
	}
}

func (v *PinRequest) SetCellID(id CellID) {
	v.PinCellIDx0 = id[0]
	v.PinCellIDx1 = id[1]
	v.PinCellIDx2 = id[2]
}

func (v *PinRequest) ContextID() TimeID {
	return TimeIDFromInts(v.ContextID_0, v.ContextID_1)
}

func (v *PinRequest) TargetCell() CellID {
	return [3]uint64{
		v.PinCellIDx0,
		v.PinCellIDx1,
		v.PinCellIDx2,
	}
}

func (v *PinRequest) FormLogLabel() string {
	var strBuf [128]byte
	str := fmt.Appendf(strBuf[:0], "[req_%s] ", v.ContextID().Base32Suffix())
	if v.PinURL != "" {
		str = fmt.Append(str, " ", v.PinURL)
	}
	return string(str)
}
