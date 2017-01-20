package qgroup

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

// Func has func(context.Context) and context.
// The context uses with DoWithContext().
type Func struct {
	fn  func(context.Context)
	ctx context.Context
}

//QGroup represents  function calling queues and settings
type QGroup struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	maxQueue   int
	timeout    time.Duration
	mu         sync.Mutex
	q          map[string]chan *Func
}

type QGroupOption func(g *QGroup)

func WithTimeout(d time.Duration) QGroupOption {
	return func(g *QGroup) {
		g.timeout = d
	}
}

func WithMaxQueue(maxQueue int) QGroupOption {
	return func(g *QGroup) {
		g.maxQueue = maxQueue
	}
}

func NewGroup(opts ...QGroupOption) *QGroup {
	g := &QGroup{} //Default QGroup
	g.ctx, g.cancelFunc = context.WithCancel(context.Background())
	g.q = make(map[string]chan *Func)

	for _, o := range opts {
		o(g)
	}
	return g
}

func (g *QGroup) Do(key string, fn func(context.Context)) error {
	g.startLoopCall(key)
	g.q[key] <- &Func{fn: fn}

	return nil
}

func (g *QGroup) DoWithContext(key string, ctx context.Context, fn func(context.Context)) error {
	g.startLoopCall(key)
	g.q[key] <- &Func{fn, ctx}

	return nil
}

func (g *QGroup) Cancel() error {
	g.cancelFunc()
	return nil
}

func (g *QGroup) startLoopCall(key string) {
	g.mu.Lock()
	if _, ok := g.q[key]; !ok {
		g.q[key] = make(chan *Func, g.maxQueue)
		go g.loopCall(key) //Calling funcs from queue
	}
	g.mu.Unlock()

}

func (g *QGroup) loopCall(key string) {
	g.mu.Lock()
	q := g.q[key]
	g.mu.Unlock()

	for {
		select {
		case fn := <-q:
			g.doCall(fn)
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *QGroup) doCall(fn *Func) {
	var (
		ctx  context.Context
		pctx context.Context
	)

	if fn.ctx != nil {
		pctx = fn.ctx
	} else {
		pctx = context.Background()
	}

	if g.timeout > 0 {
		ctx, _ = context.WithTimeout(pctx, g.timeout)
	} else {
		ctx = pctx
	}

	c := make(chan struct{}, 1)

	go func() {
		fn.run(ctx)
		c <- struct{}{}
	}()

	<-c
}

func (fn Func) run(ctx context.Context) {
	fn.fn(ctx)
}
