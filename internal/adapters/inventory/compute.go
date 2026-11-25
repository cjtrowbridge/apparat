// Measures sustained scalar FP32 multiply-add throughput in a short parallel CPU sample.
// Every monitor advertises the process sample without repeating it on inventory refresh.
package inventory

import (
	"runtime"
	"sync"
	"time"
)

var cpuFP32 = sync.OnceValue(measureCPUFP32)

func measureCPUFP32() float64 {
	count := min(runtime.NumCPU(), runtime.GOMAXPROCS(0))
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(count)
	work := make(chan int64, count)
	var until time.Time
	for i := range count {
		go func() {
			a, b, c, d := float32(i+1), float32(2), float32(3), float32(4)
			var operations int64
			ready.Done()
			<-start
			for time.Now().Before(until) {
				for range 1024 {
					a = a*0.9999 + 0.1
					b = b*0.9998 + 0.2
					c = c*0.9997 + 0.3
					d = d*0.9996 + 0.4
				}
				operations += 8 * 1024
			}
			if a+b+c+d <= 0 {
				operations = 0
			}
			work <- operations
		}()
	}
	ready.Wait()
	began := time.Now()
	until = began.Add(50 * time.Millisecond)
	close(start)
	var total int64
	for range count {
		total += <-work
	}
	return float64(total) / time.Since(began).Seconds()
}
