queue
=====

#### Run tasks on seperate goroutines with control

### Basic Example

    import "github.com/otium/queue"
    
    q := queue.NewQueue(func(val interface{}) {
    
    
    }, 20)
    for i := 0; i < 200; i++ {
        q.Push(i)
    }
    q.Wait()
    

Documentation: http://godoc.org/github.com/otium/queue
