package validator

import (
	"fmt"
	"syscall"

	"github.com/kkauto-net/kk-install/pkg/ui"
)

const MinDiskSpaceGB = 5

var statfsCaller = syscall.Statfs

// CheckDiskSpace verifies sufficient disk space
func CheckDiskSpace(path string) (float64, error) {
	var stat syscall.Statfs_t
	if err := statfsCaller(path, &stat); err != nil {
		return 0, fmt.Errorf("khong kiem tra duoc disk: %w", err)
	}

	// Available space in bytes
	available := float64(stat.Bavail * uint64(stat.Bsize))
	availableGB := available / (1024 * 1024 * 1024)

	return availableGB, nil
}

// WarnIfLowDiskSpace prints warning if disk < MinDiskSpaceGB
func WarnIfLowDiskSpace(path string) {
	availableGB, err := CheckDiskSpace(path)
	if err != nil {
		return // Silently ignore if can't check
	}

	if availableGB < MinDiskSpaceGB {
		ui.ShowWarningf(ui.Msg("warn_disk_low"), availableGB, MinDiskSpaceGB)
	}
}
