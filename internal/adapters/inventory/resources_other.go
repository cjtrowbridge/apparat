//go:build !linux

// Reports portable CPU identity when native capacity counters are unavailable.
// Missing memory and accelerator measurements remain explicitly unknown.
package inventory

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"runtime"
)

func resources(context.Context) domain.NodeResources {
	return domain.NodeResources{OS: runtime.GOOS, Arch: runtime.GOARCH, CPUs: runtime.NumCPU()}
}
