package arc

func RegisterConstSymbols(reg SessionRegistry) {
	consts := []struct {
		ID   ConstSymbol
		name string
	}{
		{ConstSymbol_Err, "Err"},
		{ConstSymbol_RegisterDefs, "RegisterDefs"},
		{ConstSymbol_HandleURI, "HandleURI"},
		{ConstSymbol_PinRequest, "PinRequest"},
		{ConstSymbol_Login, "Login"},
		{ConstSymbol_LoginChallenge, "LoginChallenge"},
		{ConstSymbol_LoginResponse, "LoginResponse"},
	}

	defs := RegisterDefs{
		Symbols: make([]*Symbol, len(consts)),
	}

	for i, sym := range consts {
		defs.Symbols[i] = &Symbol{
			ID:   uint32(sym.ID),
			Name: []byte(sym.name),
		}
	}

	if err := reg.RegisterDefs(&defs); err != nil {
		panic(err)
	}
}

func RegisterBuiltInTypes(reg Registry) error {

	prototypes := []AttrElemVal{
		&Err{},
		&RegisterDefs{},
		&HandleURI{},
		&Login{},
		&LoginChallenge{},
		&LoginResponse{},

		&PinRequest{},
		&CellHeader{},
		&AssetRef{},
		&AuthToken{},
		&Position{},
		//&TRS{},
	}
	for _, val := range prototypes {
		reg.RegisterElemType(val)
	}
	return nil
}

func MarshalPbToStore(src PbValue, dst []byte) ([]byte, error) {
	sz := src.Size()
	L := len(dst)
	if cap(dst)-L < sz {
		new := make([]byte, (L+sz+0x400)&^0x3FF)
		copy(new, dst)
	}
	dst = dst[:L+sz]
	_, err := src.MarshalToSizedBuffer(dst[L:])
	return dst, err
}

func ErrorToValue(v error) AttrElemVal {
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

func (v *AssetRef) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *AssetRef) ElemTypeName() string {
	return "AssetRef"
}

func (v *AssetRef) New() AttrElemVal {
	return &AssetRef{}
}

func (v *Err) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *Err) ElemTypeName() string {
	return "Err"
}

func (v *Err) New() AttrElemVal {
	return &Err{}
}

func (v *HandleURI) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *HandleURI) ElemTypeName() string {
	return "HandleURI"
}

func (v *HandleURI) New() AttrElemVal {
	return &HandleURI{}
}

func (v *Login) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *Login) ElemTypeName() string {
	return "Login"
}

func (v *Login) New() AttrElemVal {
	return &Login{}
}

func (v *LoginChallenge) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *LoginChallenge) ElemTypeName() string {
	return "LoginChallenge"
}

func (v *LoginChallenge) New() AttrElemVal {
	return &LoginChallenge{}
}

func (v *LoginResponse) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *LoginResponse) ElemTypeName() string {
	return "LoginResponse"
}

func (v *LoginResponse) New() AttrElemVal {
	return &LoginResponse{}
}

func (v *CellText) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *CellText) TypeName() string {
	return "CellText"
}

func (v *CellText) New() ElemVal {
	return &CellText{}
}

func (v *CellHeader) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *CellHeader) ElemTypeName() string {
	return "CellHeader"
}

func (v *CellHeader) New() AttrElemVal {
	return &CellHeader{}
}

func (v *RegisterDefs) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *RegisterDefs) ElemTypeName() string {
	return "RegisterDefs"
}

func (v *RegisterDefs) New() AttrElemVal {
	return &RegisterDefs{}
}

func (v *Position) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *Position) ElemTypeName() string {
	return "Position"
}

func (v *Position) New() AttrElemVal {
	return &Position{}
}

func (v *AuthToken) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *AuthToken) ElemTypeName() string {
	return "AuthToken"
}

func (v *AuthToken) New() AttrElemVal {
	return &AuthToken{}
}

func (v *PinRequest) MarshalToStore(dst []byte) ([]byte, error) {
	return MarshalPbToStore(v, dst)
}

func (v *PinRequest) ElemTypeName() string {
	return "PinRequest"
}

func (v *PinRequest) New() AttrElemVal {
	return &PinRequest{}
}

func (v *PinRequest) CellID() CellID {
	return CellIDFromU64(v.PinCellIDx0, v.PinCellIDx1)
}
