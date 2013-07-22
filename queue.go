// Package queue provides a queue for running tasks concurrently up to a certain limit.
package queue

import (
	"runtime"
	"sync"
)

type queue struct {
	Handler          func(interface{})
	ConcurrencyLimit int
	push             chan interface{}
	pop              chan struct{}
	suspend          chan bool
	suspended        bool
	stop             chan struct{}
	stopped          bool
	buffer           []interface{}
	count            int
	wg               sync.WaitGroup
}

// Queue is the queue
// Queue also has the members Handler and ConcurrencyLimit which can be set at anytime
type Queue struct {
	*queue
}

// Handler is a function that takes any value, and is called every time a task is processed from the queue
type Handler func(interface{})

// NewQueue must be called to initialize a new queue.
// The first argument is a Handler
// The second argument is an int which specifies how many operation can run in parallel in the queue, zero means unlimited.
func NewQueue(handler Handler, concurrencyLimit int) *Queue {

	q := &Queue{
		&queue{
			Handler:          handler,
			ConcurrencyLimit: concurrencyLimit,
			push:             make(chan interface{}),
			pop:              make(chan struct{}),
			suspend:          make(chan bool),
			stop:             make(chan struct{}),
		},
	}

	go q.run()
	runtime.SetFinalizer(q, (*Queue).Stop)
	return q
}

// Push pushes a new value to the end of the queue
func (q *Queue) Push(val interface{}) {
	q.push <- val
}

// Stop stops the queue from executing anymore tasks, and waits for the currently executing tasks to finish.
// The queue can not be started again once this is called
func (q *Queue) Stop() {

	q.stop <- struct{}{}
	runtime.SetFinalizer(q, nil)
}

// Wait calls wait on a waitgroup, that waits for all items of the queue to finish execution
func (q *Queue) Wait() {

	q.wg.Wait()
}

// Count returns the number of currently executing tasks and the number of tasks waiting to be executed
func (q *Queue) Len() (_, _ int) {

	return q.count, len(q.buffer)
}

func (q *queue) run() {

	defer func() {
		q.wg.Add(-len(q.buffer))
		q.buffer = nil
	}()
	for {

		select {

		case val := <-q.push:
			q.buffer = append(q.buffer, val)
			q.wg.Add(1)
		case <-q.pop:
			q.count--
		case suspend := <-q.suspend:
			if suspend != q.suspended {

				if suspend {
					q.wg.Add(1)
				} else {
					q.wg.Done()
				}
				q.suspended = suspend
			}
		case <-q.stop:
			q.stopped = true
		}

		if q.stopped && q.count == 0 {

			return
		}

		for (q.count < q.ConcurrencyLimit || q.ConcurrencyLimit == 0) && len(q.buffer) > 0 && !(q.suspended || q.stopped) {

			val := q.buffer[0]
			q.buffer = q.buffer[1:]
			q.count++
			go func() {
				defer func() {
					q.pop <- struct{}{}
					q.wg.Done()
				}()
				q.Handler(val)

			}()
		}

	}

}
