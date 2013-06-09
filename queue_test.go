package queue

import (
	"testing"
	"time"
)

func assert(b bool, s string, t *testing.T) {

	if !b {

		t.Fatal(s)
	}
}

func TestSuspend(t *testing.T) {

	q := Init()
	q.ConcurrencyLimit = 5
	for i := 0; i < 10; i++ {
		q.Push(func() {
			time.Sleep(100 * time.Millisecond)
		})
	}
	q.Suspend()
	time.Sleep(time.Millisecond * 200)
	q.Resume()
	q.Wait()
	q.Close(false)

}

func TestCloseImmediate(t *testing.T) {

	q := Init()
	q.ConcurrencyLimit = 1
	count := 0
	for i := 0; i < 8; i++ {

		q.Push(func() {

			defer func() { count++ }()
			time.Sleep(50 * time.Millisecond)

		})
	}
	time.Sleep(time.Millisecond * 180)
	q.Close(true)
	q.Wait()
	assert(count == 4, "", t)
}

func TestCloseAfter(t *testing.T) {

	q := Init()
	q.ConcurrencyLimit = 1
	count := 0

	for i := 0; i < 5; i++ {

		q.Push(func() {
			defer func() { count++ }()
			time.Sleep(100 * time.Millisecond)
		})
	}
	q.Close(false)
	q.Wait()
	assert(count == 5, "", t)
}
