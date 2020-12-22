package rpc

type semaphore chan struct{}

func makeSemaphore(n int) semaphore {
	return make(chan struct{}, n)
}

func (s semaphore) acquire() {
	s <- struct{}{}
}

func (s semaphore) release() {
	<-s
}
