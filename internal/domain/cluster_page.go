// Describes navigable cluster detail pages and explicit worker commands.
// Each command belongs to its displayed revision and runs only on activation.
package domain

import "context"

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
	Title, Detail string
	Options       []ClusterOption
}

type PoolDefault struct{ Name, Priority string }

func DefaultPools() []PoolDefault {
	return []PoolDefault{{"Chat Pool", "High"}, {"Project Pool", "Medium"}, {"Comrade Pool", "Low"}, {"Research Pool", "Last"}}
}
