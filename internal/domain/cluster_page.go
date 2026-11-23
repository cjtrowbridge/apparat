// Describes navigable cluster detail pages and explicit worker commands.
// Each command belongs to its displayed revision and runs only on activation.
package domain

import "context"

type ClusterRoute struct {
	Worker, Page, Pool, Target string
}
type ClusterOption struct {
	Checked        *bool
	Label          string
	Route          ClusterRoute
	Input, Confirm string
	Run            func(context.Context, string) error
}
type ClusterPage struct {
	Title, Detail string
	Options       []ClusterOption
}

type PoolDefault struct{ Name, Priority string }

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
	if d.Worker {
		return workerHost
	}
	if d.Host != "" {
		return d.Host
	}
	return "App device " + d.ID[:min(8, len(d.ID))]
}
