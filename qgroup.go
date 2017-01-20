package qgroup

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

type Func func(context.Context)

type QGroup struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	maxQueue   int
	timeout    time.Duration
	mu         sync.Mutex
	q          map[string]chan Func
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
	g.q = make(map[string]chan Func)

	for _, o := range opts {
		o(g)
	}
	return g
}

func (g *QGroup) Do(key string, f Func) error {
	g.mu.Lock()
	if _, ok := g.q[key]; !ok {
		g.q[key] = make(chan Func, g.maxQueue)
		go g.loopCall(key) //Calling funcs from queue
	}
	g.mu.Unlock()

	g.q[key] <- f

	return nil
}

func (g *QGroup) Cancel() error {
	g.cancelFunc()
	return nil
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

func (g *QGroup) doCall(fn Func) {
	var ctx context.Context

	if g.timeout > 0 {
		ctx, _ = context.WithTimeout(g.ctx, g.timeout)
	} else {
		ctx = context.Background()
	}

	c := make(chan struct{}, 1)

	go func() {
		fn(ctx)
		c <- struct{}{}
	}()

	<-c
}
