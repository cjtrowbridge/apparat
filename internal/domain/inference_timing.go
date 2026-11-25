// Tracks applied-event propagation and full-roster claim rounds on the observer's clock.
// The adaptive TTL schedules inspections only. It never changes reservation ownership.
package domain

import (
	"slices"
	"sync"
	"time"
)

const timingWindow = 128
const initialRecheck = 3 * time.Second

type timingSeries struct {
	values      [timingWindow]time.Duration
	count, next int
}

func (s *timingSeries) add(d time.Duration) {
	s.values[s.next] = d
	s.next = (s.next + 1) % timingWindow
	s.count = min(s.count+1, timingWindow)
}

type TimingSummary struct {
	Count       int
	Latest, P99 time.Duration
}

func (s *timingSeries) summary() TimingSummary {
	if s.count == 0 {
		return TimingSummary{}
	}
	values := slices.Clone(s.values[:s.count])
	slices.Sort(values)
	return TimingSummary{s.count, s.values[(s.next+timingWindow-1)%timingWindow], values[(99*s.count+99)/100-1]}
}

type InferenceTiming struct {
	mu         sync.Mutex
	series     [2]timingSeries
	recheck    time.Duration
	incomplete uint64
}
type AgreementTiming struct {
	Confirmation, Propagation TimingSummary
	Recheck                   time.Duration
	Incomplete                uint64
}

func (t *InferenceTiming) Observe(propagation, complete bool, elapsed time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !complete {
		t.incomplete++
		return
	}
	if elapsed < 0 {
		return
	}
	i := 0
	if propagation {
		i = 1
	}
	t.series[i].add(elapsed)
	if t.series[0].count < 5 || t.series[1].count < 5 {
		return
	}
	target := max(250*time.Millisecond, 3*max(t.series[0].summary().P99, t.series[1].summary().P99))
	if t.recheck == 0 {
		t.recheck = initialRecheck
	}
	t.recheck = max(target, t.recheck-t.recheck/16)
}
func (t *InferenceTiming) Snapshot() AgreementTiming {
	t.mu.Lock()
	defer t.mu.Unlock()
	ttl := t.recheck
	if ttl == 0 {
		ttl = initialRecheck
	}
	return AgreementTiming{t.series[0].summary(), t.series[1].summary(), ttl, t.incomplete}
}
