package amp

import (
	"bytes"
	fmt "fmt"
	io "io"
	"testing"

	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
)

/*
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

}
*/
func TestTxSerialize(t *testing.T) {
	// Test serialization of a simple TxMsg

	tx := NewTxMsg(true)
	tx.Status = ReqStatus_Syncing
	tx.ContextID_0 = 888854513
	tx.ContextID_1 = 7777435
	tx.ContextID_2 = 77743773
	{
		op := TxOp{
			OpCode:   TxOpCode_MetaAttr,
			TargetID: tag.ID{4, 555, 666},
			AttrID:   tag.ID{111312232, 22232334444},
			SI:       tag.ID{7383, 76549, 3773},
			Hash:     0xfeedbeef,
		}
		tx.MarshalOp(&op, &Login{
			UserUID:  "alan1",
			HostAddr: "batwing ave",
		})
		tx.DataStore = append(tx.DataStore, []byte("bytes not used but stored -- not normal!")...)

		op.SI[1] = 50454123
		op.Height = 234
		data := []byte("hello-world")
		for i := 0; i < 7; i++ {
			data = append(data, data...)
		}
		tx.MarshalOp(&op, &Login{
			UserUID:  "cmdr6",
			HostAddr: string(data),
		})

		for i := 0; i < 5500; i++ {
			op.SI[0] = uint64(i)
			if i%5 == 0 {
				op.Height += 1
			}
			tx.MarshalOp(&op, &LoginResponse{
				HashResponse: append(data, fmt.Sprintf("-%d", i)...),
			})
		}

		op.SI[0] = 111111
		op.Hash = 55445544
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

		if op1.OpCode != op2.OpCode || op1.TargetID != op2.TargetID || op1.AttrID != op2.AttrID || op1.SI != op2.SI || op1.DataOfs != op2.DataOfs || op1.DataLen != op2.DataLen {
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

func TestNewTag(t *testing.T) {
	var prevIDs [64]tag.ID

	prevIDs[0] = tag.ID{100, (^uint64(0)) - 500}

	delta := tag.ID{100, 100}
	for i := 1; i < 64; i++ {
		prevIDs[i] = prevIDs[i-1].Add(delta)
	}
	for i := 1; i < 64; i++ {
		prev := prevIDs[i-1]
		curr := prevIDs[i]
		if prev.CompareTo(curr) >= 0 {
			t.Errorf("tag.ID.Add() returned a non-increasing value: %v <= %v", prev, curr)
		}
		if curr.Sub(prev) != delta {
			t.Errorf("tag.ID.Diff() returned a wrong value: %v != %v", curr.Sub(prev), delta)
		}
	}

	epsilon := tag.ID{0, tag.EntropyMask}

	for i := range prevIDs {
		prevIDs[i] = tag.New()
	}

	for i := 0; i < 10000000; i++ {
		now := tag.New()
		fence := now.Sub(epsilon)

		for _, prev := range prevIDs {
			comp := prev.CompareTo(fence)
			if comp >= 0 {
				t.Errorf("%v > %v ", prev, now)
			}
		}

		prevIDs[i&63] = now
	}
}

func TestEncodings(t *testing.T) {
	tid := tag.ID{0x7777777777777777, 0x123456789abcdef0}
	if tid.Base32Suffix() != "g2ectrrh" {
		t.Errorf("tag.ID.Base32Suffix() failed")
	}
	if tid.Base16Suffix() != "abcdef0" {
		t.Errorf("tag.ID.Base16Suffix() failed")
	}

}
