package arc

import (
	"encoding/binary"
	"io"
	"sync"
)

// TxDataStore is a message packet sent to / from a client.
// It is leads with a fixed-size header (TxHeader_Size) followed by a variable-size body.
type TxDataStore []byte

type TxHeader [Const_TxHeader_Size]byte

// func (tx TxDataStore) GetTxTotalLen() int {
// 	if tx[TxHeader_OpOfs] != byte(TxHeader_OpRecvTx) {
// 		return 0
// 	}
// 	bodySz := int(binary.BigEndian.Uint32(tx[3:7]))
// 	return bodySz
// }

// func (tx TxDataStore) SetTxBodyLen(bodyLen int) {
// 	txLen := bodyLen + int(TxHeader_Size)
// 	binary.BigEndian.PutUint32(tx[3:7], uint32(txLen))
// 	tx[TxHeader_OpOfs] = byte(TxHeader_OpRecvTx)
// }

// func (tx TxDataStore) InitHeader(bodyLen int) {
// 	tx[0] = 0
// 	tx[1] = 0
// 	tx[2] = 0
// 	txLen := bodyLen + int(TxHeader_Size)
// 	binary.BigEndian.PutUint32(tx[3:7], uint32(txLen))
// 	tx[TxHeader_OpOfs] = byte(TxHeader_OpRecvTx)
// }

// ///////////////////////// host -> client /////////////////////////

// func (msg *TxMsg) MarshalToTxBuffer(txBuf []byte) error {
// 	n, err := msg.MarshalToSizedBuffer(txBuf[TxHeader_Size:])
// 	if err != nil {
// 		return err
// 	}
// 	TxDataStore(txBuf).InitHeader(n)
// 	return nil
// }

func (hdr TxHeader) TxBodyLen() int {
	return int(binary.BigEndian.Uint32(hdr[4:8]))
}

func (hdr TxHeader) TxDataLen() int {
	return int(binary.BigEndian.Uint32(hdr[8:12]))
}

func (tx *TxMsg) FormTxHeader(txBodyLen int) TxHeader {
	hdr := TxHeader{}
	hdr[0] = byte(Const_TxHeader_Size)
	hdr[1] = byte((Const_TxHeader_Marker >> 16) & 0xFF)
	hdr[2] = byte((Const_TxHeader_Marker >> 8) & 0xFF)
	hdr[3] = byte((Const_TxHeader_Marker >> 0) & 0xFF)
	binary.BigEndian.PutUint32(hdr[4:8], uint32(txBodyLen))
	binary.BigEndian.PutUint32(hdr[8:12], uint32(len(tx.DataStore)))
	return hdr
}

func (hdr TxHeader) NewTxMsg() (*TxMsg, error) {
	if hdr[0] != byte(Const_TxHeader_Size) {
		return nil, ErrMalformedTx
	}
	marker := uint32(hdr[1])<<16 | uint32(hdr[2])<<8 | uint32(hdr[3])
	if marker != uint32(Const_TxHeader_Marker) {
		return nil, ErrMalformedTx
	}

	return NewTxMsg(), nil
}

func NewTxMsg() *TxMsg {
	tx := gTxMsgPool.Get().(*TxMsg)
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

func (tx *TxMsg) Init() {
	*tx = TxMsg{
		Ops: tx.Ops[:0],
	}
}

func (tx *TxMsg) Reclaim() {
	if tx != nil {
		tx.Init()
		gTxMsgPool.Put(tx)
	}
}

// func (msg *TxMsg) MarshalAttrElem(attrID uint32, src PbValue) error {
// 	msg.AttrID = attrID
// 	sz := src.Size()
// 	if sz > cap(msg.ValBuf) {
// 		msg.ValBuf = make([]byte, sz, (sz+0x3FF)&^0x3FF)
// 	} else {
// 		msg.ValBuf = msg.ValBuf[:sz]
// 	}
// 	_, err := src.MarshalToSizedBuffer(msg.ValBuf)
// 	return err
// }

// func (msg *TxMsg) UnmarshalValue(dst PbValue) error {
// 	return dst.Unmarshal(msg.ValBuf)
// }

// func (attr AttrElemVal) MarshalToMsg(id CellID) (*TxMsg, error) {
// 	msg := NewMsg()
// 	msg.Op = MsgOp_PushAttrElem
// 	msg.AttrID = attr.AttrID
// 	msg.SI = attr.SI
// 	msg.CellID = int64(id)
// 	err := attr.Val.MarshalToBuf(&msg.ValBuf)
// 	return msg, err
// }

// type CellMarshaller struct {
// 	Txs []*CellTxPb

// 	marshalBuf []byte
// 	fatalErr   error
// }

// func (tx *AttrOp) Clear(op AttrOpCode) {
// 	tx.Op = op
// 	tx.CellID = CellID{}
// 	tx.Elems = tx.Elems[:0]
// }

// func (op *TxOp) MarshalToStore(tx *TxMsg, val AttrElemVal) error {
// 	var err error

// 	op.DataStoreOfs = int64(len(tx.DataStore))
// 	tx.DataStore, err = val.MarshalToStore(tx.DataStore)
// 	if err != nil {
// 		return err
// 	}
// 	op.DataLen = int64(len(tx.DataStore))
// }

func (op *TxOp) HasAttrUID() bool {
	return op.AttrID[0] != 0 || op.AttrID[1] != 0
}

func (op *TxOp) NilAttrUID() bool {
	return op.AttrID[0] == 0 && op.AttrID[1] == 0
}

/*
func (tx *CellTx) MarshalAttrs() error {
	if cap(tx.ElemsPb) < len(tx.Elems) {
		tx.ElemsPb = make([]*AttrElemPb, len(tx.Elems))
	} else {
		tx.ElemsPb = tx.ElemsPb[:len(tx.Elems)]
	}
	for j, srcELem := range tx.Elems {
		elem := tx.ElemsPb[j]
		if elem == nil {
			elem = &AttrElemPb{}
			tx.ElemsPb[j] = elem
		}
		elem.SI = srcELem.SI
		elem.AttrID = uint64(srcELem.AttrID)
		if err := srcELem.Val.MarshalToBuf(&elem.ValBuf); err != nil {
			return err
		}
	}
	return nil
}


func (tx *CellTx) MarshalToPb(dst *CellTxPb) error {
	tx.MarshalAttrs()
	dst.Op = tx.Op
	dst.CellSpec = tx.CellSpec
	dst.TargetCell = int64(tx.TargetCell)
	dst.Elems = tx.ElemsPb
	return nil
}
*/

// Form
// If reqID == 0, then this sends an attr to the client's session controller (vs a specific request)
func SendClientMetaAttr(sess HostSession, reqID uint64, val AttrElemVal, status ReqStatus) error {
	metaOp := TxOp{
		OpCode: TxOpCode_MetaAttr,
		AttrElem: AttrElem{
			AttrID: FormAttrID(val.ElemTypeName()),
		},
	}
	tx := NewTxMsg()
	tx.ReqID = reqID
	tx.Status = status
	err := tx.MarshalOp(&metaOp, val)
	if err != nil {
		return err
	}
	return sess.SendTx(tx)
}

/*
func (tx *TxMsg) UpsertAttr(target CellID, attrSpec AttrSpec, seriesIndex UID, val AttrElemVal) error {
	op := TxOp{
		OpCode:     TxOpCode_UpsertAttr,
		AttrElem:   AttrElem{},
		TargetCell: target,
	}
	return tx.MarshalOp(&op, val)
}
*/

// On success:
//   - TxOp.DataStoreOfs and TxOp.DataLen are overwritten,
//   - the TxMsg.DataStore is appended to,
//   - and the TxOp is appended to TxMsg.Ops.
func (tx *TxMsg) MarshalOp(op *TxOp, val AttrElemVal) error {
	var err error
	startOfs := int64(len(tx.DataStore))
	tx.DataStore, err = val.MarshalToStore(tx.DataStore)
	if err != nil {
		return err
	}
	op.DataStoreOfs = startOfs
	op.DataLen = int64(len(tx.DataStore)) - startOfs

	tx.Ops = append(tx.Ops, *op)
	return nil
}

func (tx *TxMsg) UnmarshalAttrElem(opIndex int, out AttrElemVal) error {
	if opIndex < 0 || opIndex >= len(tx.Ops) {
		panic("opIndex out of range")
	}
	op := &tx.Ops[opIndex]
	return out.Unmarshal(tx.DataStore[op.DataStoreOfs : op.DataStoreOfs+op.DataLen])
}

// func (tx *TxMsg) IsMetaAttrElemType() (attr *AttrElem, err error) {

// }

func (tx *TxMsg) GetMetaAttr() (attr AttrElem, err error) {
	if len(tx.Ops) == 0 || tx.Ops[0].OpCode != TxOpCode_MetaAttr {
		return AttrElem{}, ErrCode_MalformedTx.Error("expected meta attr")
	}
	return tx.Ops[0].AttrElem, nil
	// val AttrElemVal

	// return out.Unmarshal(tx.DataStore[op.DataStoreOfs:op.DataStoreOfs+op.DataLen])

	// return msg.Ops[0].Elems[0], nil
}

/*
func (tx *TxMsg) UnmarshalFrom(msg *TxMsg, reg SessionRegistry, native bool) error {
	tx.ReqID = msg.ReqID
	tx.Status = msg.Status
	tx.Ops = tx.Ops[:0]

	elemCount := 0

	srcTxs := msg.Ops
	if cap(tx.Ops) < len(srcTxs) {
		tx.Ops = make([]TxOp, len(srcTxs))
	} else {
		tx.Ops = tx.Ops[:len(srcTxs)]
	}
	for i, cellTx := range srcTxs {
		elems := make([]AttrElemVal, len(cellTx.Elems))
		for j, srcElem := range cellTx.Elems {
			attrID := uint32(srcElem.AttrID)
			elem := AttrElem{
				SeriesIndex:     srcElem.SeriesIndex,
				AttrID: attrID,
			}
			var err error
			elem, err = reg.NewAttrElem(attrID, native)
			if err == nil {
				err = elem.Val.Unmarshal(srcElem.ValBuf)
			}
			if err != nil {
				return err
			}
			elems[j] = elem
			elemCount++
		}

		tx.Ops[i] = TxOp{
			OpCode: cellTx.Op,
			//Elems:      elems,
		}
		tx.Ops[i].TargetCell.AssignFromU64(cellTx.TargetCell[0], cellTx.TargetCell[1])
	}

	if elemCount == 0 {
		return ErrBadCellTx
	}
	return nil
}


// Pushes a attr mutation to the client, returning true if the msg was sent (false if the client has been closed).
func (bat *CellTx) PushBatch(ctx PinContext) error {

	for _, attr := range bat.Attrs {
		msg, err := attr.MarshalToMsg(bat.Target)
		if err != nil {
			ctx.Warnf("MarshalToMsg() err: %v", err)
			continue
		}

		// if i == len(bat.Attrs)-1 {
		// 	msg.Flags |= MsgFlags_CellCheckpoint
		// }

		if !ctx.PushTx(msg) {
			return ErrPinCtxClosed
		}
	}

	return nil

}
*/

func (tx *TxMsg) Write(r io.Writer) error {
	panic("TODO")
}

func ReadTxMsg(r io.Reader) (*TxMsg, error) {

	readBytes := func(dst []byte) error {
		for L := 0; L < len(dst); {
			n, err := r.Read(dst[L:])
			if err != nil {
				return err
			}
			L += n
		}
		return nil
	}

	var hdr TxHeader
	if err := readBytes(hdr[:]); err != nil {
		return nil, err
	}

	tx, err := hdr.NewTxMsg()
	if err != nil {
		return nil, err
	}
	bodyLen := hdr.TxBodyLen()
	dataLen := hdr.TxDataLen()

	// Use the DataStore buf as a scratch buffer to load the body for processing.
	needSz := max(bodyLen, dataLen)
	if cap(tx.DataStore) < needSz {
		tx.DataStore = make([]byte, max(needSz, 4096))
	}

	// Read TxMsg fields & Ops
	{
		buf := tx.DataStore[:bodyLen]
		if err = readBytes(buf); err != nil {
			return nil, err
		}
		if err = tx.UnmarshalBody(buf); err != nil {
			return nil, err
		}
	}

	// Read TxMsg attr data store (only unserialized on-demand)
	tx.DataStore = tx.DataStore[:dataLen]
	if err = readBytes(tx.DataStore); err != nil {
		return nil, err
	}

	return tx, nil
}

func (tx *TxMsg) MarshalHeaderAndBody(dst *[]byte) {
	body := (*dst)[:0]
	if cap(body) < int(Const_TxHeader_Size) {
		body = make([]byte, 1024)
	}
	body = body[:Const_TxHeader_Size]
	body = tx.MarshalBody(body)

	hdr := tx.FormTxHeader(len(body))
	copy(body[0:], hdr[:])
	*dst = body
}

func (tx *TxMsg) MarshalToBytes(buf *[]byte) {
	tx.MarshalHeaderAndBody(buf)
	*buf = append(*buf, tx.DataStore...)
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

func (tx *TxMsg) MarshalBody(dst []byte) []byte {
	dst = append(dst, byte(tx.Status))
	dst = binary.BigEndian.AppendUint64(dst, tx.ReqID)

	flags := TxOpFlags(0)
	prev := TxOp{}
	dst = binary.AppendUvarint(dst, 0) // reserved
	dst = binary.AppendUvarint(dst, uint64(len(tx.Ops)))

	for _, op := range tx.Ops {
		dst = binary.AppendUvarint(dst, uint64(op.OpCode))
		if op.TargetCell == prev.TargetCell {
			flags |= TxOpFlags_TargetCell_Repeat
		}
		if op.ParentCell == prev.ParentCell {
			flags |= TxOpFlags_ParentCell_Repeat
		}
		if op.AttrID == prev.AttrID {
			flags |= TxOpFlags_Attr_Repeat
		}
		if op.SeriesIndex == prev.SeriesIndex {
			flags |= TxOpFlags_SI_Repeat
		}
		dst = binary.AppendUvarint(dst, uint64(flags))

		if flags&TxOpFlags_TargetCell_Repeat == 0 {
			dst = binary.BigEndian.AppendUint64(dst, op.TargetCell[0])
			dst = binary.BigEndian.AppendUint64(dst, op.TargetCell[1])
		}
		if flags&TxOpFlags_ParentCell_Repeat == 0 {
			dst = binary.BigEndian.AppendUint64(dst, op.ParentCell[0])
			dst = binary.BigEndian.AppendUint64(dst, op.ParentCell[1])
		}
		if flags&TxOpFlags_Attr_Repeat == 0 {
			dst = binary.BigEndian.AppendUint64(dst, op.AttrID[0])
			dst = binary.BigEndian.AppendUint64(dst, op.AttrID[1])
		}
		if flags&TxOpFlags_SI_Repeat == 0 {
			dst = binary.BigEndian.AppendUint64(dst, op.SeriesIndex[0])
			dst = binary.BigEndian.AppendUint64(dst, op.SeriesIndex[1])
		}

		dst = binary.AppendVarint(dst, op.DataStoreOfs)
		dst = binary.AppendVarint(dst, op.DataLen)
		prev = op
	}

	return dst
}

var ErrMalformedTx = ErrCode_MalformedTx.Error("bad varint")

const MaxTxOpEncodingSz = 64

func (tx *TxMsg) UnmarshalBody(src []byte) error {
	//var buf [MaxTxOpEncodingSz]byte
	// N, err := src.Read(buf[:])
	// if err != io.EOF {
	// 	return err
	// }
	_, n := binary.Uvarint(src[0:])
	if n <= 0 {
		return ErrMalformedTx
	}
	pos := n
	var numOps uint64
	if numOps, n = binary.Uvarint(src[pos:]); n <= 0 {
		return ErrMalformedTx
	}
	pos += n

	var op TxOp

	for i := uint64(0); i < numOps; i++ {
		// copy(src[:], src[pos:n-pos])
		// n -= pos
		// pos = 0

		// n, err = src.Read(src[n:])
		// if err != io.EOF {
		// 	return err
		// }

		var opCode uint64
		if opCode, n = binary.Uvarint(src[pos:]); n <= 0 {
			return ErrCode_MalformedTx.Error("unexpected TxOps count")
		}
		pos += n
		op.OpCode = TxOpCode(opCode)

		var fl64 uint64
		if fl64, n = binary.Uvarint(src[pos:]); n <= 0 {
			return ErrCode_MalformedTx.Error("unexpected TxOps count")
		}
		pos += n
		flags := TxOpFlags(fl64)
		if flags&TxOpFlags_TargetCell_Repeat == 0 {
			if op.TargetCell[0], n = binary.Uvarint(src[pos:]); n <= 0 {
				return ErrMalformedTx
			}
			pos += n
			if op.TargetCell[1], n = binary.Uvarint(src[pos:]); n <= 0 {
				return ErrMalformedTx
			}
			pos += n
		}
		if flags&TxOpFlags_ParentCell_Repeat == 0 {
			if op.ParentCell[0], n = binary.Uvarint(src[pos:]); n <= 0 {
				return ErrMalformedTx
			}
			pos += n
			if op.ParentCell[1], n = binary.Uvarint(src[pos:]); n <= 0 {
				return ErrMalformedTx
			}
			pos += n
		}
		if flags&TxOpFlags_Attr_Repeat == 0 {
			if op.AttrID[0], n = binary.Uvarint(src[pos:]); n <= 0 {
				return ErrMalformedTx
			}
			pos += n
			if op.AttrID[1], n = binary.Uvarint(src[pos:]); n <= 0 {
				return ErrMalformedTx
			}
			pos += n
		}
		if flags&TxOpFlags_SI_Repeat == 0 {
			if op.SeriesIndex[0], n = binary.Uvarint(src[pos:]); n <= 0 {
				return ErrMalformedTx
			}
			pos += n
			if op.SeriesIndex[1], n = binary.Uvarint(src[pos:]); n <= 0 {
				return ErrMalformedTx
			}
			pos += n
		}

		if op.DataStoreOfs, n = binary.Varint(src[pos:]); n <= 0 {
			return ErrMalformedTx
		}
		pos += n

		if op.DataLen, n = binary.Varint(src[pos:]); n <= 0 {
			return ErrMalformedTx
		}
		pos += n

		tx.Ops = append(tx.Ops, op)
	}

	return nil
}

/*

func (v *TxMsgPb) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *TxMsgPb) TypeName() string {
	return "TxMsg"
}

func (v *TxMsgPb) New() AttrElemVal {
	return &TxMsgPb{}
}

*/

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
