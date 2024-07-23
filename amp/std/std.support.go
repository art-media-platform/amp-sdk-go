package std

import (
	fmt "fmt"
	reflect "reflect"
	"time"

	"github.com/amp-3d/amp-sdk-go/amp"
	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
	"github.com/amp-3d/amp-sdk-go/stdlib/task"
)

func (root *CellNode[AppT]) Root() *CellNode[AppT] {
	return root
}

func PinAndServe[AppT amp.AppInstance](cell Cell[AppT], app AppT, op amp.Requester) (amp.Pin, error) {
	root := cell.Root()
	if root.ID.IsNil() {
		root.ID = tag.Now()
	}

	pin := &Pin[AppT]{
		Op:       op,
		App:      app,
		Cell:     cell,
		children: make(map[tag.ID]Cell[AppT]),
	}

	label := "pin: " + root.ID.Base32Suffix()
	if app.Info().DebugMode {
		label += fmt.Sprintf(", Cell.(*%v)", reflect.TypeOf(cell).Elem().Name())
	}

	var err error
	pin.ctx, err = app.StartChild(&task.Task{
		Info: task.Info{
			Label:     label,
			IdleClose: time.Microsecond,
		},
		OnRun: func(pinContext task.Context) {
			err := pin.App.MakeReady(op)
			if err == nil {
				err = cell.PinInto(pin)
			}
			if err == nil {
				err = pin.pushState()
			}
			if err != nil {
				if err != amp.ErrShuttingDown {
					pinContext.Log().Warnf("op failed: %v", err)
				}
			} else if op.Request().StateSync == amp.StateSync_Maintain {
				<-pinContext.Closing()
			}
			op.OnComplete(err)
		},
		OnClosing: func() {
			pin.ReleasePin()
		},
	})
	if err != nil {
		return nil, err
	}

	return pin, nil
}

func (app *App[AppT]) MakeReady(op amp.Requester) error {
	return nil
}

func (app *App[AppT]) PinAndServe(cell Cell[AppT], op amp.Requester) (amp.Pin, error) {
	return PinAndServe(cell, app.Instance, op)
}

func (app *App[AppT]) OnClosing() {
}

// Called when this Pin is closed.
// This allows a Cell to release resources it may locked during PinInto()..
func (pin *Pin[AppT]) ReleasePin() {
	// override for cleanup
}

func (pin *Pin[AppT]) AddChild(sub Cell[AppT]) {
	child := sub.Root()
	childID := child.ID
	if childID.IsNil() {
		childID = tag.Now()
		child.ID = childID
	}
	pin.children[childID] = sub
}

func (pin *Pin[AppT]) GetCell(target tag.ID) Cell[AppT] {
	if target == pin.Cell.Root().ID {
		return pin.Cell
	}
	if cell, exists := pin.children[target]; exists {
		return cell
	}
	return nil
}

func (pin *Pin[AppT]) Context() task.Context {
	return pin.ctx
}

func (pin *Pin[AppT]) ServeRequest(op amp.Requester) (amp.Pin, error) {
	req := op.Request()
	cell := pin.GetCell(req.TargetID())
	if cell == nil {
		return nil, amp.ErrCellNotFound
	}
	return PinAndServe(cell, pin.App, op)
}

func (pin *Pin[AppT]) pushState() error {
	tx := amp.NewTxMsg(true)

	if pin.Op.Request().StateSync > amp.StateSync_None {
		pinnedID := pin.Cell.Root().ID

		w := cellWriter{
			tx:     tx,
			cellID: pinnedID,
		}

		tx.Upsert(amp.MetaNodeID, CellChildren.ID, pinnedID, nil) // export the root cell ID
		pin.Cell.MarshalAttrs(&w)
		if w.err != nil {
			return w.err
		}

		for childID, child := range pin.children {
			w.cellID = childID
			tx.Upsert(pinnedID, CellChildren.ID, childID, nil) // link child to pinned cell
			child.MarshalAttrs(&w)
			if w.err != nil {
				return w.err
			}
		}
	}

	tx.Status = amp.OpStatus_Synced
	return pin.Op.PushTx(tx)
}

type cellWriter struct {
	cellID tag.ID     // cache for Cell.Root().ID
	tx     *amp.TxMsg // in-progress transaction
	err    error
}

func (w *cellWriter) PutText(propertyID tag.ID, val string) {
	if w.err != nil {
		return
	}
	txOp := amp.TxOp{
		OpCode: amp.TxOpCode_UpsertElement,
		CellID: w.cellID,
		AttrID: CellProperties.ID,
		SI:     propertyID,
	}
	err := w.tx.MarshalOp(&txOp, &amp.Tag{
		Text: val,
	})
	if err != nil {
		w.err = err
	}
}

func (w *cellWriter) PutItem(propertyID tag.ID, val tag.Value) {
	if w.err != nil {
		return
	}
	txOp := amp.TxOp{
		OpCode: amp.TxOpCode_UpsertElement,
		CellID: w.cellID,
		AttrID: CellProperties.ID,
		SI:     propertyID,
	}
	if err := w.tx.MarshalOp(&txOp, val); err != nil {
		w.err = err
	}
}

func (w *cellWriter) Upsert(op *amp.TxOp, val tag.Value) {
	if w.err != nil {
		return
	}
	if err := w.tx.MarshalOp(op, val); err != nil {
		w.err = err
	}
}

/*
func (tx *TxMsg) PutMultiple(propertyIDs []tag.ID, serialize tag.Value) error {
	op := PropertyOp{}

	// serialize the value
	if serialize != nil {
		var err error
		op.DataOfs = uint64(len(tx.DataStore))
		tx.DataStore, err = serialize.MarshalToStore(tx.DataStore)
		if err != nil {
			tx.Error = err
			return err
		}
		op.DataLen = uint64(len(tx.DataStore)) - op.DataOfs
	}

	// add the op to the set
	for _, propID := range propertyIDs {
		op.PropertyID = propID
		tx.Ops = append(tx.Ops, op)
	}
	tx.OpsSorted = false
	return nil
}

func (tx *TxMsg) Upsert(literal any, keys ...tag.ID) {
	var tag *amp.Tag

	switch v := literal.(type) {
	case string:
		tag = &amp.Tag{
			InlineText: v,
			Use:        TagUse_Text,
		}
		break
	case *string:
		tag.InlineText = *v
		tag.Use = TagUse_Text
		break
	case *amp.Tag:
		tag = v
		break
	caseTag:
		tag = &v
		break
	// case *tag.ID:
	// 	tag = &amp.Tag{}
	// 	tag.SetTagID(v)
	default:
		panic("unsupported type")
		return
	}

	for _, key := range keys {
		tx.Put(tag, key)
	}
}
*/
