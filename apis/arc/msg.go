package arc

import (
	"sync"
)




func NewTxMsg() *TxMsg {
	msg := gTxMsgPool.Get().(*TxMsg)
	return msg
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
		CellOps: tx.CellOps[:0],
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

var gTxMsgPool = sync.Pool{
	New: func() interface{} {
		return &TxMsg{}
	},
}




// func (tx *AttrOp) Clear(op AttrOpCode) {
// 	tx.Op = op
// 	tx.CellID = CellID{}
// 	tx.Elems = tx.Elems[:0]
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

// If reqID == 0, then this sends an attr to the client's session controller (vs a specific request)
func SendClientMetaAttr(sess HostSession, reqID uint64, val AttrElemVal) error {
	msg, err := FormClientMetaAttrMsg(sess, val.ElemTypeName(), val)
	msg.ReqID = reqID
	if err != nil {
		return err
	}
	return sess.SendTx(msg)
}

func FormClientMetaAttrMsg(reg SessionRegistry, attrSpec string, val AttrElemVal) (*TxMsg, error) {
	spec, err := reg.ResolveAttrSpec(attrSpec, false)
	if err != nil {
		return nil, err
	}

	return FormMetaAttrTx(spec, val)
}

func FormMetaAttrTx(attrSpec AttrSpec, val AttrElemVal) (*TxMsg, error) {
	elemPb := &AttrElemPb{
		AttrID: uint64(attrSpec.DefID),
	}
	if err := val.MarshalToBuf(&elemPb.ValBuf); err != nil {
		return nil, err
	}

	tx := &CellTxPb{
		Op: CellTxOp_MetaAttr,
		Elems: []*AttrElemPb{
			elemPb,
		},
	}

	msg := NewMsg()
	msg.ReqID = 0 // signals a meta message
	msg.Status = ReqStatus_Synced
	msg.CellTxs = append(msg.CellTxs, tx)
	return msg, nil
}

func (msg *TxMsg) GetMetaAttr() (attr *AttrElemPb, err error) {
	if len(msg.CellTxs) == 0 || msg.CellTxs[0].Op != CellTxOp_MetaAttr || msg.CellTxs[0].Elems == nil || len(msg.CellTxs[0].Elems) == 0 {
		return nil, ErrCode_MalformedTx.Error("missing meta attr")
	}

	return msg.CellTxs[0].Elems[0], nil
}

func (tx *TxMsg) UnmarshalFrom(msg *TxMsg, reg SessionRegistry, native bool) error {
	tx.ReqID = msg.ReqID
	tx.Status = msg.Status
	tx.CellTxs = tx.CellTxs[:0]

	elemCount := 0

	srcTxs := msg.CellTxs
	if cap(tx.CellTxs) < len(srcTxs) {
		tx.CellTxs = make([]CellTx, len(srcTxs))
	} else {
		tx.CellTxs = tx.CellTxs[:len(srcTxs)]
	}
	for i, cellTx := range srcTxs {
		elems := make([]AttrElemVal, len(cellTx.Elems))
		for j, srcElem := range cellTx.Elems {
			attrID := uint32(srcElem.AttrID)
			elem := AttrElemVal{
				SI:     srcElem.SI,
				AttrID: attrID,
			}
			var err error
			elem.Val, err = reg.NewAttrElem(attrID, native)
			if err == nil {
				err = elem.Val.Unmarshal(srcElem.ValBuf)
			}
			if err != nil {
				return err
			}
			elems[j] = elem
			elemCount++
		}

		tx.CellTxs[i] = CellTx{
			Op:         cellTx.Op,
			//Elems:      elems,
		}
		tx.CellTxs[i].TargetCell.AssignFromU64(cellTx.CellID_0, cellTx.CellID_1)
	}

	if elemCount == 0 {
		return ErrBadCellTx
	}
	return nil
}

/*
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

func (tx *TxMsg) MarshalToBuf(dst *[]byte) error {
	pb := TxMsgPb{
		ReqID:   tx.ReqID,
		CellTxs: make([]*CellTxPb, len(tx.CellTxs)),
	}
	for i, srcTx := range tx.CellTxs {
		cellTx := &CellTxPb{
			Op:           srcTx.Op,
			CellSpec:     srcTx.CellSpec,
			TargetCellID: int64(srcTx.Target),
			Elems:        make([]*AttrElemPb, len(srcTx.Elems)),
		}
		for j, AttrElemVal := range srcTx.Elems {
			//AttrElemVal.ValBuf = make([]byte, AttrElemVal.Val.Marhal
			elem := &AttrElemPb{
				SI:     AttrElemVal.SI,
				AttrID: AttrElemVal.AttrID,
			}
			AttrElemVal.Val.MarshalToBuf(&elem.ValBuf)
			cellTx.Elems[j] = elem
		}
		pb.CellTxs[i] = cellTx
	}
	sz := pb.Size()
	if cap(*dst) < sz {
		*dst = make([]byte, sz)
	} else {
		*dst = (*dst)[:sz]
	}
	_, err := pb.MarshalToSizedBuffer(*dst)
	return err
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
