package qgroup

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

type Func func() error

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

func (g *QGroup) Do(key string, f Func) error {
	g.mu.Lock()
	if _, ok := g.q[key]; !ok {
		g.q[key] = make(chan *Func, g.maxQueue)
		go g.doCall(key) //Calling funcs from queue
	}
	g.mu.Unlock()

	g.q[key] <- &f

	return nil
}

func (g *QGroup) Cancel() error {
	g.cancelFunc()
	return nil
}

func (g *QGroup) doCall(key string) {
	for {
		select {
		case fn := <-g.q[key]:
			f := *fn
			f()
		case <-g.ctx.Done():
			return
		}
	}
}
