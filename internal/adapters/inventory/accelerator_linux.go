//go:build linux

// Estimates AMD FP32 throughput using the reported CU count and maximum clock.
// Matches KFD and DRM by render minor so multiple adapters are counted independently.
package inventory

import (
	"context"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func amdComputeRates() map[string]float64 {
	rates := map[string]float64{}
	nodes, _ := filepath.Glob("/sys/class/kfd/kfd/topology/nodes/*/properties")
	for _, path := range nodes {
		b, _ := os.ReadFile(path)
		fields := map[string]uint64{}
		for _, line := range strings.Split(string(b), "\n") {
			k, v, _ := strings.Cut(line, " ")
			fields[k], _ = number(v)
		}
		simds := fields["simd_per_cu"]
		if simds == 0 || fields["vendor_id"] != 0x1002 {
			continue
		}
		cus := fields["simd_count"] / simds
		// One 64-lane FMA per CU; omit optional dual-issue uplift.
		rate := float64(cus) * 64 * 2 * float64(fields["max_engine_clk_fcompute"]) * 1e6
		rates["renderD"+strconv.FormatUint(fields["drm_render_minor"], 10)] = rate
	}
	return rates
}

func drmFP32(device string, rates map[string]float64) float64 {
	renders, _ := filepath.Glob(filepath.Join(device, "drm", "renderD*"))
	for _, r := range renders {
		v := rates[filepath.Base(r)]
		if v > 0 {
			return v
		}
	}
	id := func(field string) string {
		b, _ := os.ReadFile(filepath.Join(device, field))
		return strings.TrimSpace(string(b))
	}
	if id("vendor") == "0x1002" && id("device") == "0x6798" && id("subsystem_vendor") == "0x106b" {
		switch id("subsystem_device") {
		case "0x0127", "0x0128":
			return 3.5e12
		}
	}
	return 0
}

// Enumerate installed PCI display controllers even when their driver has no counters.
func readGPUs(ctx context.Context, r *domain.NodeResources) {
	rates := amdComputeRates()
	paths, _ := filepath.Glob("/sys/bus/pci/devices/*")
	byBus := map[string]int{}
	for _, device := range paths {
		class, _ := os.ReadFile(filepath.Join(device, "class"))
		if !strings.HasPrefix(string(class), "0x03") {
			continue
		}
		bus := filepath.Base(device)
		name := "GPU " + bus
		cards, _ := filepath.Glob(filepath.Join(device, "drm", "card[0-9]*"))
		if len(cards) > 0 {
			name = filepath.Base(cards[0])
		}
		capacity := domain.Capacity{Name: name}
		total, _ := os.ReadFile(filepath.Join(device, "mem_info_vram_total"))
		used, _ := os.ReadFile(filepath.Join(device, "mem_info_vram_used"))
		capacity.Total, _ = number(string(total))
		usedBytes, ok := number(string(used))
		if ok && usedBytes <= capacity.Total {
			capacity.Used = usedBytes
			capacity.Available = capacity.Total - usedBytes
		}
		rate := drmFP32(device, rates)
		if rate > 0 {
			r.GPUFP32PerSecond += rate
			r.GPUComputeReports++
		}
		byBus[bus] = len(r.GPUs)
		r.GPUs = append(r.GPUs, capacity)
	}
	drm, _ := filepath.Glob("/sys/class/drm/card[0-9]*/device")
	for _, path := range drm {
		device, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}
		bus := filepath.Base(device)
		_, known := byBus[bus]
		if known {
			continue
		}
		byBus[bus] = len(r.GPUs)
		r.GPUs = append(r.GPUs, domain.Capacity{Name: filepath.Base(filepath.Dir(path))})
	}
	gpuCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	output, err := exec.CommandContext(gpuCtx, "nvidia-smi", "--query-gpu=pci.bus_id,index,memory.total,memory.used,memory.free", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		f := strings.Split(line, ",")
		if len(f) != 5 {
			continue
		}
		bus := strings.ToLower(strings.TrimSpace(f[0]))
		if len(bus) == 16 && strings.HasPrefix(bus, "0000") {
			bus = bus[4:]
		}
		idx, ok := number(f[1])
		if !ok {
			continue
		}
		g := domain.Capacity{Name: fmt.Sprintf("NVIDIA GPU %d", idx)}
		t, tok := number(f[2])
		u, uok := number(f[3])
		a, aok := number(f[4])
		if tok && uok && aok && u <= t && a <= t {
			g.Total = t << 20
			g.Used = u << 20
			g.Available = a << 20
		}
		i, exists := byBus[bus]
		if exists {
			if g.Total > 0 || r.GPUs[i].Total == 0 {
				r.GPUs[i] = g
			}
		} else {
			r.GPUs = append(r.GPUs, g)
		}
	}
}
