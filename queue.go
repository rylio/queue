// Package queue is a utility to queue goroutines
package queue

import "sync"

// Queue is the queue.
type Queue struct {
	push             chan func()
	pop              chan func()
	suspend          chan bool
	close            chan bool
	ConcurrencyLimit int
	wg               sync.WaitGroup
	suspended        bool
}
// Init initializes a new queue.
func Init() (q *Queue) {

	q = &Queue{

		push:    make(chan func()),
		pop:     make(chan func()),
		suspend: make(chan bool),
		close:   make(chan bool),
	}
	go q.run()

	return q

}
// Push addes a new function to the end of the Queue
func (q *Queue) Push(f func()) {

	q.push <- f
}

func (q *Queue) run() {

	suspended := false
	closed := false
	finishing := false
	count := 0
	buffer := make([]func(), 0)

	exec := func() {

		if !(suspended || closed) {
			for len(buffer) > 0 && (count < q.ConcurrencyLimit || q.ConcurrencyLimit == 0) {

				f := buffer[0]
				buffer = buffer[1:]
				count++
				go func() {

					f()
					defer func() {
						q.pop <- f
					}()
				}()

			}

		}

	}

	cleanUp := func() {
		q.wg.Add(-len(buffer))
		close(q.push)
		close(q.pop)
		close(q.suspend)
		close(q.close)
	}

	for {

		select {

		case closed = <-q.close:
			finishing = true

		case suspended = <-q.suspend:

		case <-q.pop:
			count--
			q.wg.Done()

		case v := <-q.push:

			if !finishing {
				q.wg.Add(1)
				buffer = append(buffer, v)
			}
		}
		exec()
		if count == 0 && finishing {
			if len(buffer) == 0 {
				closed = true
			}
			if closed {
				cleanUp()
				return
			}
		}

	}

}

// Wait calls Wait on a WaitGroup, which finishes when the Queue is drained or when closed
func (q *Queue) Wait() {

	q.wg.Wait()
}

// Suspend stops the execution of functions on the queue, new functions can still be queued.
func (q *Queue) Suspend() {

	q.suspend <- true
	q.suspended = true
}

// Resume starts the execution of functions on the queue, after Suspend has been called.
func (q *Queue) Resume() {

	q.suspend <- false
	q.suspended = false

}

// Suspended returns whether or not the queue is in a suspended state.
func (q *Queue) Suspended() bool {

	return q.suspended
}

// Close cleanups the queue by closing the internal channels.
// If true is passed as an arguement, the queue closes immediatley only waiting for the currently executing functions to finish.
// If false is passed the queue finishes executing all the functions in the queue.
// When called the queue does not allow for anymore functions to be enqueued.
func (q *Queue) Close(immediate bool) {
	q.close <- immediate

}
