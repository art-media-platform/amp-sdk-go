package arc

import (
	"encoding/binary"
	"sync"
)

// TxDataStore is a message packet sent to / from a client.
// It is leads with a fixed-size header (TxHeader_Size) followed by a variable-size body.
type TxDataStore []byte

func (tx TxDataStore) GetTxTotalLen() int {
	if tx[TxHeader_OpOfs] != byte(TxHeader_OpRecvTx) {
		return 0
	}
	bodySz := int(binary.BigEndian.Uint32(tx[3:7]))
	return bodySz
}

func (tx TxDataStore) SetTxBodyLen(bodyLen int) {
	txLen := bodyLen + int(TxHeader_Size)
	binary.BigEndian.PutUint32(tx[3:7], uint32(txLen))
	tx[TxHeader_OpOfs] = byte(TxHeader_OpRecvTx)
}

func (tx TxDataStore) InitHeader(bodyLen int) {
	tx[0] = 0
	tx[1] = 0
	tx[2] = 0
	txLen := bodyLen + int(TxHeader_Size)
	binary.BigEndian.PutUint32(tx[3:7], uint32(txLen))
	tx[TxHeader_OpOfs] = byte(TxHeader_OpRecvTx)
}

///////////////////////// host -> client /////////////////////////

func (msg *TxMsg) MarshalToTxBuffer(txBuf []byte) error {
	n, err := msg.MarshalToSizedBuffer(txBuf[TxHeader_Size:])
	if err != nil {
		return err
	}
	TxDataStore(txBuf).InitHeader(n)
	return nil
}



func NewTxMsg() *TxMsg {
	msg := gTxMsgPool.Get().(*TxMsg)
	return msg
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

// func (op *CellOp) MarshalToStore(tx *TxMsg, val AttrElemVal) error {
// 	var err error
	
// 	op.DataStoreOfs = int64(len(tx.DataStore))
// 	tx.DataStore, err = val.MarshalToStore(tx.DataStore)
// 	if err != nil {
// 		return err
// 	}
// 	op.DataLen = int64(len(tx.DataStore))
// }


func (op *CellOp) HasAttrUID() bool {
	return op.AttrID[0] != 0 || op.AttrID[1] != 0 
}

func (op *CellOp) NilAttrUID() bool {
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
	metaOp := CellOp{
		OpCode: CellOpCode_MetaAttr,
	}
	tx := NewTxMsg()
	tx.ReqID = reqID
	tx.Status = status
	err := tx.MarshalCellOp(&metaOp, val)
	if err != nil {
		return err
	}
	return sess.SendTx(tx)
}

func (tx *TxMsg) MarshalUpsert(attrSpec AttrSpec, seriesIndex UID, val AttrElemVal) error {
	op := CellOp{
		OpCode:   CellOpCode_UpsertAttr,
		AttrElem: AttrElem{
			
		},
	}
	return tx.MarshalCellOp(&op, val)
}

func (tx *TxMsg) MarshalCellOp(op *CellOp, val AttrElemVal) error {
	op.DataStoreOfs = int64(len(tx.AttrStore))
	var err error
	tx.AttrStore, err = val.MarshalToStore(tx.AttrStore)
	if err != nil {
		return err
	}
	op.DataLen = int64(len(tx.AttrStore)) - op.DataStoreOfs
	
	tx.Ops = append(tx.Ops, *op)
	return nil
}


func (tx *TxMsg) UnmarshalAttrElem(opIndex int, out AttrElemVal) error {
	if opIndex < 0 || opIndex >= len(tx.Ops) {
		panic("opIndex out of range")
	}
	op := &tx.Ops[opIndex]
	return out.Unmarshal(tx.AttrStore[op.DataStoreOfs:op.DataStoreOfs+op.DataLen])
}



// func (tx *TxMsg) IsMetaAttrElemType() (attr *AttrElem, err error) {

// }

func (tx *TxMsg) GetMetaAttr() (attr AttrElem, err error) {
	if len(tx.Ops) == 0 || tx.Ops[0].OpCode != CellOpCode_MetaAttr  {
		return AttrElem{}, ErrCode_MalformedTx.Error("expected meta attr")
	}
	return tx.Ops[0].AttrElem, nil
	// val AttrElemVal
	
	// return out.Unmarshal(tx.AttrStore[op.DataStoreOfs:op.DataStoreOfs+op.DataLen])

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
		tx.Ops = make([]CellOp, len(srcTxs))
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

		tx.Ops[i] = CellOp{
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

func (tx *TxMsg) MarshalTo(dst []byte) []byte {
	flags := CellOpFlags(0)
	
	targetCell := CellID{}
	parentCell := CellID{}
	attrID := AttrUID{}
	SI := SeriesIndex{}

	dst = binary.AppendUvarint(dst, 0) // reserved
	dataStoreOfs := int64(len(dst))
	dst = append(dst, tx.AttrStore...)
	dst = binary.AppendUvarint(dst, 0) // reserved

	dst = binary.AppendUvarint(dst, uint64(len(tx.Ops)))
	for _, op := range tx.Ops {
		dst = binary.AppendUvarint(dst, uint64(op.OpCode))
		if op.TargetCell == targetCell {
			flags |= CellOpFlags_TargetCell_Repeat
		}
		if op.ParentCell == parentCell {
			flags |= CellOpFlags_ParentCell_Repeat
		}
		if op.AttrID == attrID {
			flags |= CellOpFlags_Attr_Repeat
		}
		if op.SeriesIndex == SI {
			flags |= CellOpFlags_SI_Repeat
		}
		dst = append(dst, byte(flags))
		
		if flags & CellOpFlags_TargetCell_Repeat == 0 {
			dst = binary.BigEndian.AppendUint64(dst, op.TargetCell[0])
			dst = binary.BigEndian.AppendUint64(dst, op.TargetCell[1])
		}
		if flags & CellOpFlags_ParentCell_Repeat == 0 {
			dst = binary.BigEndian.AppendUint64(dst, op.ParentCell[0])
			dst = binary.BigEndian.AppendUint64(dst, op.ParentCell[1])
		}
		if flags & CellOpFlags_Attr_Repeat == 0 {
			dst = binary.BigEndian.AppendUint64(dst, op.AttrID[0])
			dst = binary.BigEndian.AppendUint64(dst, op.AttrID[1])
		}	
		if flags & CellOpFlags_SI_Repeat == 0 {
			dst = binary.BigEndian.AppendUint64(dst, op.SeriesIndex[0])
			dst = binary.BigEndian.AppendUint64(dst, op.SeriesIndex[1])
		}
		
		dst = binary.AppendVarint(dst, op.DataStoreOfs + dataStoreOfs)
		dst = binary.AppendVarint(dst, op.DataLen)
	}

	return dst
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
