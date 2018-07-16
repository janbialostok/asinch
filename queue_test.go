package asinch

import (
	"log"
	"testing"
	"time"
)

func TestCreateQueue(t *testing.T) {
	processor := func(argv []interface{}) interface{} {
		timer := time.NewTimer(time.Duration(1) * time.Second)
		value := argv[0].(int)
		<-timer.C
		log.Println("logged value", value)
		return value + 1
	}
	concurrency := 2
	results, err := AsyncMap([][]interface{}{
		[]interface{}{1},
		[]interface{}{2},
		[]interface{}{3},
		[]interface{}{4},
		[]interface{}{5},
		[]interface{}{6},
		[]interface{}{7},
	}, &concurrency, processor)
	log.Println("queue result", results, err)
}
