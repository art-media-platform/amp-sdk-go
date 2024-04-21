package amp

import (
	"bytes"
	"fmt"
	io "io"
	"testing"
)

func TestExpr(t *testing.T) {
	var tsts = []string{
		"elem-type.org",
		"[UTC16]elem",
		"elem:name",
		"elem-type.org:name",
		"[Surface.Name]elem:name",
		"[Locale.Name]elem-type:name.ext",
	}

	for _, tst := range tsts {
		expr, err := ParseAttrDef(tst)
		if err != nil {
			fmt.Printf("%-30s %v\n", tst, err)
		} else {
			fmt.Printf("%-30s %-15v %-15v %-15v\n", tst, expr.SeriesSpec, expr.ElemType, expr.AttrName)
		}
	}

	uid := StringToUID("hello")
	uuid := uid.ToUUID()
	uuidStr := uuid.String()
	if uuidStr != "5d41402a-bc4b-2a76-b971-9d911017c592" {
		t.Errorf("StringToUID failed")
	}
	parsedUID, err := ParseUUID(uuidStr)
	if err != nil || parsedUID != uid {
		t.Errorf("ParseUUID failed")
	}

}

func TestTxSerialize(t *testing.T) {
	// Test serialization of a simple TxMsg

	tx := NewTxMsg(true)
	tx.Status = ReqStatus_Syncing
	tx.RouteTo_0 = 888854513
	tx.RouteTo_1 = 7777435
	{
		op := TxOp{
			OpCode:   TxOpCode_MetaAttr,
			ParentID: CellID{1, 2, 3},
			TargetID: CellID{4, 555, 666},
			AttrID:   AttrID{32232, 32334},
			SI:       SeriesIndex{7383, 76549},
		}
		tx.MarshalOpValue(&op, &Login{
			UserUID:  "alan1",
			HostAddr: "batwing ave",
		})
		tx.DataStore = append(tx.DataStore, []byte("bytes not used but stored -- not normal!")...)

		op.SI[1] = 50454123
		op.ParentID[2] = 40411236
		data := []byte("hello.world-")
		for i := 0; i < 7; i++ {
			data = append(data, data...)
		}
		tx.MarshalOpValue(&op, &Login{
			UserUID:  "cmdr6",
			HostAddr: string(data),
		})

		op.SI[0] = 111111
		op.OpCode = TxOpCode_RemoveAttr
		tx.MarshalOpWithBuf(&op, nil)
	}

	var txBuf []byte
	tx.MarshalToBuffer(&txBuf)

	r := bufReader{
		buf: txBuf,
	}
	tx2, err := ReadTxMsg(&r)
	if err != nil {
		t.Errorf("ReadTxMsg failed: %v", err)
	}
	if tx2.TxInfo != tx.TxInfo {
		t.Errorf("ReadTxMsg failed: TxInfo mismatch")
	}
	if len(tx2.Ops) != len(tx.Ops) {
		t.Errorf("ReadTxMsg failed: TxInfo mismatch")
	}
	if !bytes.Equal(tx.DataStore, tx2.DataStore) {
		t.Errorf("ReadTxMsg failed: DataStore mismatch")
	}
	for i, op1 := range tx.Ops {
		op2 := tx2.Ops[i]

		if op1.OpCode != op2.OpCode || op1.TargetID != op2.TargetID || op1.ParentID != op2.ParentID || op1.AttrID != op2.AttrID || op1.SI != op2.SI || op1.DataStoreOfs != op2.DataStoreOfs || op1.DataLen != op2.DataLen {
			t.Errorf("ReadTxMsg failed: Op mismatch")
		}
	}
}

type bufReader struct {
	buf []byte
	pos int
}

func (r *bufReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.buf) {
		return 0, io.EOF
	}
	n = copy(p, r.buf[r.pos:])
	r.pos += n
	return n, nil
}

func TestNewTimeID(t *testing.T) {
	var prevIDs [64]TimeID

	prevIDs[0] = TimeID{100, (^uint64(0)) - 500}

	delta := TimeID{100, 100}
	for i := 1; i < 64; i++ {
		prevIDs[i] = prevIDs[i-1].Add(delta)
	}
	for i := 1; i < 64; i++ {
		prev := prevIDs[i-1]
		curr := prevIDs[i]
		if prev.CompareTo(curr) >= 0 {
			t.Errorf("TimeID.Add() returned a non-increasing value: %v <= %v", prev, curr)
		}
		if curr.Sub(prev) != delta {
			t.Errorf("TimeID.Diff() returned a wrong value: %v != %v", curr.Sub(prev), delta)
		}
	}

	epsilon := TimeID{0, TimeID_EntropyMask}

	for i := range prevIDs {
		prevIDs[i] = NewTimeID()
	}

	for i := 0; i < 10000000; i++ {
		now := NewTimeID()

		for _, prev := range prevIDs {
			prev = prev.Sub(epsilon)
			if now.CompareTo(prev) < 0 {
				t.Errorf("%v > %v ", prev, now)
			}
		}

		prevIDs[i&63] = now
	}
}

func TestEncodings(t *testing.T) {
	tid := TimeID{0x7777777777777777, 0x123456789abcdef0}
	if tid.Base32Suffix() != "g2ectrrh" {
		t.Errorf("TimeID.Base32Suffix() failed")
	}
	if tid.Base16Suffix() != "bcdef0" {
		t.Errorf("TimeID.Base16Suffix() failed")
	}

}
