package asinch

import "sync"

type inProgressTracker struct {
	sync.RWMutex
	i *int
}

func (p *inProgressTracker) incr() {
	p.Lock()
	if p.i != nil {
		incremented := *(p.i) + 1
		p.i = &incremented
	}
	p.Unlock()
}

func (p *inProgressTracker) decr() {
	p.Lock()
	if p.i != nil {
		decremented := *(p.i) - 1
		p.i = &decremented
	}
	p.Unlock()
}
