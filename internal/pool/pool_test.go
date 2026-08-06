package pool

import (
	"sync"
	"testing"
)

type testObject struct {
	value    int
	data     []int
	resetCnt int
}

func (o *testObject) Reset() {
	o.value = 0
	o.data = o.data[:0]
	o.resetCnt++
}

func TestNew(t *testing.T) {
	p := New(func() *testObject { return &testObject{} })
	if p == nil {
		t.Fatal("New вернул nil")
	}
}

func TestGetReturnsNewObject(t *testing.T) {
	created := 0
	p := New(func() *testObject {
		created++
		return &testObject{value: 42}
	})

	obj := p.Get()
	if obj == nil {
		t.Fatal("Get вернул nil")
	}
	if obj.value != 42 {
		t.Errorf("ожидалось значение 42, получено %d", obj.value)
	}
	if created != 1 {
		t.Errorf("ожидался один вызов newFn, получено %d", created)
	}
}

func TestPutCallsReset(t *testing.T) {
	p := New(func() *testObject { return &testObject{} })

	obj := &testObject{value: 100, data: []int{1, 2, 3}}
	p.Put(obj)

	if obj.value != 0 {
		t.Errorf("ожидалось, что value сбросится в 0, получено %d", obj.value)
	}
	if len(obj.data) != 0 {
		t.Errorf("ожидалось, что data сбросится, длина %d", len(obj.data))
	}
	if obj.resetCnt != 1 {
		t.Errorf("ожидался один вызов Reset, получено %d", obj.resetCnt)
	}
}

func TestGetAfterPutReuses(t *testing.T) {
	created := 0
	p := New(func() *testObject {
		created++
		return &testObject{}
	})

	obj := p.Get() // created == 1
	obj.value = 55
	p.Put(obj) // Reset -> value == 0

	got := p.Get()
	if got.value != 0 {
		t.Errorf("ожидался сброшенный объект, получено value %d", got.value)
	}
	// sync.Pool не гарантирует переиспользование, но newFn не должен
	// вызываться больше, чем число промахов по пулу.
	if created > 2 {
		t.Errorf("неожиданное число аллокаций: %d", created)
	}
}

func TestConcurrentUsage(t *testing.T) {
	p := New(func() *testObject { return &testObject{} })

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				obj := p.Get()
				obj.value = j
				obj.data = append(obj.data, j)
				p.Put(obj)
			}
		}()
	}
	wg.Wait()
}
