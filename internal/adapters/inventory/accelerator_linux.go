//go:build linux

// Estimates AMD FP32 throughput using the reported CU count and maximum clock.
// Matches KFD and DRM by render minor so multiple adapters are counted independently.
package inventory

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
