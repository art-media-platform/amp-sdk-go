package arc

func RegisterConstSymbols(reg SessionRegistry) {
	consts := []struct {
		ID   ConstSymbol
		name string
	}{
		{ConstSymbol_Err, "Err"},
		{ConstSymbol_RegisterDefs, "RegisterDefs"},
		{ConstSymbol_HandleURI, "HandleURI"},
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

	prototypes := []ElemVal{
		&AssetRef{},
		&Err{},
		&RegisterDefs{},
		&HandleURI{},
		&LoginChallenge{},
		&LoginResponse{},
		&Login{},
		&CellInfo{},
		//&GeoFix{},
		//&TRS{},
	}
	for _, val := range prototypes {
		reg.RegisterElemType(val)
	}
	return nil
}

func MarshalPbValueToBuf(src PbValue, dst *[]byte) error {
	sz := src.Size()
	if cap(*dst) < sz {
		*dst = make([]byte, sz)
	} else {
		*dst = (*dst)[:sz]
	}
	_, err := src.MarshalToSizedBuffer(*dst)
	return err
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

func (v *AssetRef) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *AssetRef) AttrSpec() string {
	return "AssetRef"
}

func (v *AssetRef) New() ElemVal {
	return &AssetRef{}
}

func (v *Err) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *Err) AttrSpec() string {
	return "Err"
}

func (v *Err) New() ElemVal {
	return &Err{}
}

func (v *HandleURI) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *HandleURI) AttrSpec() string {
	return "HandleURI"
}

func (v *HandleURI) New() ElemVal {
	return &HandleURI{}
}

func (v *Login) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *Login) AttrSpec() string {
	return "Login"
}

func (v *Login) New() ElemVal {
	return &Login{}
}

func (v *LoginChallenge) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *LoginChallenge) AttrSpec() string {
	return "LoginChallenge"
}

func (v *LoginChallenge) New() ElemVal {
	return &LoginChallenge{}
}

func (v *LoginResponse) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *LoginResponse) AttrSpec() string {
	return "LoginResponse"
}

func (v *LoginResponse) New() ElemVal {
	return &LoginResponse{}
}

func (v *CellInfo) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *CellInfo) AttrSpec() string {
	return "CellInfo"
}

func (v *CellInfo) New() ElemVal {
	return &CellInfo{}
}

func (v *RegisterDefs) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *RegisterDefs) AttrSpec() string {
	return "RegisterDefs"
}

func (v *RegisterDefs) New() ElemVal {
	return &RegisterDefs{}
}
