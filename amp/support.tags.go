package amp

import (
	"crypto/md5"
	"encoding/binary"
	"time"

	"github.com/amp-3d/amp-sdk-go/stdlib/bufs"
)

// URI form of a glyph typically followed by a media (mime) type.
const GenericGlyphURL = "amp:glyph/"

// Describes an asset to be an image stream but not specify format / codec
var GenericImageType = "image/*"

// TagID is a signed 16 byte UTC time index in big endian form, with 80 bits (10 bytes) of fractional precision.
// This means there are 47 bits dedicated for whole seconds => +/- 4.4 million years
//
// Note: (TagID[0] >> 16) yields a standard 64-bit Unix UTC time.
type TagID [3]uint64

const (
	TagID_x0_SecondsShift = 16                  // shift to get the 64-bit Unix time from TagID[0]
	TagID_NanosecStep     = uint64(0x44B82FA1C) // 1<<64 div 1/1e9 -- reflects Go's single nanosecond resolution spread over a 64 bits
	TagID_EntropyMask     = uint64(0x3FFFFFFFF) // entropy bit mask for TagID[1] -- slightly smaller than 1 ns resolution
)

func TagIDFromInts(x0 int64, x1 uint64) TagID {
	return TagID{uint64(x0), x1}
}

// Returns the current time as a TagID, statistically guaranteed to be unique even when called in rapid succession.
func NewTagID() TagID {
	return TimeToTagID(time.Now(), true)
}

func TimeToTagID(t time.Time, addEntropy bool) TagID {
	ns_b10 := uint64(t.Nanosecond())
	ns_f64 := ns_b10 * TagID_NanosecStep // map 0..999999999 to 0..(2^64-1)

	t_00_06 := uint64(t.Unix()) << TagID_x0_SecondsShift
	t_06_08 := ns_f64 >> (64 - TagID_x0_SecondsShift)
	t_08_15 := ns_f64 << TagID_x0_SecondsShift
	if addEntropy {
		gTagIDSeed = ns_f64 ^ (gTagIDSeed * 5237732522753)
		t_08_15 ^= gTagIDSeed & TagID_EntropyMask
	}

	tid := TagID{
		0,
		t_00_06 | uint64(t_06_08),
		t_08_15,
	}
	return tid
}

func StringToTagID(canonicExpr string) TagID {
	if canonicExpr == "" {
		return TagID{}
	}
	hasher := md5.New()
	hasher.Write([]byte(canonicExpr))

	var hashBuf [16]byte
	hash := hasher.Sum(hashBuf[:0])
	tagID := TagID{
		0,
		binary.LittleEndian.Uint64(hash[0:8]),
		binary.LittleEndian.Uint64(hash[8:16]),
	}
	return tagID
}

func (id TagID) String() string {
	return id.Base32Suffix()
}

func (tid TagID) CompareTo(oth TagID) int {
	if tid[0] < oth[0] {
		return -1
	}
	if tid[0] > oth[0] {
		return 1
	}
	if tid[1] < oth[1] {
		return -1
	}
	if tid[1] > oth[1] {
		return 1
	}
	return 0
}

func (tid TagID) Add(oth TagID) TagID {
	sum := TagID{
		tid[0] + oth[0],
		tid[1] + oth[1],
	}
	if sum[1] < tid[1] {
		sum[0]++
		sum[1] = ^sum[1] - 1
	}
	return sum
}

func (tid TagID) Sub(oth TagID) TagID {
	diff := TagID{
		tid[0] - oth[0],
		tid[1] - oth[1],
	}
	if diff[1] > tid[1] {
		diff[0]--
		diff[1] = ^diff[1] + 1
	}
	return diff
}

// Returns Unix UTC time in milliseconds
func (tid TagID) UnixMilli() int64 {
	return int64(tid[0]*1000) >> 16
}

// Returns Unix UTC time in seconds
func (tid TagID) Unix() int64 {
	return int64(tid[0]) >> 16
}

func (tid TagID) IsNil() bool {
	return tid[0] == 0 && tid[1] == 0 && tid[2] == 0
}

func (tid *TagID) SetFromInts(x0 int64, x1 uint64) {
	tid[0] = uint64(x0)
	tid[1] = x1
}

func (tid TagID) ToInts() (int64, uint64) {
	return int64(tid[0]), tid[1]
}

// Base32Suffix returns the last few digits of this TID in string form (for easy reading, logs, etc)
func (tid TagID) Base32Suffix() string {
	const lcm_bits = 40 // divisible by 5 (bits) and 8 (bytes).
	const lcm_bytes = lcm_bits / 8

	var suffix [lcm_bytes]byte
	for i := uint(0); i < lcm_bytes; i++ {
		shift := uint(8 * (lcm_bytes - 1 - i))
		suffix[i] = byte(tid[1] >> shift)
	}
	base32 := bufs.Base32Encoding.EncodeToString(suffix[:])
	return base32
}

// Base16Suffix returns the last few digits of this TID in string form (for easy reading, logs, etc)
func (tid TagID) Base16Suffix() string {
	const nibbles = 6
	const HexChars = "0123456789abcdef"

	var suffix [nibbles]byte
	for i := uint(0); i < nibbles; i++ {
		shift := uint(4 * (nibbles - 1 - i))
		hex := byte(tid[1]>>shift) & 0xF
		suffix[i] = HexChars[hex]
	}
	base16 := string(suffix[:])
	return base16
}

// CopyNext copies the given TID and increments it by 1, typically useful for seeking the next entry after a given one.
func (tid TagID) ImprintTag(other TagID) TagID {
	return TagID{
		tid[0] ^ other[0],
		tid[1] ^ other[1],
		tid[2] ^ other[2],
	}
}

// Forms an amp.UID explicitly from two uint64 values.
func IntsToTag(x0 int64, x1, x2 uint64) TagID {
	return TagID{
		uint64(x0),
		x1,
		x2,
	}
}

var gTagIDSeed = uint64(0x3773000000003773)

var (
	NilTag         = TagID{}
	DevicePlanet   = IntsToTag(0, 0, 1)
	HostPlanet     = IntsToTag(0, 0, 2)
	AppHomePlanet  = IntsToTag(0, 0, 3)
	UserHomePlanet = IntsToTag(0, 0, 4)
)

func BytesToTag(in []byte) (tid TagID, err error) {
	var buf [24]byte
	startAt := max(0, 24-len(in))
	copy(buf[startAt:], in)

	tid[0] = binary.BigEndian.Uint64(buf[0:8])
	tid[1] = binary.BigEndian.Uint64(buf[8:16])
	tid[2] = binary.BigEndian.Uint64(buf[16:24])
	return tid, nil
}

func (tid TagID) AppendTo(dst []byte) []byte {
	dst = binary.BigEndian.AppendUint64(dst, tid[0])
	dst = binary.BigEndian.AppendUint64(dst, tid[1])
	dst = binary.BigEndian.AppendUint64(dst, tid[2])
	return dst
}

/*
// ParseUID decodes s into a UID or returns an error.  Accepted forms:
//   - xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   - urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   - {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
//   - xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
func ParseUUID(s string) (TagID, error) {
	uidBytes, err := uuid.Parse(s)
	if err != nil {
		return TagID{}, err
	}
	return Bin24ToTag(uidBytes[:])
}

// MustParseUID decodes s into a UID or panics -- see ParseUID().
func MustParseUID(s string) UID {
	uidBytes := uuid.MustParse(s)
	uid, _ := BytesToUID(uidBytes[:])
	return uid
}

// String returns the string form of uid: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx or "" if uuid is zero.
func (uid UID) ToUUID() (uuid uuid.UUID) {
	binary.BigEndian.PutUint64(uuid[:8], uid[0])
	binary.BigEndian.PutUint64(uuid[8:], uid[1])
	return uuid
}

// String returns the string form of uid: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx or "" if uuid is zero.
func (uid UID) String() string {
	return uid.ToUUID().String()
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

// SetTimeAndHash writes the given timestamp and the right-most part of inSig into this TID.
//
// See comments for TIDBinaryLen
func (tid TID) SetTimeAndHash(time UTC16, hash []byte) {
	tid.SetUTC(time)
	tid.SetHash(hash)
}

// SetHash sets the sig/hash portion of this ID
func (tid TID) SetHash(hash []byte) {
	const TIDHashSz = int(Const_TIDBinaryLen - 8)
	pos := len(hash) - TIDHashSz
	if pos >= 0 {
		copy(tid[TIDHashSz:], hash[pos:])
	} else {
		for i := 8; i < int(Const_TIDBinaryLen); i++ {
			tid[i] = hash[i]
		}
	}
}

// SelectEarlier looks in inTime a chooses whichever is earlier.
//
// If t is later than the time embedded in this TID, then this function has no effect and returns false.
//
// If t is earlier, then this TID is initialized to t (and the rest zeroed out) and returns true.
func (tid TID) SelectEarlier(t UTC16) bool {

	TIDt := tid.ExtractUTC()

	// Timestamp of 0 is reserved and should only reflect an invalid/uninitialized TID.
	if t < 0 {
		t = 0
	}

	if t < TIDt || t == 0 {
		tid.SetUTC(t)
		for i := 8; i < len(tid); i++ {
			tid[i] = 0
		}
		return true
	}

	return false
}

// func StringToTagID(s string) TagID {
// 	uid := StringToTagID(s)
// 	return [3]uint64{
// 	   0,
// 	   uid[0],
// 	   uid[1],
// 	}
// }

// func (id *TagID) AssignFromU64(x0, x1 uint64) {
// 	binary.BigEndian.PutUint64(id[0:8], x0)
// 	binary.BigEndian.PutUint64(id[8:16], x1)
// }

// func (id *TagID) ExportAsU64() (x0, x1 uint64) {
// 	x0 = binary.BigEndian.Uint64(id[0:8])
// 	x1 = binary.BigEndian.Uint64(id[8:16])
// 	return
// }

// func (id *TagID) String() string {
// 	return bufs.Base32Encoding.EncodeToString(id[:])
// }

/*
// Issues a TagID using the given random number generator to generate the TagID hash portion.
func IssueTagID(rng *rand.Rand) (id TagID) {
	return TagIDFromU64(uint64(ConvertToUTC16(time.Now())), rng.Uint64())
}

// TID identifies a Tx (or Cell) by secure hash ID.
type TxID struct {
	UTC16 UTC16
	Hash1 uint64
	Hash2 uint64
	Hash3 uint64
}

// Base32 returns this TID in Base32 form.
func (tid *TxID) Base32() string {
	var bin [TIDBinaryLen]byte
	binStr := tid.AppendAsBinary(bin[:0])
	return bufs.Base32Encoding.EncodeToString(binStr)
}

// Appends the base 32 ASCII encoding of this TID to the given buffer
func (tid *TxID) AppendAsBase32(io []byte) []byte {
	L := len(io)

	needed := L + TIDStringLen
	dst := io
	if needed > cap(dst) {
		dst = make([]byte, (needed+0x100)&^0xFF)
		dst = append(dst[:0], io...)
	}
	dst = dst[:needed]

	var bin [TIDBinaryLen]byte
	binStr := tid.AppendAsBinary(bin[:0])

	bufs.Base32Encoding.Encode(dst[L:needed], binStr)
	return dst
}

// Appends the base 32 ASCII encoding of this TID to the given buffer
func (tid *TxID) AppendAsBinary(io []byte) []byte {
	L := len(io)
	needed := L + TIDBinaryLen
	dst := io
	if needed > cap(dst) {
		dst = make([]byte, needed)
		dst = append(dst[:0], io...)
	}
	dst = dst[:needed]

	binary.BigEndian.PutUint64(dst[L+0:L+8], uint64(tid.UTC16))
	binary.BigEndian.PutUint64(dst[L+8:L+16], tid.Hash1)
	binary.BigEndian.PutUint64(dst[L+16:L+24], tid.Hash2)
	binary.BigEndian.PutUint64(dst[L+24:L+32], tid.Hash3)
	return dst
}
*/

/*

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
			if ai.TypedName == field.Name {
				for _, msg := range msgs {
					if msg.TagSpecID == ai.TagSpecID {
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
		msg.ValType = ValType_SchemaID.Ord()
		msg.ValInt = int64(schema.SchemaID)
		msg.ValBuf = append(append(msg.ValBuf[:0], []byte(ctx.StateScope())...), []byte(subKey)...)

		numFields := src.NumField()
		valType := src.Type()

		for _, attr := range schema.Attrs {
			msg := tx.AddMsg()
			msg.Op = MsgOp_PushAttr
			msg.TagSpecID = attr.TagSpecID
			for i := 0; i < numFields; i++ {
				if valType.Field(i).Name == attr.TypedName {
					msg.setVal(src.Field(i).Interface())
					break
				}
			}
			if msg.ValType == ValType_nil.Ord() {
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

func (req *CellReq) PushBeginPin(target TagID) {
	m := NewTxMsg()
	m.TagID = target.U64()
	m.Op = MsgOp_PinCell
	req.PushTx(m)
}

func (req *CellReq) PushInsertCell(target TagID, schema *AttrSchema) {
	if schema != nil {
		m := NewTxMsg()
		m.TagID = target.U64()
		m.Op = MsgOp_InsertChildCell
		m.ValType = int32(ValType_SchemaID)
		m.ValInt = int64(schema.SchemaID)
		req.PushTx(m)
	}
}

// Pushes the given attr to the client
func (req *CellReq) PushAttr(target TagID, schema *AttrSchema, attrURI string, val Value) {
	attr := schema.LookupAttr(attrURI)
	if attr == nil {
		return
	}

	m := NewTxMsg()
	m.TagID = target.U64()
	m.Op = MsgOp_PushAttr
	m.TagSpecID = attr.TagSpecID
	if attr.SeriesType == SeriesType_Fixed {
		m.SI = attr.BoundSI
	}
	val.MarshalToMsg(m)
	if attr.ValTypeID != 0 { // what is this for!?
		m.ValType = int32(attr.ValTypeID)
	}
	req.PushTx(m)
}

func (req *CellReq) PushCheckpoint(err error) {
	m := NewTxMsg()
	m.Op = MsgOp_Commit
	m.TagID = req.PinCell.U64()
	if err != nil {
		m.setVal(err)
	}
	req.PushTx(m)
}

*/
