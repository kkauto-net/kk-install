package selfupdate

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestBinaryReplaceHintForNpmVendor(t *testing.T) {
	path := "/home/user/.hermes/node/lib/node_modules/@kkauto/kkcli/vendor/kk"
	if hint := binaryReplaceHint(path); hint == "" {
		t.Fatal("binaryReplaceHint() expected npm hint")
	}
}

func TestBinaryReplaceHintForStandaloneBinary(t *testing.T) {
	if hint := binaryReplaceHint("/usr/local/bin/kk"); hint != "" {
		t.Fatalf("binaryReplaceHint() = %q, want empty", hint)
	}
}

func TestMoveOrCopyFileUsesCopyOnCrossDeviceRename(t *testing.T) {
	oldRename := osRenameFn
	oldCopy := copyFileFn
	t.Cleanup(func() {
		osRenameFn = oldRename
		copyFileFn = oldCopy
	})

	dir := t.TempDir()
	dst := filepath.Join(dir, "kk")
	src := filepath.Join(dir, "kk-new")
	if err := os.WriteFile(src, []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	osRenameFn = func(oldpath, newpath string) error {
		return syscall.EXDEV
	}
	copied := false
	copyFileFn = func(dstPath, srcPath string) error {
		copied = true
		return copyFileAtomic(dstPath, srcPath)
	}

	if err := moveOrCopyFile(dst, src); err != nil {
		t.Fatalf("moveOrCopyFile() error = %v", err)
	}
	if !copied {
		t.Fatal("expected copy fallback on cross-device rename")
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "binary" {
		t.Fatalf("ReadFile() = %q, want binary", data)
	}
}
