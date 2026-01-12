package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	RepoOwner = "kkauto-net"
	RepoName  = "kk-install"
	Binary    = "kk"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateResult contains the result of an update check or operation
type UpdateResult struct {
	CurrentVersion string
	LatestVersion  string
	UpdateNeeded   bool
	DownloadURL    string
	AssetName      string
}

// CheckForUpdate checks if a newer version is available
func CheckForUpdate(ctx context.Context, currentVersion string) (*UpdateResult, error) {
	release, err := getLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	result := &UpdateResult{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
	}

	// Compare versions (strip 'v' prefix)
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(release.TagName, "v")

	result.UpdateNeeded = latest != current

	// Find the right asset for this platform
	assetName := getAssetName(release.TagName)
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			result.DownloadURL = asset.BrowserDownloadURL
			result.AssetName = asset.Name
			break
		}
	}

	if result.UpdateNeeded && result.DownloadURL == "" {
		return nil, fmt.Errorf("no compatible binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return result, nil
}

// Update downloads and installs the latest version
func Update(ctx context.Context, result *UpdateResult) error {
	if !result.UpdateNeeded {
		return nil
	}

	// Get current binary path
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	binaryPath, err = filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "kk-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the archive
	archivePath := filepath.Join(tmpDir, result.AssetName)
	if err := downloadFile(ctx, result.DownloadURL, archivePath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Extract the binary
	newBinaryPath := filepath.Join(tmpDir, Binary)
	if err := extractBinary(archivePath, newBinaryPath); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Make it executable
	if err := os.Chmod(newBinaryPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Replace the current binary
	if err := replaceBinary(binaryPath, newBinaryPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

func getLatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func getAssetName(version string) string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Version without 'v' prefix
	ver := strings.TrimPrefix(version, "v")

	return fmt.Sprintf("kkcli_%s_%s_%s.tar.gz", ver, os, arch)
}

func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for the binary
		if header.Name == Binary || filepath.Base(header.Name) == Binary {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, tr)
			return err
		}
	}

	return fmt.Errorf("binary not found in archive")
}

func replaceBinary(oldPath, newPath string) error {
	// Check if we have write permission
	dir := filepath.Dir(oldPath)
	if !isWritable(dir) {
		// Need sudo
		return replaceBinaryWithSudo(oldPath, newPath)
	}

	// Backup old binary
	backupPath := oldPath + ".old"
	if err := os.Rename(oldPath, backupPath); err != nil {
		return err
	}

	// Move new binary
	if err := os.Rename(newPath, oldPath); err != nil {
		// Restore backup
		os.Rename(backupPath, oldPath)
		return err
	}

	// Remove backup
	os.Remove(backupPath)
	return nil
}

func replaceBinaryWithSudo(oldPath, newPath string) error {
	// Print newline before sudo prompt for better formatting
	fmt.Println()

	// Use sudo to replace the binary
	cmd := exec.Command("sudo", "mv", newPath, oldPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo mv failed: %w", err)
	}

	// Set permissions
	cmd = exec.Command("sudo", "chmod", "755", oldPath)
	return cmd.Run()
}

func isWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if current user can write
	if info.Mode().Perm()&0200 != 0 {
		// Try to create a temp file
		tmpFile := filepath.Join(path, ".kk-write-test")
		f, err := os.Create(tmpFile)
		if err != nil {
			return false
		}
		f.Close()
		os.Remove(tmpFile)
		return true
	}

	return false
}
