package asinch

import (
	"sync"
)

type AsyncMapResults struct {
	sync.RWMutex
	Results []interface{}
}

func MakeAsyncMapProcessor(method func([]interface{}) interface{}, length int) (QueueProcessor, *AsyncMapResults) {
	results := &AsyncMapResults{
		Results: make([]interface{}, length),
	}
	setResult := func(i int, value interface{}) {
		results.Lock()
		results.Results[i] = value
		results.Unlock()
	}
	return QueueProcessor(func(node *QueueNode) {
		currentIndex := node.Argv[len(node.Argv)-1]
		result := method(node.Argv)
		setResult(currentIndex.(int), result)
		node.Resolve <- result
	}), results
}

func AsyncMap(iterable [][]interface{}, concurrency *int, processor func([]interface{}) interface{}) ([]interface{}, error) {
	qp, results := MakeAsyncMapProcessor(processor, len(iterable))
	asyncQueue := CreateQueue(qp, concurrency)
	for i, argv := range iterable {
		asyncQueue.Append(append(argv, i))
	}
	asyncQueue.Start()
	<-asyncQueue.Done
	if asyncQueue.Error != nil {
		return nil, asyncQueue.Error
	}
	return results.Results, nil
}
