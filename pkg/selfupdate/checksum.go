package selfupdate

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

const checksumAssetName = "checksums.txt"

func parseChecksum(checksumsPath, assetName string) (string, error) {
	file, err := os.Open(checksumsPath)
	if err != nil {
		return "", fmt.Errorf("open checksum file: %w", err)
	}
	defer closeReader(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || fields[1] != assetName {
			continue
		}

		checksum := fields[0]
		if !isSHA256Hex(checksum) {
			return "", fmt.Errorf("invalid checksum for %s", assetName)
		}
		return strings.ToLower(checksum), nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read checksum file: %w", err)
	}

	return "", fmt.Errorf("checksum entry not found for %s", assetName)
}

func verifyChecksum(archivePath, checksumsPath, assetName string) error {
	expected, err := parseChecksum(checksumsPath, assetName)
	if err != nil {
		return err
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer closeReader(file)

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash archive: %w", err)
	}

	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}

	return nil
}

func isSHA256Hex(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
