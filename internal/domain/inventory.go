// Defines the authenticated device inventory returned by headless workers.
// Unknown measurements are omitted; all capacity counters are in bytes.
package domain

import "time"

type Capacity struct {
	Name                   string
	Used, Available, Total uint64
}
type NodeResources struct {
	OS, Arch    string
	CPUs        int
	CPUPercent  *float64
	RAM, Swap   *Capacity
	Disks, GPUs []Capacity
}
type InventoryModel struct {
	Name   string
	Size   int64
	Loaded bool
	Text   bool
}
type WorkerInventory struct {
	Configured       bool
	Observed         int64
	Resources        NodeResources
	Port             int
	Version, Problem string
	Models           []InventoryModel
}

func (v WorkerInventory) Fresh(now time.Time) bool {
	age := now.Unix() - v.Observed
	return v.Observed > 0 && age >= -30 && age <= 120
}
func (v WorkerInventory) HasText(now time.Time) bool {
	if !v.Fresh(now) || v.Version == "" || v.Problem != "" {
		return false
	}
	for _, m := range v.Models {
		if m.Text {
			return true
		}
	}
	return false
}
