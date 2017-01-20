package qgroup

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func Test_QGroupCalls(t *testing.T) {
	var wg sync.WaitGroup
	var (
		cnt    int
		result []int
		index  []int
	)
	index = []int{3, 1}
	result = make([]int, len(index))

	fn := func(ctx context.Context) {
		result[cnt] = index[cnt]
		cnt++
		wg.Done()
	}

	g := NewGroup(WithMaxQueue(10))
	wg.Add(2)
	g.Do("test", fn)
	g.Do("test", fn)
	wg.Wait()

	for k, i := range index {
		if result[k] != i {
			t.Errorf("calling is not lineration (%v != %v)", index, result)
		}
	}
}

func Test_QGroupMultiCalls(t *testing.T) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var tests = []struct {
		dest map[string][]string
		src  map[string][]string
	}{
		{
			dest: make(map[string][]string),
			src:  map[string][]string{"test": {"1", "2"}},
		},
		{
			dest: make(map[string][]string),
			src:  map[string][]string{"test": {"1", "2"}, "test2": {"3", "4"}},
		},
	}

	for i, test := range tests {
		g := NewGroup(WithMaxQueue(10))
		for k, xs := range test.src {
			_k := k
			wg.Add(len(xs))
			test.dest[k] = make([]string, len(xs))
			for index, x := range xs {
				_index := index
				_x := x
				g.Do(k, func(ctx context.Context) {
					mu.Lock()
					test.dest[_k][_index] = _x
					mu.Unlock()

					wg.Done()
				})
			}
		}
		wg.Wait()
		for k, xs := range test.src {
			for j, x := range xs {
				if test.dest[k][j] != x {
					t.Errorf("%v. %v mismatched: want=%v have=%v", i+1, k, test.src[k], test.dest[k])
				}
			}
		}
	}
}

func Test_QGroupCallWithTimeout(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	fn := func(ctx context.Context) {
		<-ctx.Done()
		wg.Done()
	}

	g := NewGroup(WithTimeout(100 * time.Millisecond))
	g.Do("test", fn)
	g.Do("test", fn)
	wg.Wait()

}
