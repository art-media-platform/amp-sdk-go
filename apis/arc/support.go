package arc

import (
	"bytes"
	"path"
	"reflect"
	"time"

	"github.com/arcspace/go-arc-sdk/stdlib/bufs"
)

// TID identifies a specific planet, node, or transaction.
//
// Unless otherwise specified a TID in the wild should always be considered read-only.
type TID []byte

// TIDBuf is the blob version of a TID
type TIDBuf [TIDBinaryLen]byte

// Byte size of a TID, a hash with a leading embedded big endian binary time index.
const TIDBinaryLen = int(Const_TIDBinaryLen)

// ASCII-compatible string length of a (binary) TID encoded into its base32 form.
const TIDStringLen = int(Const_TIDStringLen)

// nilTID is a zeroed TID that denotes a void/nil/zero value of a TID
var nilTID = TIDBuf{}

// TimeFS is a signed int64 that stores a UTC in 1/2^16 sec ticks elapsed since Jan 1, 1970 UTC ("FS" = fractional seconds)
//
// timeFS := TimeNowFS()
//
// Shifting this right 16 bits will yield standard Unix time.
// This means there are 47 bits dedicated for seconds, implying max timestamp of 4.4 million years.
type TimeFS int64

const (
	SI_DistantFuture = TimeFS(0x7FFFFFFFFFFFFFFF)
)

// TimeNowFS returns the current time (a standard unix UTC timestamp in 1/1<<16 seconds)
func TimeNowFS() TimeFS {
	return ConvertToTimeFS(time.Now())
}

// Converts a time.Time to a TimeFS.
func ConvertToTimeFS(t time.Time) TimeFS {
	timeFS := t.Unix() << 16
	frac := uint16((2199 * (uint32(t.Nanosecond()) >> 10)) >> 15)
	return TimeFS(timeFS | int64(frac))
}

// TID is a convenience function that returns the TID contained within this TIDBuf.
func (tid *TIDBuf) TID() TID {
	return tid[:]
}

// Base32 returns this TID in Base32 form.
func (tid *TIDBuf) Base32() string {
	return bufs.Base32Encoding.EncodeToString(tid[:])
}

// IsNil returns true if this TID length is 0 or is equal to NilTID
func (tid TID) IsNil() bool {
	if len(tid) == 0 {
		return true
	}

	if bytes.Equal(tid, nilTID[:]) {
		return true
	}

	return false
}

// Clone returns a duplicate of this TID
func (tid TID) Clone() TID {
	dupe := make([]byte, len(tid))
	copy(dupe, tid)
	return dupe
}

// Buf is a convenience function that make a new TIDBuf from a TID byte slice.
func (tid TID) Buf() TIDBuf {
	var blob TIDBuf
	copy(blob[:], tid)
	return blob
}

// Base32 returns this TID in Base32 form.
func (tid TID) Base32() string {
	return bufs.Base32Encoding.EncodeToString(tid)
}

// Appends the base 32 ASCII encoding of this TID to the given buffer
func (tid TID) AppendAsBase32(in []byte) []byte {
	encLen := bufs.Base32Encoding.EncodedLen(len(tid))
	needed := len(in) + encLen
	buf := in
	if needed > cap(buf) {
		buf = make([]byte, (needed+0x100)&^0xFF)
		buf = append(buf[:0], in...)
	}
	buf = buf[:needed]
	bufs.Base32Encoding.Encode(buf[len(in):needed], tid)
	return buf
}

// SuffixStr returns the last few digits of this TID in string form (for easy reading, logs, etc)
func (tid TID) SuffixStr() string {
	const summaryStrLen = 5

	R := len(tid)
	L := R - summaryStrLen
	if L < 0 {
		L = 0
	}
	return bufs.Base32Encoding.EncodeToString(tid[L:R])
}

// SetTimeAndHash writes the given timestamp and the right-most part of inSig into this TID.
//
// See comments for TIDBinaryLen
func (tid TID) SetTimeAndHash(time TimeFS, hash []byte) {
	tid.SetTimeFS(time)
	tid.SetHash(hash)
}

// SetHash sets the sig/hash portion of this ID
func (tid TID) SetHash(hash []byte) {
	const TIDHashSz = int(Const_TIDBinaryLen - Const_TIDTimestampSz)
	pos := len(hash) - TIDHashSz
	if pos >= 0 {
		copy(tid[TIDHashSz:], hash[pos:])
	} else {
		for i := 8; i < int(Const_TIDBinaryLen); i++ {
			tid[i] = hash[i]
		}
	}
}

// SetTimeFS writes the given timestamp into this TIS
func (tid TID) SetTimeFS(t TimeFS) {
	tid[0] = byte(t >> 56)
	tid[1] = byte(t >> 48)
	tid[2] = byte(t >> 40)
	tid[3] = byte(t >> 32)
	tid[4] = byte(t >> 24)
	tid[5] = byte(t >> 16)
	tid[6] = byte(t >> 8)
	tid[7] = byte(t)
}

// ExtractTimeFS returns the unix timestamp embedded in this TID (a unix timestamp in 1<<16 seconds UTC)
func (tid TID) ExtractTimeFS() TimeFS {
	t := int64(tid[0])
	t = (t << 8) | int64(tid[1])
	t = (t << 8) | int64(tid[2])
	t = (t << 8) | int64(tid[3])
	t = (t << 8) | int64(tid[4])
	t = (t << 8) | int64(tid[5])
	t = (t << 8) | int64(tid[6])
	t = (t << 8) | int64(tid[7])

	return TimeFS(t)
}

// ExtractTime returns the unix timestamp embedded in this TID (a unix timestamp in seconds UTC)
func (tid TID) ExtractTime() int64 {
	t := int64(tid[0])
	t = (t << 8) | int64(tid[1])
	t = (t << 8) | int64(tid[2])
	t = (t << 8) | int64(tid[3])
	t = (t << 8) | int64(tid[4])
	t = (t << 8) | int64(tid[5])

	return t
}

// SelectEarlier looks in inTime a chooses whichever is earlier.
//
// If t is later than the time embedded in this TID, then this function has no effect and returns false.
//
// If t is earlier, then this TID is initialized to t (and the rest zeroed out) and returns true.
func (tid TID) SelectEarlier(t TimeFS) bool {

	TIDt := tid.ExtractTimeFS()

	// Timestamp of 0 is reserved and should only reflect an invalid/uninitialized TID.
	if t < 0 {
		t = 0
	}

	if t < TIDt || t == 0 {
		tid.SetTimeFS(t)
		for i := 8; i < len(tid); i++ {
			tid[i] = 0
		}
		return true
	}

	return false
}

// CopyNext copies the given TID and increments it by 1, typically useful for seeking the next entry after a given one.
func (tid TID) CopyNext(inTID TID) {
	copy(tid, inTID)
	for j := len(tid) - 1; j > 0; j-- {
		tid[j]++
		if tid[j] > 0 {
			break
		}
	}
}

func (schema *AttrSchema) SchemaDesc() string {
	return path.Join(schema.ScopeID, schema.CellDataModel, schema.SchemaName)
}

func (schema *AttrSchema) LookupAttr(attrURI string) *AttrSpec {
	for _, attr := range schema.Attrs {
		if attr.AttrURI == attrURI {
			return attr
		}
	}
	return nil
}

// func (schema *AttrSchema) SelectAppForSchema() (*AppModule, error) {
// 	return gAppRegistry.SelectAppForSchema(schema)
// }

func MakeSchemaForType(scopeID string, valTyp reflect.Type) (*AttrSchema, error) {
	numFields := valTyp.NumField()

	schema := &AttrSchema{
		ScopeID:       scopeID,
		CellDataModel: valTyp.Name(),
		SchemaName:    "on-demand-reflect",
		Attrs:         make([]*AttrSpec, 0, numFields),
	}

	for i := 0; i < numFields; i++ {

		// Importantly, AttrID is always set to the field index + 1, so we know what field to inspect when given an AttrID.
		field := valTyp.Field(i)
		if !field.IsExported() {
			continue
		}

		attr := &AttrSpec{
			AttrURI: field.Name,
			AttrID:  int32(i + 1),
		}

		attrType := field.Type
		attrKind := attrType.Kind()
		switch attrKind {
		case reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64:
			attr.ValTypeID = ValType_int
		case reflect.String:
			attr.ValTypeID = ValType_string
		case reflect.Slice:
			elementType := attrType.Elem().Kind()
			switch elementType {
			case reflect.Uint8, reflect.Int8:
				attr.ValTypeID = ValType_bytes
			}
		}

		if attr.ValTypeID == 0 {
			return nil, ErrCode_ExportErr.Errorf("unsupported type '%s.%s (%v)", schema.CellDataModel, attr.AttrURI, attrKind)
		}

		schema.Attrs = append(schema.Attrs, attr)
	}
	return schema, nil
}

// ReadCell loads a cell with the given URI having the inferred schema (built from its fields using reflection).
// The URI is scoped into the user's home planet and AppID.
func ReadCell(ctx AppContext, subKey string, schema *AttrSchema, dstStruct any) error {

	dst := reflect.Indirect(reflect.ValueOf(dstStruct))
	switch dst.Kind() {
	case reflect.Pointer:
		dst = dst.Elem()
	case reflect.Struct:
	default:
		return ErrCode_ExportErr.Errorf("expected struct, got %v", dst.Kind())
	}

	var keyBuf [128]byte
	cellKey := append(append(keyBuf[:0], []byte(ctx.StateScope())...), []byte(subKey)...)

	msgs := make([]*Msg, 0, len(schema.Attrs))
	err := ctx.User().HomePlanet().ReadCell(cellKey, schema, func(msg *Msg) {
		switch msg.Op {
		case MsgOp_PushAttr:
			msgs = append(msgs, msg)
		}
	})
	if err != nil {
		return err
	}

	numFields := dst.NumField()
	valType := dst.Type()

	for fi := 0; fi < numFields; fi++ {
		field := valType.Field(fi)
		for _, ai := range schema.Attrs {
			if ai.AttrURI == field.Name {
				for _, msg := range msgs {
					if msg.AttrID == ai.AttrID {
						msg.LoadVal(dst.Field(fi).Addr().Interface())
						goto nextField
					}
				}
			}
		}
	nextField:
	}
	return err
}

// WriteCell is the write analog of ReadCell.
func WriteCell(ctx AppContext, subKey string, schema *AttrSchema, srcStruct any) error {

	src := reflect.Indirect(reflect.ValueOf(srcStruct))
	switch src.Kind() {
	case reflect.Pointer:
		src = src.Elem()
	case reflect.Struct:
	default:
		return ErrCode_ExportErr.Errorf("expected struct, got %v", src.Kind())
	}

	{
		tx := NewMsgBatch()
		msg := tx.AddMsg()
		msg.Op = MsgOp_UpsertCell
		msg.ValType = ValType_SchemaID
		msg.ValInt = int64(schema.SchemaID)
		msg.ValBuf = append(append(msg.ValBuf[:0], []byte(ctx.StateScope())...), []byte(subKey)...)

		numFields := src.NumField()
		valType := src.Type()

		for _, attr := range schema.Attrs {
			msg := tx.AddMsg()
			msg.Op = MsgOp_PushAttr
			msg.AttrID = attr.AttrID
			for i := 0; i < numFields; i++ {
				if valType.Field(i).Name == attr.AttrURI {
					msg.SetVal(src.Field(i).Interface())
					break
				}
			}
			if msg.ValType == ValType_nil {
				panic("missing field")
			}
		}

		msg = tx.AddMsg()
		msg.Op = MsgOp_Commit

		if err := ctx.User().HomePlanet().PushTx(tx); err != nil {
			return err
		}
	}

	return nil
}

func (req *CellReq) GetKwArg(argKey string) (string, bool) {
	for _, arg := range req.Args {
		if arg.Key == argKey {
			if arg.Val != "" {
				return arg.Val, true
			}
			return string(arg.ValBuf), true
		}
	}
	return "", false
}

func (req *CellReq) GetChildSchema(modelURI string) *AttrSchema {
	for _, schema := range req.ChildSchemas {
		if schema.CellDataModel == modelURI {
			return schema
		}
	}
	return nil
}

func (req *CellReq) PushBeginPin(target CellID) {
	m := NewMsg()
	m.CellID = target.U64()
	m.Op = MsgOp_PinCell
	req.PushMsg(m)
}

func (req *CellReq) PushInsertCell(target CellID, schema *AttrSchema) {
	if schema != nil {
		m := NewMsg()
		m.CellID = target.U64()
		m.Op = MsgOp_InsertChildCell
		m.ValType = ValType_SchemaID
		m.ValInt = int64(schema.SchemaID)
		req.PushMsg(m)
	}
}

// Pushes the given attr to the client
func (req *CellReq) PushAttr(target CellID, schema *AttrSchema, attrURI string, attrVal interface{}) {
	attr := schema.LookupAttr(attrURI)
	if attr == nil {
		return
	}

	m := NewMsg()
	m.CellID = target.U64()
	m.Op = MsgOp_PushAttr
	m.AttrID = attr.AttrID
	if attr.SeriesType == SeriesType_Fixed {
		m.SI = attr.BoundSI
	}
	m.SetVal(attrVal)
	if attr.ValTypeID != 0 {
		m.ValType = attr.ValTypeID
	}
	req.PushMsg(m)
}

func (req *CellReq) PushCheckpoint(err error) {
	m := NewMsg()
	m.Op = MsgOp_Commit
	m.CellID = req.PinCell.U64()
	if err != nil {
		m.SetVal(err)
	}
	req.PushMsg(m)
}
