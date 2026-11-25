//go:build linux

// Reads Linux memory, mount and accelerator counters with bounded sampling.
// Device labels exclude user mount paths and model allocations never stand in for total VRAM.
package inventory

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func number(text string) (uint64, bool) {
	v, err := strconv.ParseUint(strings.TrimSpace(text), 10, 64)
	return v, err == nil
}
func cpu() (uint64, uint64) {
	data, _ := os.ReadFile("/proc/stat")
	line, _, _ := strings.Cut(string(data), "\n")
	f := strings.Fields(line)
	if len(f) < 6 {
		return 0, 0
	}
	var total, idle uint64
	for i, v := range f[1:min(9, len(f))] {
		ticks, _ := number(v)
		total += ticks
		if i == 3 || i == 4 {
			idle += ticks
		}
	}
	return total, idle
}
func resources(ctx context.Context) domain.NodeResources {
	r := domain.NodeResources{OS: runtime.GOOS, Arch: runtime.GOARCH, CPUs: runtime.NumCPU()}
	r.PhysicalCores = physicalCores()
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err == nil && len(strings.TrimSpace(string(boot))) == 36 {
		r.HostBootDigest = fmt.Sprintf("%x", sha256.Sum256(boot))
	}
	start, idle := cpu()
	data, _ := os.ReadFile("/proc/meminfo")
	mem := map[string]uint64{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[2] == "kB" {
			n, ok := number(fields[1])
			if ok {
				mem[fields[0]] = n * 1024
			}
		}
	}
	capacity := func(name, k, free string) *domain.Capacity {
		total, ok := mem[k]
		avail, found := mem[free]
		if !ok || !found || avail > total {
			return nil
		}
		return &domain.Capacity{Name: name, Total: total, Used: total - avail, Available: avail}
	}
	r.RAM = capacity("RAM", "MemTotal:", "MemAvailable:")
	r.Swap = capacity("Swap", "SwapTotal:", "SwapFree:")
	data, _ = os.ReadFile("/proc/self/mountinfo")
	seen := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		left, right, ok := strings.Cut(line, " - ")
		f, fs := strings.Fields(left), strings.Fields(right)
		if !ok || len(f) < 5 || len(fs) < 2 || seen[f[2]] || len(r.Disks) >= 64 {
			continue
		}
		if f[4] != "/" && !strings.HasPrefix(fs[1], "/dev/") {
			continue
		}
		mount := strings.NewReplacer(`\040`, " ", `\011`, "\t", `\134`, "\\").Replace(f[4])
		var st unix.Statfs_t
		if unix.Statfs(mount, &st) != nil || st.Blocks == 0 || st.Bfree > st.Blocks {
			continue
		}
		unit := uint64(st.Bsize)
		label := "Disk " + f[2]
		if mount == "/" {
			label += " (system)"
		}
		r.Disks = append(r.Disks, domain.Capacity{Name: label, Total: st.Blocks * unit, Used: (st.Blocks - st.Bfree) * unit, Available: st.Bavail * unit})
		seen[f[2]] = true
	}
	files, _ := filepath.Glob("/sys/class/drm/card*/device/mem_info_vram_total")
	for _, file := range files {
		totalBytes, _ := os.ReadFile(file)
		usedBytes, _ := os.ReadFile(filepath.Join(filepath.Dir(file), "mem_info_vram_used"))
		total, tok := number(string(totalBytes))
		used, uok := number(string(usedBytes))
		if tok && uok && total > 0 && used <= total {
			name := filepath.Base(filepath.Dir(filepath.Dir(file)))
			r.GPUs = append(r.GPUs, domain.Capacity{Name: name, Total: total, Used: used, Available: total - used})
		}
	}
	gpuCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	cmd := exec.CommandContext(gpuCtx, "nvidia-smi", "--query-gpu=index,memory.total,memory.used,memory.free", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			f := strings.Split(line, ",")
			if len(f) != 4 {
				continue
			}
			idx, iok := number(f[0])
			t, tok := number(f[1])
			u, uok := number(f[2])
			a, aok := number(f[3])
			if iok && tok && uok && aok && u <= t && a <= t {
				r.GPUs = append(r.GPUs, domain.Capacity{Name: fmt.Sprintf("NVIDIA GPU %d", idx), Total: t << 20, Used: u << 20, Available: a << 20})
			}
		}
	}
	select {
	case <-ctx.Done():
		return r
	case <-time.After(100 * time.Millisecond):
	}
	end, idleEnd := cpu()
	if end > start && idleEnd >= idle && idleEnd-idle <= end-start {
		v := 100 * (1 - float64(idleEnd-idle)/float64(end-start))
		r.CPUPercent = &v
	}
	return r
}

// Count distinct SMT sibling sets among CPUs available to the current process.
func physicalCores() int {
	var affinity unix.CPUSet
	if unix.SchedGetaffinity(0, &affinity) != nil {
		return 0
	}
	cores := map[string]bool{}
	for id := 0; affinity.Count() > 0; id++ {
		if !affinity.IsSet(id) {
			continue
		}
		affinity.Clear(id)
		b, err := os.ReadFile(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/topology/thread_siblings_list", id))
		key := strings.TrimSpace(string(b))
		if err != nil || key == "" {
			return 0
		}
		cores[key] = true
	}
	return len(cores)
}
