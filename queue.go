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

type Queue struct {
	*queue
}

func NewQueue(handler func(interface{}), concurrencyLimit int) *Queue {

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

func (q *Queue) Push(val interface{}) {
	q.push <- val
}

func (q *Queue) Stop() {

	q.stop <- struct{}{}
	runtime.SetFinalizer(q, nil)
}

func (q *Queue) Wait() {

	q.wg.Wait()
}

func (q *Queue) Count() (_, _ int) {

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
