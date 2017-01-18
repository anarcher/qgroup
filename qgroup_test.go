package qgroup

import (
	"sync"
	"testing"
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

	fn := func() error {
		result[cnt] = index[cnt]
		cnt++
		wg.Done()
		return nil
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
			mu.Lock()
			test.dest[k] = make([]string, len(xs))
			mu.Unlock()
			for index, x := range xs {
				_index := index
				_x := x
				g.Do(k, func() error {
					mu.Lock()
					test.dest[_k][_index] = _x
					mu.Unlock()
					wg.Done()
					return nil
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
