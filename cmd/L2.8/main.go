package main

import (
	"fmt"
	"sync"
	"time"
)

var or = func(channels ...<-chan interface{}) <-chan interface{} {
	if len(channels) == 0 {
		ch := make(chan interface{})
		close(ch)
		return ch
	}

	orChan := make(chan interface{})
	var mu sync.Mutex
	closed := false

	for _, ch := range channels {
		go func(ch <-chan interface{}) {
			_, ok := <-ch
			if !ok {
				mu.Lock()
				if !closed {
					close(orChan)
					closed = true
				}
				mu.Unlock()
			}
		}(ch)
	}

	return orChan
}

func sig(after time.Duration) <-chan interface{} {
	c := make(chan interface{})
	go func() {
		defer close(c)
		time.Sleep(after)
	}()
	return c
}

func main() {
	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v\n", time.Since(start))
}
