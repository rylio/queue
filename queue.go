package queue

import (
	"sync"
)

type Queue struct {
	push             chan func()
	pop              chan func()
	suspend          chan bool
	close            chan bool
	ConcurrencyLimit int
	wg               sync.WaitGroup
	suspended        bool
}

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

func (q *Queue) Wait() {

	q.wg.Wait()
}

func (q *Queue) Suspend() {

	q.suspend <- true
	q.suspended = true
}

func (q *Queue) Resume() {

	q.suspend <- false
	q.suspended = false

}

func (q *Queue) Suspended() bool {

	return q.suspended
}

func (q *Queue) Close(immediate bool) {
	q.close <- immediate

}
