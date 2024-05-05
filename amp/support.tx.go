package amp

import (
	"encoding/binary"
	"io"
	"sync"
	"sync/atomic"

	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
)

// TxDataStore is a message packet sent to / from a client.
// It is leads with a fixed-size header (TxHeader_Size) followed by a variable-size body.
type TxDataStore []byte

// TxHeader is the fixed-size header that leads every TxMsg.
// See comments for Const_TxHeader_Size.
type TxHeader [Const_TxHeader_Size]byte

func (header TxHeader) TxBodyLen() int {
	return int(binary.LittleEndian.Uint32(header[4:8]))
}

func (header TxHeader) TxDataLen() int {
	return int(binary.LittleEndian.Uint32(header[8:12]))
}

func NewTxMsg(genesis bool) *TxMsg {
	tx := gTxMsgPool.Get().(*TxMsg)
	tx.refCount = 1
	if genesis {
		tid := tag.New()
		tx.GenesisID_0 = int64(tid[0])
		tx.GenesisID_1 = tid[1]
		tx.GenesisID_2 = tid[2]
	}

	return tx
}

var gTxMsgPool = sync.Pool{
	New: func() interface{} {
		return &TxMsg{}
	},
}

/*
	func CopyMsg(src *TxMsg) *TxMsg {
		msg := NewMsg()

		if src != nil {
			valBuf := append(msg.ValBuf[:0], src.ValBuf...)
			*msg = *src
			msg.ValBuf = valBuf

		}
		return msg
	}
*/
func (tx *TxInfo) ContextID() tag.ID {
	return tag.ID{uint64(tx.ContextID_0), tx.ContextID_1, tx.ContextID_2}
}

func (tx *TxInfo) GenesisID() tag.ID {
	return tag.ID{uint64(tx.GenesisID_0), tx.GenesisID_1, tx.GenesisID_2}
}

func (tx *TxMsg) AddRef() {
	atomic.AddInt32(&tx.refCount, 1)
}

func (tx *TxMsg) ReleaseRef() {
	if atomic.AddInt32(&tx.refCount, -1) > 0 {
		return
	}

	*tx = TxMsg{
		Ops:       tx.Ops[:0],
		DataStore: tx.DataStore[:0],
	}
	gTxMsgPool.Put(tx)
}

func MarshalMetaAttr(attrSpec tag.Spec, attrVal ElemVal) (*TxMsg, error) {
	if attrSpec.ID.IsNil() {
		attrSpec = tag.FormSpec(attrSpec, attrVal.ElemTypeName())
	}

	tx := NewTxMsg(true)
	metaOp := TxOp{
		OpCode: TxOpCode_MetaAttr,
		AttrID: attrSpec.ID,
	}
	if err := tx.MarshalOp(&metaOp, attrVal); err != nil {
		return nil, err
	}
	return tx, nil
}

func (tx *TxMsg) UnmarshalOpValue(idx int, out ElemVal) error {
	if idx < 0 || idx >= len(tx.Ops) {
		return ErrCode_MalformedTx.Error("UnmarshalElemVal: index out of range")
	}
	op := tx.Ops[idx]
	ofs := op.DataOfs
	return out.Unmarshal(tx.DataStore[ofs : ofs+op.DataLen])
}

// If reqID == 0, then this sends an attr to the client's session controller (vs a specific request)
func SendMetaAttr(sess HostSession, context tag.ID, status ReqStatus, val ElemVal) error {
	tx, err := MarshalMetaAttr(tag.Spec{}, val)
	if err != nil {
		return err
	}

	tx.ContextID_0, tx.ContextID_1, tx.ContextID_2 = context.ToInts()
	tx.Status = status
	return sess.SendTx(tx)
}

func (tx *TxMsg) ExtractMetaAttr(reg Registry) (ElemVal, error) {
	if len(tx.Ops) == 0 || tx.Ops[0].OpCode != TxOpCode_MetaAttr {
		return nil, ErrCode_MalformedTx.Error("expected meta attr")
	}

	val, err := reg.NewAttrElem(tx.Ops[0].AttrID)
	if err != nil {
		return nil, err
	}
	if err = tx.UnmarshalOpValue(0, val); err != nil {
		return nil, err
	}
	return val, nil
}

func (tx *TxMsg) MarshalUpsert(targetID tag.ID, attrSpecID tag.ID, val ElemVal) error {
	op := TxOp{
		OpCode:   TxOpCode_UpsertAttr,
		TargetID: targetID,
		AttrID:   attrSpecID,
	}
	return tx.MarshalOp(&op, val)
}

// Marshals a TxOp and optional value to the given Tx's to and data store.
//
// On success:
//   - TxOp.DataOfs and TxOp.DataLen are overwritten,
//   - TxMsg.DataStore is appended with the serialization of val, and
//   - the TxOp is appended to TxMsg.Ops.
func (tx *TxMsg) MarshalOp(op *TxOp, val ElemVal) error {
	err := op.Validate()
	if err != nil {
		return err
	}
	if val == nil {
		op.DataOfs = 0
		op.DataLen = 0
	} else {
		var err error
		op.DataOfs = uint64(len(tx.DataStore))
		tx.DataStore, err = val.MarshalToStore(tx.DataStore)
		if err != nil {
			return err
		}
		op.DataLen = uint64(len(tx.DataStore)) - op.DataOfs
	}

	tx.NumOps += 1
	tx.Ops = append(tx.Ops, *op)
	return nil
}

func (tx *TxMsg) MarshalOpWithBuf(op *TxOp, valBuf []byte) {
	op.DataOfs = uint64(len(tx.DataStore))
	op.DataLen = uint64(len(valBuf))
	tx.DataStore = append(tx.DataStore, valBuf...)
	tx.NumOps += 1
	tx.Ops = append(tx.Ops, *op)
}

func ReadTxMsg(stream io.Reader) (*TxMsg, error) {
	readBytes := func(dst []byte) error {
		for L := 0; L < len(dst); {
			n, err := stream.Read(dst[L:])
			if err != nil {
				return err
			}
			L += n
		}
		return nil
	}

	var header TxHeader
	if err := readBytes(header[:]); err != nil {
		return nil, err
	}

	marker := uint32(header[0])<<16 | uint32(header[1])<<8 | uint32(header[2])
	if marker != uint32(Const_TxHeader_Marker) {
		return nil, ErrMalformedTx
	}
	if header[3] < byte(Const_TxHeader_Version) {
		return nil, ErrMalformedTx
	}

	tx := NewTxMsg(false)
	bodyLen := header.TxBodyLen()
	dataLen := header.TxDataLen()

	// Use tx.DataStore to hold the body for unmarshalling.
	// The tx body contains TxMsg fields and TxOps
	{
		needSz := max(bodyLen, dataLen)
		if cap(tx.DataStore) < needSz {
			tx.DataStore = make([]byte, max(needSz, 2048))
		}

		buf := tx.DataStore[:bodyLen-int(Const_TxHeader_Size)]
		if err := readBytes(buf); err != nil {
			return nil, err
		}
		if err := tx.UnmarshalBody(buf); err != nil {
			return nil, err
		}
	}

	// Read tx data store -- used for on-demand ElemVal unmarshalling
	tx.DataStore = tx.DataStore[:dataLen]
	if err := readBytes(tx.DataStore); err != nil {
		return nil, err
	}

	return tx, nil
}

func (tx *TxMsg) MarshalToWriter(scrap *[]byte, w io.Writer) (err error) {
	writeBytes := func(src []byte) error {
		for L := 0; L < len(src); {
			n, err := w.Write(src[L:])
			if err != nil {
				return err
			}
			L += n
		}
		return nil
	}

	tx.MarshalHeaderAndBody(scrap)
	if err = writeBytes(*scrap); err != nil {
		return
	}
	if err = writeBytes(tx.DataStore); err != nil {
		return
	}
	return
}

func (tx *TxMsg) MarshalToBuffer(dst *[]byte) {
	tx.MarshalHeaderAndBody(dst)
	*dst = append(*dst, tx.DataStore...)
}

func (tx *TxMsg) MarshalHeaderAndBody(dst *[]byte) {
	buf := (*dst)[:0]
	if cap(buf) < 300 {
		buf = make([]byte, 2048)
	}

	headerBody := tx.MarshalBody(buf[:Const_TxHeader_Size])

	header := headerBody[:Const_TxHeader_Size]
	header[0] = byte((Const_TxHeader_Marker >> 16) & 0xFF)
	header[1] = byte((Const_TxHeader_Marker >> 8) & 0xFF)
	header[2] = byte((Const_TxHeader_Marker >> 0) & 0xFF)
	header[3] = byte(Const_TxHeader_Version)

	binary.LittleEndian.PutUint32(header[4:8], uint32(len(headerBody)))
	binary.LittleEndian.PutUint32(header[8:12], uint32(len(tx.DataStore)))

	*dst = headerBody
}

func (tx *TxMsg) MarshalBody(dst []byte) []byte {

	// TxInfo
	{
		tx.NumOps = uint64(len(tx.Ops))
		infoLen := tx.TxInfo.Size()
		dst = binary.AppendUvarint(dst, uint64(infoLen))

		p := len(dst)
		dst = append(dst, make([]byte, infoLen)...)
		tx.TxInfo.MarshalToSizedBuffer(dst[p : p+infoLen])
	}

	var (
		op_prv [TxField_MaxFields]uint64
		op_cur [TxField_MaxFields]uint64
	)

	for _, op := range tx.Ops {

		dst = binary.AppendUvarint(dst, 0) // skip bytes (future use)
		dst = binary.AppendUvarint(dst, uint64(op.OpCode))
		dst = binary.AppendUvarint(dst, op.Height)
		dst = binary.AppendUvarint(dst, op.DataLen)
		dst = binary.AppendUvarint(dst, op.DataOfs)

		// detect body repeated fields and write only what changes (with corresponding flags)
		{
			op_cur[TxField_TargetIDx0] = op.TargetID[0]
			op_cur[TxField_TargetIDx1] = op.TargetID[1]
			op_cur[TxField_TargetIDx2] = op.TargetID[2]

			op_cur[TxField_AttrID_0] = op.AttrID[0]
			op_cur[TxField_AttrID_1] = op.AttrID[1]
			op_cur[TxField_AttrID_2] = op.AttrID[2]

			op_cur[TxField_SI_0] = op.SI[0]
			op_cur[TxField_SI_1] = op.SI[1]
			op_cur[TxField_SI_2] = op.SI[2]

			op_cur[TxField_Hash] = op.Hash

			hasFields := uint64(0)
			for i, fi := range op_cur {
				if fi != op_prv[i] {
					hasFields |= (1 << i)
				}
			}

			dst = binary.AppendUvarint(dst, hasFields)

			for i, fi := range op_cur {
				if hasFields&(1<<i) != 0 {
					dst = binary.LittleEndian.AppendUint64(dst, fi)
				}
			}

			op_prv = op_cur
		}
	}

	return dst
}

func (tx *TxMsg) UnmarshalBody(src []byte) error {
	p := 0

	// TxInfo
	{
		infoLen, n := binary.Uvarint(src[0:])
		if n <= 0 {
			return ErrMalformedTx
		}
		p += n

		tx.TxInfo = TxInfo{}
		err := tx.TxInfo.Unmarshal(src[p : p+int(infoLen)])
		if err != nil {
			return ErrMalformedTx
		}
		p += int(infoLen)
	}

	var (
		op_cur [TxField_MaxFields]uint64
	)

	for i := uint64(0); i < tx.NumOps; i++ {
		var op TxOp
		var n int

		// skip (future use)
		var skip uint64
		if skip, n = binary.Uvarint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n + int(skip)

		// OpCode
		var opCode uint64
		if opCode, n = binary.Uvarint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n
		op.OpCode = TxOpCode(opCode)

		// Revision Height
		if op.Height, n = binary.Uvarint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n

		// DataLen
		if op.DataLen, n = binary.Uvarint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n

		// DataOfs
		if op.DataOfs, n = binary.Uvarint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n

		var hasFields uint64
		if hasFields, n = binary.Uvarint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n

		for i := 0; i < int(TxField_MaxFields); i++ {
			if hasFields&(1<<i) != 0 {
				if p+8 > len(src) {
					return ErrMalformedTx
				}
				op_cur[i] = binary.LittleEndian.Uint64(src[p:])
				p += 8
			}
		}

		op.TargetID[0] = op_cur[TxField_TargetIDx0]
		op.TargetID[1] = op_cur[TxField_TargetIDx1]
		op.TargetID[2] = op_cur[TxField_TargetIDx2]

		op.AttrID[0] = op_cur[TxField_AttrID_0]
		op.AttrID[1] = op_cur[TxField_AttrID_1]
		op.AttrID[2] = op_cur[TxField_AttrID_2]

		op.SI[0] = op_cur[TxField_SI_0]
		op.SI[1] = op_cur[TxField_SI_1]
		op.SI[2] = op_cur[TxField_SI_2]

		tx.Ops = append(tx.Ops, op)
	}

	return nil
}

func (op *TxOp) Validate() error {
	if op.TargetID.IsNil() {
		return ErrBadTarget
	}
	if op.AttrID.IsNil() {
		return ErrCode_MalformedTx.Error("missing TagSpecID")
	}
	return nil
}

const (
	kCellOfs   = 0                // 24 byte cell index
	kAttrOfs   = 24               // 24 byte attr index
	kIndexOfs  = 24 + 24          // 24 byte series index
	kHeightOfs = 24 + 24 + 24     // 8 byte height counter
	kHashOfs   = 24 + 24 + 24 + 8 // 8 byte hash
	AttrKeyLen = kHashOfs + 8     // total byte length of key
)

type TxOpKey [AttrKeyLen]byte

func FormKey(dst []byte, tag tag.ID) {
	binary.BigEndian.PutUint64(dst[0:], uint64(tag[0]))
	binary.BigEndian.PutUint64(dst[8:], tag[1])
	binary.BigEndian.PutUint64(dst[16:], tag[2])
}


func (op *TxOp) FormAttrKey() TxOpKey {
	var key TxOpKey
	FormKey(key[kCellOfs:], op.TargetID)
	FormKey(key[kAttrOfs:], op.AttrID)
	FormKey(key[kIndexOfs:], op.SI)

	binary.BigEndian.PutUint64(key[kHeightOfs:], uint64(op.Height))
	binary.BigEndian.PutUint64(key[kHashOfs:], uint64(op.Hash))
	return key
}

func (op *TxOp) ApplyAttrKey(key []byte) error {
	if len(key) < AttrKeyLen {
		return ErrCode_InternalErr.Error("bad db attr key")
	}

	op.TargetID[0] = binary.BigEndian.Uint64(key[kCellOfs:])
	op.TargetID[1] = binary.BigEndian.Uint64(key[kCellOfs+8:])
	op.TargetID[2] = binary.BigEndian.Uint64(key[kCellOfs+16:])

	op.AttrID[0] = binary.BigEndian.Uint64(key[kAttrOfs:])
	op.AttrID[1] = binary.BigEndian.Uint64(key[kAttrOfs+8:])
	op.AttrID[2] = binary.BigEndian.Uint64(key[kAttrOfs+16:])

	op.SI[0] = binary.BigEndian.Uint64(key[kIndexOfs:])
	op.SI[1] = binary.BigEndian.Uint64(key[kIndexOfs+8:])
	op.SI[2] = binary.BigEndian.Uint64(key[kIndexOfs+16:])

	op.Height = binary.BigEndian.Uint64(key[kHeightOfs:])
	op.Hash = binary.BigEndian.Uint64(key[kHashOfs:])
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
