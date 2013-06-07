package queue

import "testing"
import "time"

func assert(b bool, s string, t *testing.T) {

	if !b {

		t.Fatal(s)
	}
}

func TestLimitedQueue(t *testing.T) {

	q := Init(6)
	counter := 0
	for i := 0; i < 200; i++ {

		q.Push(func() {
			defer func() { counter++ }()
			time.Sleep(10 * time.Millisecond)
		})
	}
	q.Wait()
	q.Close()
	assert(counter == 200, "Incorrect Operation Count", t)
}

func TestUnlimitedQueue(t *testing.T) {

	q := Init(0)
	counter := 0
	for i := 0; i < 200; i++ {

		q.Push(func() {

			defer func() { counter++ }()
			time.Sleep(200 * time.Millisecond)
		})
	}
	q.Wait()
	q.Close()
	assert(counter == 200, "Incorrect Operation Count", t)
}

func TestLargeScaleQueue(t *testing.T) {

	q := Init(500)
	counter := 0
	for i := 0; i < 100000; i++ {

		q.Push(func() {
			defer func() { counter++ }()
			time.Sleep(time.Millisecond)

		})
	}

	q.Wait()
	q.Close()
	assert(counter == 100000, "Incorrect Operation Count", t)
}
