package queue

import "sync"

type Queue struct {
	limit chan struct{}
	wg    sync.WaitGroup
}

func Init(maxOps uint) (q *Queue) {

	q = new(Queue)
	if maxOps > 0 {
		q.limit = make(chan struct{}, maxOps)
	}

	return q
}

func (q *Queue) Push(f func()) {
	q.wg.Add(1)
	go func() {

		if q.limit != nil {
			q.limit <- struct{}{}
			defer func() { <-q.limit }()
		}
		defer q.wg.Done()

		f()
	}()
}

func (q *Queue) Wait() {

	q.wg.Wait()
}

func (q *Queue) Close() {
	if q.limit != nil {
		close(q.limit)
	}
}
