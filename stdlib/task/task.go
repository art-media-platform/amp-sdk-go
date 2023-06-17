package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/arcspace/go-arc-sdk/stdlib/log"
)

// ctx implements Context
type ctx struct {
	log.Logger

	task Task

	id             int64
	state          int32
	idleClose      int32
	idleCloseDelay time.Duration
	idle           bool
	chClosing      chan struct{}  // signals Close() has been called and close execution has begun.
	chClosed       chan struct{}  // signals Close() has been called and all close execution is done.
	err            error          // See context.Err() for spec
	busy           sync.WaitGroup // blocks until all execution is complete
	subsMu         sync.Mutex     // Locked when .subs is being accessed
	subs           []Context
}

// Errors
var (
	ErrAlreadyStarted = errors.New("already started")
	ErrUnstarted      = errors.New("unstarted")
	ErrClosed         = errors.New("closed")
)

var gSpawnCounter = int64(0)

func (p *ctx) Close() error {
	first := atomic.CompareAndSwapInt32(&p.state, Running, Closing)
	if first {
		close(p.chClosing)
	}
	return nil
}

func (p *ctx) CloseWhenIdle(delay time.Duration) {

	// Allow subsequent calls to set a new delay
	p.idleCloseDelay = delay

	// Ensure only one timer for a ctx is every running
	// Can this be folded into the main go routine in StartChild() to save a goroutine?
	first := atomic.CompareAndSwapInt32(&p.idleClose, 0, 1)
	if first {
		go func() {
			var timer *time.Timer

			for waiting := true; waiting; {
				p.idle = true // setup idle detection
				p.busy.Wait() // wait until there is a chance of catching ctx idle

				if delay := p.idleCloseDelay; delay > time.Microsecond {
					if timer == nil {
						timer = time.NewTimer(delay)
					} else {
						timer.Reset(delay)
					}
					select {
					case <-timer.C:
					case <-p.Closing():
					}
				}

				// If no new children were added while we were waiting, then we have been idle and can close.
				// Note in the case that we're closing, the below has no effect
				p.subsMu.Lock()
				if p.idle {
					p.Close()
					waiting = false
				}
				p.subsMu.Unlock()
			}
		}()
	}
}

func (p *ctx) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (p *ctx) Err() error {
	select {
	case <-p.Done():
		if p.err == nil {
			return context.Canceled
		}
		return p.err
	default:
		return nil
	}
}

func (p *ctx) Value(key interface{}) interface{} {
	return nil
}

func (p *ctx) TaskRef() interface{} {
	return p.task.TaskRef
}

func (p *ctx) ContextID() int64 {
	return p.id
}

func (p *ctx) Label() string {
	return p.task.Label
}

func printContextTree(ctx Context, out *strings.Builder, depth int) {
	out.WriteString(fmt.Sprintf("%s%03d %s\n", strings.Repeat("    ", depth), ctx.ContextID(), ctx.Label()))

	var subBuf [20]Context
	children := ctx.GetChildren(subBuf[:0])
	for _, ci := range children {
		printContextTree(ci, out, depth+1)
	}
}

func (p *ctx) GetChildren(in []Context) []Context {
	p.subsMu.Lock()
	defer p.subsMu.Unlock()
	return append(in, p.subs...)
}

func (p *ctx) ChildCount() int {
	p.subsMu.Lock()
	defer p.subsMu.Unlock()
	return len(p.subs)
}

// StartChild starts the given child Context as a "sub" task.
func (p *ctx) StartChild(task *Task) (Context, error) {
	child := &ctx{
		state:     Running,
		id:        atomic.AddInt64(&gSpawnCounter, 1),
		chClosing: make(chan struct{}),
		chClosed:  make(chan struct{}),
	}
	if task != nil {
		child.task = *task
	}
	if child.task.Label == "" {
		child.task.Label = fmt.Sprintf("ctx_%d", child.id)
	}
	child.Logger = log.NewLogger(child.task.Label)

	// If a parent is given, add the child to the parent's list of children.
	if p != nil {

		var err error
		p.subsMu.Lock()
		if p.state == Running {
			p.busy.Add(1)
			p.idle = false
			p.subs = append(p.subs, child)
		} else {
			err = ErrUnstarted
		}
		p.subsMu.Unlock()

		if err != nil {
			return nil, err
		}
	}

	go func() {

		// If there is a parent, wait until child.Close() *or* p.Close()
		// TODO: merge CloseWhenIdle() into this block?
		if p != nil {
			select {
			case <-p.Closing():
				child.Close()
			case <-child.Closing():
			}
		}

		// Wait for child to begin closing phase
		<-child.Closing()

		// Fire callback if given
		if child.task.OnClosing != nil {
			child.task.OnClosing()
		}

		// Once all child's children are closed, proceed with completion.
		child.busy.Wait()

		closeParent := false

		if p != nil {

			p.subsMu.Lock()
			{
				// remove the child from its parent
				N := len(p.subs)
				for i := 0; i < N; i++ {
					if p.subs[i] == child {
						copy(p.subs[i:], p.subs[i+1:N])
						N--
						p.subs[N] = nil // show GC some love
						p.subs = p.subs[:N]
						break
					}
				}

				// If removing the last child and in IdleClose mode, queue the parent to be closed
				if N == 0 && p.task.IdleClose > 0 {
					closeParent = true
				}
			}
			p.subsMu.Unlock()
		}

		// Move to Closed state now that all all that remains is the OnClosed callback and release of the chClosed chan.
		child.state = Closed
		if child.task.OnClosed != nil {
			child.task.OnClosed()
		}
		close(child.chClosed)

		// With child no fully closed, the parent is no longer waiting on this child
		if p != nil {
			p.busy.Done()
		}

		if closeParent {
			p.CloseWhenIdle(p.task.IdleClose)
		}
	}()

	if child.task.OnStart != nil {
		err := child.task.OnStart(child)
		child.task.OnStart = nil
		if err != nil {
			child.Close()
			return nil, err
		}
	}

	if child.task.OnRun != nil {
		child.busy.Add(1)
		go func() {
			child.task.OnRun(child)
			child.task.OnRun = nil
			child.busy.Done()

			// If idleclose is set, try to do so
			if child.task.IdleClose > 0 {
				child.CloseWhenIdle(child.task.IdleClose)
			}
		}()
	}

	return child, nil
}

func (p *ctx) Go(label string, fn func(ctx Context)) (Context, error) {
	return p.StartChild(&Task{
		Label:     label,
		IdleClose: time.Nanosecond,
		OnRun:     fn,
	})
}

func (p *ctx) Closing() <-chan struct{} {
	return p.chClosing
}

func (p *ctx) Done() <-chan struct{} {
	return p.chClosed
}

const (
	Unstarted int32 = iota
	Running
	Closing
	Closed
)
