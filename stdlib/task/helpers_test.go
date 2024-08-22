package task_test

import (
	"context"

	"github.com/art-media-platform/amp-sdk-go/stdlib/task"
	"github.com/art-media-platform/amp-sdk-go/stdlib/testutils"
)

func makeItems() (item [5]*workItem) {
	item[0] = &workItem{id: "a", processed: testutils.NewAwaiter()}
	item[1] = &workItem{id: "b1", processed: testutils.NewAwaiter()}
	item[2] = &workItem{id: "b2", processed: testutils.NewAwaiter()}
	item[3] = &workItem{id: "c", processed: testutils.NewAwaiter()}
	item[4] = &workItem{id: "d", processed: testutils.NewAwaiter()}
	return
}

type workItem struct {
	id        string
	retry     bool
	processed testutils.Awaiter
	block     chan struct{}
}

func (i workItem) ID() task.PoolUniqueID { return i.id[:1] }

func (i *workItem) Work(ctx context.Context) (retry bool) {
	i.processed.ItHappened()
	if i.block != nil {
		select {
		case <-i.block:
			i.block = nil
		case <-ctx.Done():
			return false
		}
	}
	retry = i.retry
	i.retry = false
	return retry
}
