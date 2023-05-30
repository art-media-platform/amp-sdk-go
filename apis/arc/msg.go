package arc

import (
	"reflect"
	"sync"
	"time"

	"github.com/arcspace/go-arc-sdk/stdlib/symbol"
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

func (msg *Msg) SetValInt(valType ValType, valInt int64) {
	msg.ValType = valType
	msg.ValInt = valInt
	msg.ValBuf = msg.ValBuf[:0]
}

// Sets ValType to the given type and resizes ValBuf to the given size in preparation to receive value data.
func (msg *Msg) SetupValBuf(valType ValType, sz int) {
	msg.ValInt = int64(sz)
	msg.ValType = valType
	if sz > cap(msg.ValBuf) {
		msg.ValBuf = make([]byte, sz, (sz+0x3FF)&^0x3FF)
	} else {
		msg.ValBuf = msg.ValBuf[:sz]
	}
}

// Sets ValType to the given type and copies value data into ValBuf.
func (msg *Msg) SetValBuf(valType ValType, valSrc []byte) {
	msg.SetupValBuf(valType, len(valSrc))
	copy(msg.ValBuf, valSrc)
}

func (msg *Msg) SetVal(val interface{}) {
	var err error

	switch v := val.(type) {

	case string:
		msg.SetValBuf(ValType_bytes, []byte(v))

	case *[]byte:
		msg.SetValBuf(ValType_bytes, *v)

	case []byte:
		msg.SetValBuf(ValType_bytes, v)

	case time.Time:
		msg.SetValInt(ValType_DateTime, int64(ConvertToTimeFS(v)))

	case *AssetRef:
		msg.SetupValBuf(ValType_AssetRef, v.Size())
		_, err = v.MarshalToSizedBuffer(msg.ValBuf)

	case TimeFS:
		msg.SetValInt(ValType_DateTime, int64(v))

	case *Defs:
		msg.SetupValBuf(ValType_Defs, v.Size())
		_, err = v.MarshalToSizedBuffer(msg.ValBuf)

		// case *CellInfo:
		// 	msg.SetupValBuf(uint64(ValType_CellInfo), v.Size())
		//     _, err = v.MarshalToSizedBuffer(msg.ValBuf)

	case *Err:
		msg.SetupValBuf(ValType_Err, v.Size())
		_, err = v.MarshalToSizedBuffer(msg.ValBuf)

	case error:
		plErr, _ := v.(*Err)
		if plErr == nil {
			err := ErrCode_UnnamedErr.Wrap(v)
			plErr = err.(*Err)
		}
		msg.SetupValBuf(ValType_Err, plErr.Size())
		_, err = plErr.MarshalToSizedBuffer(msg.ValBuf)

	case nil:
		msg.SetupValBuf(ValType_nil, 0)
	}

	if err != nil {
		panic(err)
	}
}

func (msg *Msg) LoadVal(dst interface{}) error {
	if msg == nil {
		loadNil(dst)
		return ErrCode_BadValue.Error("got nil Msg")
	}

	ok := false

	switch msg.ValType {

	case ValType_int:
		ok = true
		switch v := dst.(type) {
		case *int64:
			*v = int64(msg.ValInt)
		case *uint64:
			*v = uint64(msg.ValInt)
		default:
			ok = false
		}

	case ValType_string, ValType_bytes:
		ok = true
		switch v := dst.(type) {
		case *string:
			*v = string(msg.ValBuf)
		case *[]byte:
			*v = append(*v, msg.ValBuf...)
		default:
			ok = false
		}

	case ValType_PinReq:
		if v, match := dst.(*PinReq); match {
			tmp := PinReq{}
			if tmp.Unmarshal(msg.ValBuf) == nil {
				*v = tmp
				ok = true
			}
		}

	case ValType_Defs:
		if v, match := dst.(*Defs); match {
			tmp := Defs{}
			if tmp.Unmarshal(msg.ValBuf) == nil {
				*v = tmp
				ok = true
			}
		}

	case ValType_HandleURI:
		if v, match := dst.(*HandleURI); match {
			tmp := HandleURI{}
			if tmp.Unmarshal(msg.ValBuf) == nil {
				*v = tmp
				ok = true
			}
		}

	case ValType_LoginReq:
		if v, match := dst.(*LoginReq); match {
			tmp := LoginReq{}
			if tmp.Unmarshal(msg.ValBuf) == nil {
				*v = tmp
				ok = true
			}
		}

	}

	if !ok {
		return ErrCode_BadValue.Errorf("expected %v from Msg", reflect.TypeOf(dst))
	}

	return nil
}

func loadNil(dst interface{}) {
	switch v := dst.(type) {
	// case *Content:
	//     v.ContentType = v.ContentType[:0]
	//     v.DataLen = 0
	//     v.Data = v.Data[:0]
	case *string:
		*v = ""
	case *symbol.ID:
		*v = 0
	case *TIDBuf:
		*v = TIDBuf{}
	case *TID:
		*v = nil
	case *PinReq:
	    *v = PinReq{}
	case *Defs:
		*v = Defs{}
	case *int:
		*v = 0
	case *int64:
		*v = 0
	case *uint64:
		*v = 0
	case *float64:
		*v = 0
	case *[]byte:
		*v = nil
	default:
		panic("unexpected dst type")
	}
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

func NewMsgWithValue(value interface{}) *Msg {
	msg := NewMsg()
	msg.SetVal(value)
	return msg
}
