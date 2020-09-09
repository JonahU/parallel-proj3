package barrier

// This barrier only works for two channels, it could be improved further
type ChanBarrier struct {
	threadCount int
	done 		[]<-chan struct{}
}

func BarrierCreate(threadCount int, done ...<-chan struct{}) *ChanBarrier {
	return &ChanBarrier{threadCount, done}
}

func (bar *ChanBarrier) Wait() {
	// this could be further improved/ generalized with reflection select
	// or using the merging channels idiom
	for i:=0; i<bar.threadCount; i++ {
		select {
		case <-bar.done[0]: // wait for writers to finish writing current frame
		// do nothing
		case <-bar.done[1]: // wait for simulation worker to finish updating next frame
		// do nothing
		}
	}
}