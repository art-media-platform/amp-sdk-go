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
		&Err{},
		&RegisterDefs{},
		&HandleURI{},
		&Login{},
		&LoginChallenge{},
		&LoginResponse{},

		&CellInfo{},
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

func (v *AssetRef) TypeName() string {
	return "AssetRef"
}

func (v *AssetRef) New() ElemVal {
	return &AssetRef{}
}

func (v *Err) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *Err) TypeName() string {
	return "Err"
}

func (v *Err) New() ElemVal {
	return &Err{}
}

func (v *HandleURI) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *HandleURI) TypeName() string {
	return "HandleURI"
}

func (v *HandleURI) New() ElemVal {
	return &HandleURI{}
}

func (v *Login) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *Login) TypeName() string {
	return "Login"
}

func (v *Login) New() ElemVal {
	return &Login{}
}

func (v *LoginChallenge) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *LoginChallenge) TypeName() string {
	return "LoginChallenge"
}

func (v *LoginChallenge) New() ElemVal {
	return &LoginChallenge{}
}

func (v *LoginResponse) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *LoginResponse) TypeName() string {
	return "LoginResponse"
}

func (v *LoginResponse) New() ElemVal {
	return &LoginResponse{}
}

func (v *CellInfo) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *CellInfo) TypeName() string {
	return "CellInfo"
}

func (v *CellInfo) New() ElemVal {
	return &CellInfo{}
}

func (v *RegisterDefs) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *RegisterDefs) TypeName() string {
	return "RegisterDefs"
}

func (v *RegisterDefs) New() ElemVal {
	return &RegisterDefs{}
}

func (v *Position) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *Position) TypeName() string {
	return "Position"
}

func (v *Position) New() ElemVal {
	return &Position{}
}

func (v *AuthToken) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *AuthToken) TypeName() string {
	return "AuthToken"
}

func (v *AuthToken) New() ElemVal {
	return &AuthToken{}
}
