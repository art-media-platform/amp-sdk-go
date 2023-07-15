package arc

import (
	"sync"
)

// MsgBatch is an ordered list os Msgs
// See NewMsgBatch()
type MsgBatch struct {
	Msgs []*Msg
}

// Sets a reasonable size beyond which buffers should be shared rather than copied.
const MsgValBufCopyLimit = 16 * 1024

func NewMsgBatch() *MsgBatch {
	return gMsgBatchPool.Get().(*MsgBatch)
}

func (batch *MsgBatch) Reset(count int) []*Msg {
	if count > cap(batch.Msgs) {
		msgs := make([]*Msg, count)
		copy(msgs, batch.Msgs)
		batch.Msgs = msgs
	} else {
		batch.Msgs = batch.Msgs[:count]
	}

	// Alloc or init  each msg
	for i, msg := range batch.Msgs {
		if msg == nil {
			batch.Msgs[i] = NewMsg()
		} else {
			msg.Init()
		}
	}

	return batch.Msgs
}

func (batch *MsgBatch) AddNew(count int) []*Msg {
	N := len(batch.Msgs)
	for i := 0; i < count; i++ {
		batch.Msgs = append(batch.Msgs, NewMsg())
	}
	return batch.Msgs[N:]
}

func (batch *MsgBatch) AddMsgs(msgs []*Msg) {
	batch.Msgs = append(batch.Msgs, msgs...)
}

func (batch *MsgBatch) AddMsg() *Msg {
	m := NewMsg()
	batch.Msgs = append(batch.Msgs, m)
	return m
}

func (batch *MsgBatch) Reclaim() {
	for i, msg := range batch.Msgs {
		msg.Reclaim()
		batch.Msgs[i] = nil
	}
	batch.Msgs = batch.Msgs[:0]
	gMsgBatchPool.Put(batch)
}

func (batch *MsgBatch) PushCopyToClient(dst PinContext) bool {
	for _, src := range batch.Msgs {
		msg := CopyMsg(src)
		if !dst.PushMsg(msg) {
			return false
		}
	}
	return true
}

func NewMsg() *Msg {
	msg := gMsgPool.Get().(*Msg)
	if msg.Flags&MsgFlags_ValBufShared != 0 {
		panic("Msg discarded as shared")
	}
	return msg
}

func CopyMsg(src *Msg) *Msg {
	msg := NewMsg()

	if src != nil {
		// If the src buffer is big share it instead of copy it
		if len(src.ValBuf) > MsgValBufCopyLimit {
			*msg = *src
			msg.Flags |= MsgFlags_ValBufShared
			src.Flags |= MsgFlags_ValBufShared
		} else {
			valBuf := append(msg.ValBuf[:0], src.ValBuf...)
			*msg = *src
			msg.Flags &^= MsgFlags_ValBufShared
			msg.ValBuf = valBuf
		}
	}
	return msg
}

func (msg *Msg) Init() {
	if msg.Flags&MsgFlags_ValBufShared != 0 {
		*msg = Msg{}
	} else {
		valBuf := msg.ValBuf[:0]
		*msg = Msg{
			ValBuf: valBuf,
		}
	}
}

func (msg *Msg) Reclaim() {
	if msg != nil {
		msg.Init()
		gMsgPool.Put(msg)
	}
}

func (msg *Msg) MarshalAttrElem(attrID uint32, src PbValue) error {
	msg.AttrID = attrID
	sz := src.Size()
	if sz > cap(msg.ValBuf) {
		msg.ValBuf = make([]byte, sz, (sz+0x3FF)&^0x3FF)
	} else {
		msg.ValBuf = msg.ValBuf[:sz]
	}
	_, err := src.MarshalToSizedBuffer(msg.ValBuf)
	return err
}

func (msg *Msg) UnmarshalValue(dst PbValue) error {
	return dst.Unmarshal(msg.ValBuf)
}

func (msg *Msg) UnmarshalAttrElem(reg SessionRegistry) (elem AttrElem, err error) {
	elem.Val, err = reg.NewAttrElem(msg.AttrID, false)
	if err != nil {
		return
	}
	err = elem.Val.Unmarshal(msg.ValBuf)
	if err != nil {
		return
	}
	elem.AttrID = msg.AttrID
	elem.SI = msg.SI
	return
}

func (attr AttrElem) MarshalToMsg() (*Msg, error) {
	msg := NewMsg()
	msg.Op = MsgOp_PushAttrElem
	msg.AttrID = attr.AttrID
	msg.SI = attr.SI
	err := attr.Val.MarshalToBuf(&msg.ValBuf)
	return msg, err
}

var gMsgPool = sync.Pool{
	New: func() interface{} {
		return &Msg{}
	},
}

var gMsgBatchPool = sync.Pool{
	New: func() interface{} {
		return &MsgBatch{
			Msgs: make([]*Msg, 0, 16),
		}
	},
}

func (bat *AttrBatch) Clear(target CellID) {
	bat.Target = target
	if cap(bat.Attrs) > 0 {
		bat.Attrs = bat.Attrs[:0]
	} else {
		bat.Attrs = make([]AttrElem, 0, 4)
	}
}

func (bat *AttrBatch) Add(attrID uint32, val ElemVal) {
	if val == nil {
		return
	}
	if attrID == 0 {
		panic("attrID == 0")
	}
	bat.Attrs = append(bat.Attrs, AttrElem{
		Val:    val,
		AttrID: attrID,
	})
}

// Pushes a attr mutation to the client, returning true if the msg was sent (false if the client has been closed).
func (bat AttrBatch) PushBatch(ctx PinContext) error {

	for _, attr := range bat.Attrs {
		msg, err := attr.MarshalToMsg()
		if err != nil {
			ctx.Warnf("MarshalToMsg() err: %v", err)
			continue
		}
		msg.CellID = int64(bat.Target)

		// if i == len(bat.Attrs)-1 {
		// 	msg.Flags |= MsgFlags_CellCheckpoint
		// }

		if !ctx.PushMsg(msg) {
			return ErrPinCtxClosed
		}
	}

	return nil

}
