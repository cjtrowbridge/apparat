// Describes navigable cluster detail pages and explicit worker commands.
// Each command belongs to its displayed revision and runs only on activation.
package domain

import (
	"context"
	"strings"
)

type ClusterRoute struct {
	Worker, Page, Pool, Target string
}
type ClusterOption struct {
	Label          string
	Route          ClusterRoute
	Input, Confirm string
	Run            func(context.Context, string) error
}
type ClusterPage struct {
	InitialValue  string
	Overview      *PoolOverview
	Title, Detail string
	Options       []ClusterOption
	Pools         []string
	Assigned      []string
	SavePools     func(context.Context, []string) error
}

type PoolDefault struct{ Name, Priority string }

func RetiredPool(name string) bool {
	name = strings.TrimSpace(name)
	return strings.EqualFold(name, "personal") || strings.EqualFold(name, "personal pool")
}

func DefaultPools() []PoolDefault {
	return []PoolDefault{{"Chat Pool", "High"}, {"Project Pool", "Medium"}, {"Comrade Pool", "Low"}, {"Research Pool", "Last"}}
}

// PoolDevice is a worker's authorized roster entry. Empty OS denotes a legacy GUI.
type PoolDevice struct {
	ID, Host, OS string
	Worker       bool
	Pools        []string
}

func (d PoolDevice) DisplayName(workerHost string) string {
	name := SystemDeviceName(d.Host)
	if name != "" {
		return name
	}
	if d.Worker {
		name := SystemDeviceName(workerHost)
		if name != "" {
			return name
		}
		return "Worker " + d.ID[:min(8, len(d.ID))]
	}
	return "App device " + d.ID[:min(8, len(d.ID))]
}

// PoolOverview is the signed-in user's last synchronized view across the selected workers.
type PoolOverview struct {
	Tasks                     []PoolTaskRow
	Devices                   []PoolDeviceRow
	Add, Policy               []ClusterOption
	GlobalModel, ProjectModel ModelRef
	Users                     []string
}
type PoolDeviceRow struct {
	ClusterOption
	Speed, ProjectSpeed ModelSpeed
}

// Each rate is absent until the worker has reported a timed sample for that phase.
type ModelSpeed struct {
	Prefill, Eval *float64
}

type PoolTaskRow struct {
	Inference               bool
	ID, State, Worker, Host string
	Changed                 int64
}

// QueueHealth is a timestamped, payload-free observation of this worker's hash queue.
type QueueHealth struct {
	ObservedMS                                         int64
	Queued, Eligible, Running, Expired                 int
	OldestEligibleMS                                   int64
	Draining                                           bool
	OllamaConfigured, OllamaFresh, OllamaTextAvailable bool
}
