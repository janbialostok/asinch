package asinch

import (
	"sync"
)

type QueueProcessor func(*QueueNode)
type ResultsHandler func([]QueueNode, QueueNode) []QueueNode

type QueueNode struct {
	Argv    []interface{}
	Result  interface{}
	Next    *QueueNode
	Prev    *QueueNode
	Resolve chan interface{}
	Reject  chan error
}

type Queue struct {
	sync.RWMutex
	Root           *QueueNode
	Concurrency    *int
	ProcessorFunc  QueueProcessor
	Paused         bool
	Running        bool
	Error          error
	Done           chan bool
	Results        []QueueNode
	ResultsHandler ResultsHandler
}

func (queue *Queue) Append(argv ...[]interface{}) *Queue {
	for _, a := range argv {
		node := &QueueNode{
			Argv:    a,
			Resolve: make(chan interface{}, 2),
			Reject:  make(chan error),
		}
		queue.Lock()
		if queue.Root == nil {
			queue.Root = node
		} else {
			var current = queue.Root
			for current.Next != nil {
				current = (*current).Next
			}
			node.Prev = current
			current.Next = node
		}
		queue.Unlock()
	}
	return queue
}

func (queue *Queue) Preppend(argv ...[]interface{}) *Queue {
	for _, a := range argv {
		node := &QueueNode{
			Argv:    a,
			Resolve: make(chan interface{}, 2),
			Reject:  make(chan error),
		}
		queue.Lock()
		if queue.Root == nil {
			queue.Root = node
		} else {
			node.Next = queue.Root
			queue.Root = node.Next
			queue.Root = node
		}
		queue.Unlock()
	}
	return queue
}

func (queue *Queue) Pause() *Queue {
	if queue.Paused != true {
		queue.Lock()
		queue.Paused = true
		queue.Running = false
		queue.Unlock()
	}
	return queue
}

func handleProcessing(queue *Queue, pt *inProgressTracker, processing *QueueNode) {
	go queue.ProcessorFunc(processing)
	select {
	case processing.Result = <-processing.Resolve:
		queue.Lock()
		if queue.ResultsHandler != nil {
			queue.Results = queue.ResultsHandler(queue.Results, *processing)
		} else {
			queue.Results = append(queue.Results, *processing)
		}
		queue.Unlock()
		pt.decr()
		pt.Lock()
		defer pt.Unlock()
		if *pt.i == 0 {
			queue.Exec()
		}
	case err := <-processing.Reject:
		pt.decr()
		queue.Error = err
	}
}

func (queue *Queue) Exec() *Queue {
	pending := 0
	inProgress := &inProgressTracker{
		i: &pending,
	}
	go func() {
		queue.Lock()
		defer queue.Unlock()
		for {
			current := queue.Root
			if current != nil && !queue.Paused {
				if queue.Concurrency != nil && *inProgress.i < *queue.Concurrency {
					inProgress.incr()
					queue.Root = current.Next
					go handleProcessing(queue, inProgress, current)
				} else if queue.Concurrency == nil {
					inProgress.incr()
					queue.Root = current.Next
					go handleProcessing(queue, inProgress, current)
				} else {
					break
				}
			} else {
				break
			}
		}
		if *inProgress.i == 0 && (queue.Paused || queue.Root == nil) {
			queue.Done <- true
		}
	}()
	return queue
}

func (queue *Queue) Start() *Queue {
	if queue.Running != true {
		queue.Lock()
		queue.Paused = false
		queue.Running = true
		queue.Done = make(chan bool)
		queue.Unlock()
		return queue.Exec()
	}
	return queue
}

func CreateQueue(processor QueueProcessor, concurrency *int) *Queue {
	return &Queue{
		ProcessorFunc: processor,
		Concurrency:   concurrency,
	}
}
