package selfupdate

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseChecksumValidLine(t *testing.T) {
	t.Parallel()

	assetName := "kkcli_1.2.3_linux_amd64.tar.gz"
	checksum := strings.Repeat("a", 64)
	checksumsPath := writeTempFile(t, "checksums.txt", fmt.Sprintf("%s  %s\n", checksum, assetName))

	got, err := parseChecksum(checksumsPath, assetName)
	if err != nil {
		t.Fatalf("parseChecksum returned error: %v", err)
	}
	if got != checksum {
		t.Fatalf("parseChecksum = %q, want %q", got, checksum)
	}
}

func TestParseChecksumMissingEntry(t *testing.T) {
	t.Parallel()

	checksumsPath := writeTempFile(t, "checksums.txt", fmt.Sprintf("%s  other.tar.gz\n", strings.Repeat("a", 64)))

	if _, err := parseChecksum(checksumsPath, "kkcli_1.2.3_linux_amd64.tar.gz"); err == nil {
		t.Fatal("parseChecksum returned nil error for missing entry")
	}
}

func TestParseChecksumMalformedHash(t *testing.T) {
	t.Parallel()

	assetName := "kkcli_1.2.3_linux_amd64.tar.gz"
	checksumsPath := writeTempFile(t, "checksums.txt", fmt.Sprintf("not-a-sha256  %s\n", assetName))

	if _, err := parseChecksum(checksumsPath, assetName); err == nil {
		t.Fatal("parseChecksum returned nil error for malformed hash")
	}
}

func TestParseChecksumRequiresExactFilename(t *testing.T) {
	t.Parallel()

	assetName := "kkcli_1.2.3_linux_amd64.tar.gz"
	checksum := strings.Repeat("b", 64)
	checksumsPath := writeTempFile(t, "checksums.txt", fmt.Sprintf("%s  %s.old\n", strings.Repeat("a", 64), assetName)+fmt.Sprintf("%s  %s\n", checksum, assetName))

	got, err := parseChecksum(checksumsPath, assetName)
	if err != nil {
		t.Fatalf("parseChecksum returned error: %v", err)
	}
	if got != checksum {
		t.Fatalf("parseChecksum = %q, want exact filename checksum %q", got, checksum)
	}
}

func TestParseChecksumLowercasesUppercaseHash(t *testing.T) {
	t.Parallel()

	assetName := "kkcli_1.2.3_linux_amd64.tar.gz"
	checksum := strings.Repeat("A", 64)
	checksumsPath := writeTempFile(t, "checksums.txt", fmt.Sprintf("%s  %s\n", checksum, assetName))

	got, err := parseChecksum(checksumsPath, assetName)
	if err != nil {
		t.Fatalf("parseChecksum returned error: %v", err)
	}
	if got != strings.ToLower(checksum) {
		t.Fatalf("parseChecksum = %q, want lowercase checksum", got)
	}
}

func TestParseChecksumIgnoresExtraFieldsAfterFilename(t *testing.T) {
	t.Parallel()

	assetName := "kkcli_1.2.3_linux_amd64.tar.gz"
	checksum := strings.Repeat("c", 64)
	checksumsPath := writeTempFile(t, "checksums.txt", fmt.Sprintf("%s  %s  ignored-extra-field\n", checksum, assetName))

	got, err := parseChecksum(checksumsPath, assetName)
	if err != nil {
		t.Fatalf("parseChecksum returned error: %v", err)
	}
	if got != checksum {
		t.Fatalf("parseChecksum = %q, want %q", got, checksum)
	}
}

func TestVerifyChecksumValidFile(t *testing.T) {
	t.Parallel()

	assetName := "kkcli_1.2.3_linux_amd64.tar.gz"
	archivePath := writeTempFile(t, assetName, "release archive")
	checksum := sha256.Sum256([]byte("release archive"))
	checksumsPath := writeTempFile(t, "checksums.txt", fmt.Sprintf("%x  %s\n", checksum, assetName))

	if err := verifyChecksum(archivePath, checksumsPath, assetName); err != nil {
		t.Fatalf("verifyChecksum returned error: %v", err)
	}
}

func TestVerifyChecksumMismatch(t *testing.T) {
	t.Parallel()

	assetName := "kkcli_1.2.3_linux_amd64.tar.gz"
	archivePath := writeTempFile(t, assetName, "tampered archive")
	checksum := sha256.Sum256([]byte("release archive"))
	checksumsPath := writeTempFile(t, "checksums.txt", fmt.Sprintf("%x  %s\n", checksum, assetName))

	if err := verifyChecksum(archivePath, checksumsPath, assetName); err == nil {
		t.Fatal("verifyChecksum returned nil error for checksum mismatch")
	}
}

func writeTempFile(t *testing.T, name, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}
