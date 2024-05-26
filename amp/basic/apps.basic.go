package basic

import (
	"time"

	"github.com/amp-3d/amp-sdk-go/amp"
	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
	"github.com/amp-3d/amp-sdk-go/stdlib/task"
)

type Cell[AppT amp.AppInstance] interface {
	Info() *CellInfo[AppT]

	MarshalAttrs(pin *Pin[AppT])

	// Tells this cell it has been pinned and should synchronously update itself accordingly.
	PinInto(dst *Pinned[AppT]) error

	// Called if PinInto() called and is no longer needed (because requests against it have closed.
	// This allows a Cell to release resources it may locked during PinInto()..
	ReleasePin()

	// Describes this Cell internally for logging and debugging.
	GetLogLabel() string
}

// CellInfo, Cell, and Pinned are optional support types for implementing amp.Pin and amp.AppInstance.
type CellInfo[AppT amp.AppInstance] struct {
	ID     tag.ID        // universal ID for this cell
	Tab    amp.TagTab    // assumed that every cell has some basic info
	Pinned *Pinned[AppT] // non-nil if this cell has been pinned
}

func (cell *CellInfo[AppT]) IsPinned() bool {
	return cell.Pinned != nil
}

func (cell *CellInfo[AppT]) ReleasePin() {
	// override for cleanup
}

func (cell *CellInfo[AppT]) Info() *CellInfo[AppT] {
	return cell
}

func (cell *CellInfo[AppT]) GetLogLabel() string {
	if cell.Tab.Label != "" {
		return cell.Tab.Label // TODO: add shortening for long labels and use Tag.Caption if available
	}
	if !cell.ID.IsNil() {
		return cell.ID.String()
	}
	return "???"
}

func (cell *CellInfo[AppT]) MarshalAttrs(pin *Pin[AppT]) {
	var attrID tag.ID
	if cell.Pinned != nil {
		attrID = amp.PinnedTabSpec.ID
	} else {
		attrID = amp.ChildTabSpec.ID
	}
	pin.Upsert(cell.ID, attrID, tag.Nil, &cell.Tab)
}

func PinAndServe[AppT amp.AppInstance](target Cell[AppT], app AppT, op amp.Requester) (amp.Pin, error) {

	cell := target.Info()
	if cell.Pinned == nil {
		if cell.ID.IsNil() {
			cell.ID = tag.New()
		}

		cell.Pinned = &Pinned[AppT]{
			App:      app,
			Cell:     target,
			children: make(map[tag.ID]Cell[AppT]),
		}

		err := target.PinInto(cell.Pinned)
		if err != nil {
			cell.Pinned = nil
			return nil, err
		}
	}

	pin := &Pin[AppT]{
		Op:     op,
		Pinned: cell.Pinned,
	}

	var err error
	pin.ctx, err = app.StartChild(&task.Task{
		Label:     "pin: " + target.GetLogLabel(),
		IdleClose: time.Microsecond,
		OnRun: func(pinContext task.Context) {
			err := pin.pushTx()
			if err != nil {
				if err != amp.ErrShuttingDown {
					pinContext.Warnf("op failed: %v", err)
				}
			} else if op.Request().PinSync == amp.PinSync_Maintain {
				<-pinContext.Closing()
			}

			op.OnComplete(err)
		},
		OnClosing: func() {
			target.ReleasePin()
		},
	})
	if err != nil {
		return nil, err
	}

	return pin, nil
}

// App is a helper for implementing AppInstance.
// It is typically extended by embedding it into a struct that builds on top of it.
type App[AppT amp.AppInstance] struct {
	amp.AppContext
	Instance AppT
}

func (app *App[AppT]) MakeReady(op amp.Requester) error {
	return nil
}

func (app *App[AppT]) PinAndServe(target Cell[AppT], op amp.Requester) (amp.Pin, error) {
	return PinAndServe(target, app.Instance, op)
}

func (app *App[AppT]) OnClosing() {
}

type Pinned[AppT amp.AppInstance] struct {
	Cell     Cell[AppT]
	App      AppT
	children map[tag.ID]Cell[AppT]
}

/*
	type AppChannel[AppT AppInstance] struct {
		Spec       tag.Spec // Implies one or more attr specs that an App sends
		Marshaller func()   //dst *Pinned[AppT])
	}

	func (op *Pinned[AppT]) Declare(ch AppChannel[AppT]) {
		op.Channels = append(op.Channels, ch)
	}
*/

func (op *Pinned[AppT]) AddChild(sub Cell[AppT]) {
	cell := sub.Info()
	if cell.ID.IsNil() {
		cell.ID = tag.New()
	}
	op.children[cell.ID] = sub
}

func (pin *Pinned[AppT]) GetCell(target tag.ID) Cell[AppT] {
	if target == pin.Cell.Info().ID {
		return pin.Cell
	}
	return pin.children[target]
}

type Pin[AppT amp.AppInstance] struct {
	Pinned *Pinned[AppT]
	Op     amp.Requester
	Tx     *amp.TxMsg

	err error
	ctx task.Context
}

func (pin *Pin[AppT]) Context() task.Context {
	return pin.ctx
}

func (pin *Pin[AppT]) Upsert(targetID, attrID, SI tag.ID, val amp.ElemVal) {
	txOp := amp.TxOp{
		OpCode:   amp.TxOpCode_UpsertAttr,
		TargetID: targetID,
		AttrID:   attrID,
		SI:       SI,
	}
	if pin.err != nil {
		return
	}
	pin.err = pin.Tx.MarshalOp(&txOp, val)
}

func (pin *Pin[AppT]) pushTx() error {
	pin.Tx = amp.NewTxMsg(true)
	pin.Pinned.Cell.MarshalAttrs(pin)
	if pin.err != nil {
		return pin.err
	}

	for _, sub := range pin.Pinned.children {
		sub.MarshalAttrs(pin)
		if pin.err != nil {
			return pin.err
		}
	}

	tx := pin.Tx
	tx.Status = amp.OpStatus_Synced
	pin.Tx = nil
	return pin.Op.PushTx(tx)
}

// func (pin *Pinned[AppT]) ProcessTx(tx *TxMsg) error {
// 	return ErrUnimplemented
// }

// func (pin *Pinned[AppT]) GetLogLabel() string {
// 	return pin.Cell.GetLogLabel()
// }

// func (pin *Pinned[AppT]) OnClosing() {
// 	// override for cleanup
// }

func (pin *Pin[AppT]) ServeRequest(op amp.Requester) (amp.Pin, error) {
	req := op.Request()
	cell := pin.Pinned.GetCell(req.TargetID())
	if cell == nil {
		return nil, amp.ErrCellNotFound
	}
	return PinAndServe(cell, pin.Pinned.App, op)
}
