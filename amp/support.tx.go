package amp

import (
	"encoding/binary"
	"io"
	"sync"
	"sync/atomic"
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
		tid := NewTimeID()
		tx.GenesisID_0 = int64(tid[0])
		tx.GenesisID_1 = tid[1]
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
func (tx *TxInfo) RouteTo() TimeID {
	return TimeID{uint64(tx.RouteTo_0), tx.RouteTo_1}
}

func (tx *TxInfo) GenesisID() TimeID {
	return TimeID{uint64(tx.GenesisID_0), tx.GenesisID_1}
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

func MarshalMetaAttr(attrSpec string, attrVal ElemVal) (*TxMsg, error) {
	if attrSpec != "" {
		attrSpec = attrVal.ElemTypeName()
	}
	attrID, err := FormAttrID(attrSpec)
	if err != nil {
		return nil, err
	}

	metaOp := TxOp{
		OpCode: TxOpCode_MetaAttr,
		AttrID: attrID,
	}
	tx := NewTxMsg(true)
	if err = tx.MarshalOpValue(&metaOp, attrVal); err != nil {
		return nil, err
	}
	return tx, nil
}

func (tx *TxMsg) UnmarshalOpValue(idx int, out ElemVal) error {
	if idx < 0 || idx >= len(tx.Ops) {
		return ErrCode_MalformedTx.Error("UnmarshalElemVal: index out of range")
	}
	op := tx.Ops[idx]
	ofs := op.DataStoreOfs
	return out.Unmarshal(tx.DataStore[ofs : ofs+op.DataLen])
}

// If reqID == 0, then this sends an attr to the client's session controller (vs a specific request)
func SendMetaAttr(sess HostSession, routeTo TimeID, status ReqStatus, val ElemVal) error {
	tx, err := MarshalMetaAttr("", val)
	if err != nil {
		return err
	}

	tx.RouteTo_0 = int64(routeTo[0])
	tx.RouteTo_1 = routeTo[1]
	tx.Status = status
	return sess.SendTx(tx)
}

func (tx *TxMsg) ExtractMetaAttr(reg Registry) (ElemVal, error) {
	if len(tx.Ops) == 0 || tx.Ops[0].OpCode != TxOpCode_MetaAttr {
		return nil, ErrCode_MalformedTx.Error("expected meta attr")
	}

	attrID := tx.Ops[0].AttrID
	val, err := reg.NewAttrElem(attrID)
	if err != nil {
		return nil, err
	}
	if err = tx.UnmarshalOpValue(0, val); err != nil {
		return nil, err
	}
	return val, nil
}

// On success:
//   - TxOp.DataStoreOfs and TxOp.DataLen are overwritten,
//   - TxMsg.DataStore is appended with the serialization of val, and
//   - the TxOp is appended to TxMsg.Ops.
func (tx *TxMsg) MarshalOpValue(op *TxOp, val ElemVal) error {
	if val == nil {
		op.DataStoreOfs = 0
		op.DataLen = 0
	} else {
		var err error
		op.DataStoreOfs = int64(len(tx.DataStore))
		tx.DataStore, err = val.MarshalToStore(tx.DataStore)
		if err != nil {
			return err
		}
		op.DataLen = int64(len(tx.DataStore)) - op.DataStoreOfs
	}

	tx.NumOps += 1
	tx.Ops = append(tx.Ops, *op)
	return nil
}

func (tx *TxMsg) MarshalOpWithBuf(op *TxOp, valBuf []byte) {
	op.DataStoreOfs = int64(len(tx.DataStore))
	op.DataLen = int64(len(valBuf))
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

		buf := tx.DataStore[:bodyLen]
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

	bodyLen := len(headerBody) - len(header)
	binary.LittleEndian.PutUint32(header[4:8], uint32(bodyLen))
	binary.LittleEndian.PutUint32(header[8:12], uint32(len(tx.DataStore)))

	*dst = headerBody
}

func (tx *TxMsg) MarshalBody(dst []byte) []byte {

	// TxInfo
	{
		tx.NumOps = uint32(len(tx.Ops))
		infoLen := tx.TxInfo.Size()
		dst = binary.AppendVarint(dst, int64(infoLen))

		p := len(dst)
		dst = append(dst, make([]byte, infoLen)...)
		tx.TxInfo.MarshalToSizedBuffer(dst[p : p+infoLen])
	}

	var (
		op_prv [TxBody_MaxFields]uint64
		op_cur [TxBody_MaxFields]uint64
	)

	for _, op := range tx.Ops {

		dst = binary.AppendVarint(dst, 0) // skip bytes (future use)
		dst = binary.AppendVarint(dst, int64(op.OpCode))
		dst = binary.AppendVarint(dst, op.DataStoreOfs)
		dst = binary.AppendVarint(dst, op.DataLen)

		// detect body repeated fields and write only what changes (with corresponding flags)
		{
			op_cur[TxBody_ParentIDx0] = op.ParentID[0]
			op_cur[TxBody_ParentIDx1] = op.ParentID[1]
			op_cur[TxBody_ParentIDx2] = op.ParentID[2]

			op_cur[TxBody_TargetIDx0] = op.TargetID[0]
			op_cur[TxBody_TargetIDx1] = op.TargetID[1]
			op_cur[TxBody_TargetIDx2] = op.TargetID[2]

			op_cur[TxBody_AttrIDx0] = op.AttrID[0]
			op_cur[TxBody_AttrIDx1] = op.AttrID[1]

			op_cur[TxBody_SIx0] = op.SI[0]
			op_cur[TxBody_SIx1] = op.SI[1]

			hasFields := int64(0)
			for i, fi := range op_cur {
				if fi != op_prv[i] {
					hasFields |= (1 << i)
				}
			}

			dst = binary.AppendVarint(dst, hasFields)

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
		infoLen, n := binary.Varint(src[0:])
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
		op_cur [TxBody_MaxFields]uint64
	)

	for i := uint32(0); i < tx.NumOps; i++ {
		var op TxOp
		var n int

		// skip (future use)
		var skip int64
		if skip, n = binary.Varint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n + int(skip)

		// OpCode
		var opCode int64
		if opCode, n = binary.Varint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n
		op.OpCode = TxOpCode(opCode)

		// DataStoreOfs
		if op.DataStoreOfs, n = binary.Varint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n

		// DataLen
		if op.DataLen, n = binary.Varint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n

		var hasFields int64
		if hasFields, n = binary.Varint(src[p:]); n <= 0 {
			return ErrMalformedTx
		}
		p += n

		for i := 0; i < int(TxBody_MaxFields); i++ {
			if hasFields&(1<<i) != 0 {
				if p+8 > len(src) {
					return ErrMalformedTx
				}
				op_cur[i] = binary.LittleEndian.Uint64(src[p:])
				p += 8
			}
		}

		op.ParentID[0] = op_cur[TxBody_ParentIDx0]
		op.ParentID[1] = op_cur[TxBody_ParentIDx1]
		op.ParentID[2] = op_cur[TxBody_ParentIDx2]

		op.TargetID[0] = op_cur[TxBody_TargetIDx0]
		op.TargetID[1] = op_cur[TxBody_TargetIDx1]
		op.TargetID[2] = op_cur[TxBody_TargetIDx2]

		op.AttrID[0] = op_cur[TxBody_AttrIDx0]
		op.AttrID[1] = op_cur[TxBody_AttrIDx1]

		op.SI[0] = op_cur[TxBody_SIx0]
		op.SI[1] = op_cur[TxBody_SIx1]

		tx.Ops = append(tx.Ops, op)
	}

	return nil
}

func (op *TxOp) Validate() error {
	if op.TargetID.IsNil() {
		return ErrBadTarget
	}
	if op.AttrID.IsNil() {
		return ErrCode_MalformedTx.Error("missing AttrID")
	}
	return nil
}

// func (op *TxOp) TargetID() CellID {
// 	return [3]uint64{
// 		op.TargetIDx0,
// 		op.TargetIDx1,
// 		op.TargetIDx2,
// 	}
// }

// func (op *TxOp) AttrID() AttrID {
// 	return [2]uint64{
// 		op.AttrIDx0,
// 		op.AttrIDx1,
// 	}
// }

// func (op *TxOp) SI() SeriesIndex {
// 	return [2]uint64{
// 		op.SIx0,
// 		op.SIx1,
// 	}
// }

// func (op *TxOp) ParentID() CellID {
// 	return [3]uint64{
// 		op.ParentIDx0,
// 		op.ParentIDx1,
// 		op.ParentIDx2,
// 	}
// }

// func (op *TxOp) TargetCell() CellID {
// 	return [3]uint64{
// 		op.TargetIDx0,
// 		op.TargetIDx1,
// 		op.TargetIDx2,
// 	}
// }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
