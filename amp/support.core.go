package amp

import (
	"strings"
	"time"

	"github.com/amp-space/amp-sdk-go/stdlib/bufs"
)

// URI form of a glyph typically followed by a media (mime) type.
const GlyphURIPrefix = "amp:glyph/"

// Describes an asset to be an image stream but not specify format / codec
var GenericImageType = "image/*"

// TimeID is a signed 16 byte UTC time index in big endian form, with 80 bits (10 bytes) of fractional precision.
// This means there are 47 bits dedicated for whole seconds => +/- 4.4 million years
//
// Note: (TimeID[0] >> 16) yields a standard 64-bit Unix UTC time.
type TimeID [2]uint64

var NilTimeID = TimeID{}

const (
	TimeID_x0_SecondsShift = 16                  // shift to get the 64-bit Unix time from TimeID[0]
	TimeID_NanosecStep     = uint64(0x44B82FA1C) // 1<<64 div 1/1e9 -- reflects Go's single nanosecond resolution spread over a 64 bits
	TimeID_EntropyMask     = uint64(0x3FFFFFFFF) // entropy bit mask for TimeID[1] -- slightly smaller than 1 ns resolution
)

func TimeIDFromInts(x0 int64, x1 uint64) TimeID {
	return TimeID{uint64(x0), x1}
}

// Returns the current time as a TimeID, statistically guaranteed to be unique even when called in rapid succession.
func NewTimeID() TimeID {
	return ConvertToTimeID(time.Now(), true)
}

func ConvertToTimeID(t time.Time, addEntropy bool) TimeID {
	ns_b10 := uint64(t.Nanosecond())
	ns_f64 := ns_b10 * TimeID_NanosecStep // map 0..999999999 to 0..(2^64-1)

	t_00_06 := uint64(t.Unix()) << TimeID_x0_SecondsShift
	t_07_08 := ns_f64 >> (64 - TimeID_x0_SecondsShift)
	t_09_15 := ns_f64 << TimeID_x0_SecondsShift
	if addEntropy {
		gTimeIDSeed = ns_f64 ^ (gTimeIDSeed * 5237732522753)
		t_09_15 ^= gTimeIDSeed & TimeID_EntropyMask
	}

	tid := TimeID{
		t_00_06 | uint64(t_07_08),
		t_09_15,
	}
	return tid
}

func (tid TimeID) CompareTo(oth TimeID) int {
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

func (tid TimeID) Add(oth TimeID) TimeID {
	sum := TimeID{
		tid[0] + oth[0],
		tid[1] + oth[1],
	}
	if sum[1] < tid[1] {
		sum[0]++
		sum[1] = ^sum[1] - 1
	}
	return sum
}

func (tid TimeID) Sub(oth TimeID) TimeID {
	diff := TimeID{
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
func (tid TimeID) UnixMilli() int64 {
	return int64(tid[0]*1000) >> 16
}

// Returns Unix UTC time in seconds
func (tid TimeID) Unix() int64 {
	return int64(tid[0]) >> 16
}

func (tid TimeID) IsNil() bool {
	return tid[0] == 0 && tid[1] == 0
}

func (tid *TimeID) SetFromInts(x0 int64, x1 uint64) {
	tid[0] = uint64(x0)
	tid[1] = x1
}

func (tid TimeID) ToInts() (int64, uint64) {
	return int64(tid[0]), tid[1]
}

// Base32Suffix returns the last few digits of this TID in string form (for easy reading, logs, etc)
func (tid TimeID) Base32Suffix() string {
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

const HexChars = "0123456789abcdef"

// Base16Suffix returns the last few digits of this TID in string form (for easy reading, logs, etc)
func (tid TimeID) Base16Suffix() string {
	const nibbles = 6

	var suffix [nibbles]byte
	for i := uint(0); i < nibbles; i++ {
		shift := uint(4 * (nibbles - 1 - i))
		hex := byte(tid[1]>>shift) & 0xF
		suffix[i] = HexChars[hex]
	}
	base16 := string(suffix[:])
	return base16
}

var gTimeIDSeed = uint64(0x3773000000003773)

// UTC16 is a signed UTC timestamp, storing the elapsed 1/65536 second ticks since Jan 1, 1970 UTC.
//
// Shifting this value to the right 16 bits will yield standard Unix time.
// This means there are 47 bits dedicated for seconds, implying a limit 4.4 million years.
//type UTC16 int64

// TID identifies a specific planet, node, or transaction.
//
// Unless otherwise specified a TID in the wild should always be considered read-only.
// type TID []byte

// // TIDBuf is embedded UTC16 value followed by a 24 byte hash.
// type TIDBuf [TIDBinaryLen]byte

// // Byte size of a TID, a hash with a leading embedded big endian binary time index.
// const TIDBinaryLen = int(Const_TIDBinaryLen)

// // ASCII string length of a CellTID encoded into its base32 form.
// const TIDStringLen = int(Const_TIDStringLen)

// // nilTID is a zeroed TID that denotes a void/nil/zero value of a TID
// var nilTID = TID{}

// const (
// 	SI_DistantFuture = UTC16(0x7FFFFFFFFFFFFFFF)
// )

/*
// Converts milliseconds to UTC16.
func ConvertMsToUTC(ms int64) UTC16 {
	return UTC16((ms << 16) / 1000)
}

// Converts UTC16 to a time.Time.
func (t UTC16) ToTime() time.Time {
	return time.Unix(int64(t>>16), int64(t&0xFFFF)*15259)
}


// TID is a convenience function that returns the TID contained within this TxID.
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

// Buf is a convenience function that make a new TxID from a TID byte slice.
func (tid TID) Buf() (buf TIDBuf) {
	copy(buf[:], tid)
	return buf
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

// SetUTC16 writes the given UTC16 into this TID
func (tid TID) SetUTC(t UTC16) {
	tid[0] = byte(t >> 56)
	tid[1] = byte(t >> 48)
	tid[2] = byte(t >> 40)
	tid[3] = byte(t >> 32)
	tid[4] = byte(t >> 24)
	tid[5] = byte(t >> 16)
	tid[6] = byte(t >> 8)
	tid[7] = byte(t)
}

// ExtractUTC16 returns the unix timestamp embedded in this TID (a unix timestamp in 1<<16 seconds UTC)
func (tid TID) ExtractUTC() UTC16 {
	t := int64(tid[0])
	t = (t << 8) | int64(tid[1])
	t = (t << 8) | int64(tid[2])
	t = (t << 8) | int64(tid[3])
	t = (t << 8) | int64(tid[4])
	t = (t << 8) | int64(tid[5])
	t = (t << 8) | int64(tid[6])
	t = (t << 8) | int64(tid[7])

	return UTC16(t)
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
*/
// // Forms a CellID from uint64s.
// func CellIDFromU64(x0, x1 uint64) (id CellID) {
// 	id.AssignFromU64(x0, x1)
// 	return id
// }

func (id *CellID) IsNil() bool {
	return id[0] == 0 && id[1] == 0 && id[2] == 0
}

// func StringToCellID(s string) CellID {
// 	uid := StringToUID(s)
// 	return [3]uint64{
// 	   0,
// 	   uid[0],
// 	   uid[1],
// 	}
// }

// func (id *CellID) AssignFromU64(x0, x1 uint64) {
// 	binary.BigEndian.PutUint64(id[0:8], x0)
// 	binary.BigEndian.PutUint64(id[8:16], x1)
// }

// func (id *CellID) ExportAsU64() (x0, x1 uint64) {
// 	x0 = binary.BigEndian.Uint64(id[0:8])
// 	x1 = binary.BigEndian.Uint64(id[8:16])
// 	return
// }

// func (id *CellID) String() string {
// 	return bufs.Base32Encoding.EncodeToString(id[:])
// }

/*
// Issues a CellID using the given random number generator to generate the UID hash portion.
func IssueCellID(rng *rand.Rand) (id CellID) {
	return CellIDFromU64(uint64(ConvertToUTC16(time.Now())), rng.Uint64())
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

// Analyses an AttrSpec's SeriesSpec and returns the index class it uses.
func GetSeriesIndexType(seriesSpec string) SeriesIndexType {
	switch {
	case strings.HasSuffix(seriesSpec, ".Name"):
		return SeriesIndexType_Name
	default:
		return SeriesIndexType_Literal
	}
}


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
		msg.ValType = ValType_SchemaID.Ord()
		msg.ValInt = int64(schema.SchemaID)
		msg.ValBuf = append(append(msg.ValBuf[:0], []byte(ctx.StateScope())...), []byte(subKey)...)

		numFields := src.NumField()
		valType := src.Type()

		for _, attr := range schema.Attrs {
			msg := tx.AddMsg()
			msg.Op = MsgOp_PushAttr
			msg.AttrID = attr.AttrID
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

func (req *CellReq) PushBeginPin(target CellID) {
	m := NewTxMsg()
	m.CellID = target.U64()
	m.Op = MsgOp_PinCell
	req.PushTx(m)
}

func (req *CellReq) PushInsertCell(target CellID, schema *AttrSchema) {
	if schema != nil {
		m := NewTxMsg()
		m.CellID = target.U64()
		m.Op = MsgOp_InsertChildCell
		m.ValType = int32(ValType_SchemaID)
		m.ValInt = int64(schema.SchemaID)
		req.PushTx(m)
	}
}

// Pushes the given attr to the client
func (req *CellReq) PushAttr(target CellID, schema *AttrSchema, attrURI string, val Value) {
	attr := schema.LookupAttr(attrURI)
	if attr == nil {
		return
	}

	m := NewTxMsg()
	m.CellID = target.U64()
	m.Op = MsgOp_PushAttr
	m.AttrID = attr.AttrID
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
	m.CellID = req.PinCell.U64()
	if err != nil {
		m.setVal(err)
	}
	req.PushTx(m)
}

*/
