// Verifies timing changes only the inspection interval and delayed attempts cannot steal work.
// Tests deliberately reordered messages, a cold validator and a long propagation tail.
package domain

import (
	"testing"
	"time"
)

func TestInferenceRecheck(t *testing.T) {
	var timing InferenceTiming
	if timing.Snapshot().Recheck != 3*time.Second {
		t.Fatal("unexpected cold recheck interval")
	}
	for range 5 {
		timing.Observe(false, true, time.Millisecond)
		timing.Observe(true, true, 4*time.Second)
	}
	v := timing.Snapshot()
	if v.Recheck != 12*time.Second || v.Propagation.P99 != 4*time.Second {
		t.Fatal(v)
	}
	timing.Observe(true, false, 20*time.Second)
	v = timing.Snapshot()
	if v.Incomplete != 1 || v.Propagation.Count != 5 || v.Recheck != 12*time.Second {
		t.Fatal(v)
	}
	for range 256 {
		timing.Observe(true, true, time.Millisecond)
	}
	if timing.Snapshot().Recheck != 250*time.Millisecond {
		t.Fatal(timing.Snapshot())
	}
}
func TestInferenceReservationReplay(t *testing.T) {
	id, err := NewID(time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	table, err := NewClaimTable("r", []string{"first", "second"})
	if err != nil {
		t.Fatal(err)
	}
	err = table.Admit(id, "digest")
	if err != nil {
		t.Fatal(err)
	}
	c := InferenceClaim{Revision: "r", Job: id, Digest: "digest", Executor: "first", Attempt: id}
	other := c
	other.Executor = "second"
	if table.Transition(c.Executor, "confirm", c).State != "recovering" {
		t.Fatal("voted before recovery")
	}
	table.Recovered()
	table.Transition(c.Executor, "withdraw", c)
	if table.Transition(c.Executor, "confirm", c).State != "withdrawn" {
		t.Fatal("revived withdrawn attempt")
	}
	table.Transition(other.Executor, "confirm", other)
	table.Transition(other.Executor, "own", other)
	if table.Transition(other.Executor, "withdraw", other).State != "owned" {
		t.Fatal("withdrawn durable owner")
	}
	if table.Transition("first", "finish", other).State != "invalid" {
		t.Fatal("unrelated caller finished job")
	}
}
