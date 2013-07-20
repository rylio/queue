package queue

import (
	"testing"
	"sync/atomic"
)

func TestQueue (t *testing.T) {

	var count int32 = 0
	handler := func(val interface{}) {

		atomic.AddInt32(&count, int32(val.(int)))

	}
	q := NewQueue(handler, 5)

	for i := 0; i < 200; i++ {

		q.Push(i)
	}

	q.Wait()
	if count != 19900 {

		t.Fail()
	}

}