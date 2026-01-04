package validator

import (
	"syscall"
	"testing"
)

func TestCheckDiskSpace(t *testing.T) {
	t.Run("Check current directory", func(t *testing.T) {
		availableGB, err := CheckDiskSpace(".")
		if err != nil {
			t.Errorf("CheckDiskSpace failed: %v", err)
		}
		if availableGB < 0 {
			t.Errorf("Expected positive disk space, got %f", availableGB)
		}
	})

	t.Run("Invalid path", func(t *testing.T) {
		_, err := CheckDiskSpace("/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Error("Expected error for invalid path")
		}
	})

	t.Run("Mock low disk space", func(t *testing.T) {
		originalStatfs := statfsCaller
		defer func() { statfsCaller = originalStatfs }()

		statfsCaller = func(path string, stat *syscall.Statfs_t) error {
			stat.Bavail = 512 * 1024
			stat.Bsize = 4096
			return nil
		}

		availableGB, err := CheckDiskSpace(".")
		if err != nil {
			t.Errorf("CheckDiskSpace failed: %v", err)
		}
		if availableGB > 5 {
			t.Errorf("Expected low disk space (< 5GB), got %.1fGB", availableGB)
		}
	})
}

func TestWarnIfLowDiskSpace(t *testing.T) {
	t.Run("Low disk space", func(t *testing.T) {
		originalStatfs := statfsCaller
		defer func() { statfsCaller = originalStatfs }()

		statfsCaller = func(path string, stat *syscall.Statfs_t) error {
			stat.Bavail = 512 * 1024
			stat.Bsize = 4096
			return nil
		}

		WarnIfLowDiskSpace(".")
	})
}
