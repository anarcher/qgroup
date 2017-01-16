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
