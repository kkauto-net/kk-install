package validator

import (
	"os"
	"testing"
)

func writeTestFile(t *testing.T, path string, content []byte, perm os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, content, perm); err != nil {
		t.Fatalf("write test file %s: %v", path, err)
	}
}
