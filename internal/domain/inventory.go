// Defines the authenticated device inventory returned by headless workers.
// Unknown measurements are omitted; all capacity counters are in bytes.
package domain

import (
	"context"
	"errors"
	"io"
	"net"
	"time"
)

type Capacity struct {
	Name                   string
	Used, Available, Total uint64
}
type NodeResources struct {
	PhysicalCores     int     `json:",omitempty"`
	CPUFP32PerSecond  float64 `json:",omitempty"`
	GPUFP32PerSecond  float64 `json:",omitempty"`
	GPUComputeReports int     `json:",omitempty"`
	HostBootDigest    string  `json:",omitempty"`
	OS, Arch          string
	CPUs              int
	CPUPercent        *float64
	RAM, Swap         *Capacity
	Disks, GPUs       []Capacity
}
type InventoryModel struct {
	Name, Digest string
	Size         int64
	Loaded       bool
	Text         bool
}
type DeviceUsage struct {
	Tokens, Prompts uint64
}

type WorkerInventory struct {
	Usage            *DeviceUsage `json:",omitempty"`
	LastContact      int64        `json:"-"`
	Retrying         bool         `json:"-"`
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
	return !v.Retrying && v.Fresh(now) && v.ReportedText()
}

// ReportedText describes the last service advertisement, independently of reachability.
func (v WorkerInventory) ReportedText() bool {
	if v.Version == "" || v.Problem != "" {
		return false
	}
	for _, m := range v.Models {
		if m.Text {
			return true
		}
	}
	return false
}

// HardwareCapacity contains installed totals only. Live utilization and model routing
// continue to use timestamped WorkerInventory observations.
type HardwareCapacity struct {
	GPUs                        int `json:",omitempty"`
	Boot                        string
	Cores, Threads, UnratedGPUs int
	RAM, VRAM, Disk             uint64
	INT8                        float64
}

func (v WorkerInventory) Capacity() HardwareCapacity {
	r := v.Resources
	h := HardwareCapacity{GPUs: len(r.GPUs), Boot: r.HostBootDigest, Cores: r.PhysicalCores, Threads: r.CPUs,
		UnratedGPUs: max(0, len(r.GPUs)-r.GPUComputeReports), INT8: 4 * (r.CPUFP32PerSecond + r.GPUFP32PerSecond)}
	if r.RAM != nil {
		h.RAM = r.RAM.Total
	}
	for _, g := range r.GPUs {
		h.VRAM += g.Total
	}
	for _, d := range r.Disks {
		h.Disk += d.Total
	}
	return h
}

// ErrWorkerSessionExpired triggers a new login before retrying a read.
var ErrWorkerSessionExpired = errors.New("worker session expired")

// ContactInterrupted identifies transport failures on an authenticated worker read.
func ContactInterrupted(err error) bool {
	var op *net.OpError
	var timeout net.Error
	return errors.Is(err, ErrWorkerSessionExpired) || errors.As(err, &op) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || (errors.As(err, &timeout) && timeout.Timeout())
}
