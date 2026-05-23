package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kkauto-net/kk-install/pkg/license"
	"github.com/kkauto-net/kk-install/pkg/templates"
	"github.com/kkauto-net/kk-install/pkg/validator"
	"github.com/spf13/cobra"
)

func TestValidateInitOptions(t *testing.T) {
	valid := initOptions{
		NonInteractive: true,
		License:        "LICENSE-ABCDEF0123456789",
		Domain:         "example.com",
		Language:       "en",
	}

	tests := []struct {
		name     string
		opts     initOptions
		wantCode int
	}{
		{name: "interactive mode ignores missing unattended flags", opts: initOptions{}, wantCode: 0},
		{name: "valid english", opts: valid, wantCode: 0},
		{name: "valid vietnamese", opts: initOptions{NonInteractive: true, License: valid.License, Domain: valid.Domain, Language: "vi"}, wantCode: 0},
		{name: "missing license", opts: initOptions{NonInteractive: true, Domain: valid.Domain, Language: valid.Language}, wantCode: exitCodeInputValidation},
		{name: "missing domain", opts: initOptions{NonInteractive: true, License: valid.License, Language: valid.Language}, wantCode: exitCodeInputValidation},
		{name: "missing language", opts: initOptions{NonInteractive: true, License: valid.License, Domain: valid.Domain}, wantCode: exitCodeInputValidation},
		{name: "invalid license", opts: initOptions{NonInteractive: true, License: "bad-license", Domain: valid.Domain, Language: valid.Language}, wantCode: exitCodeInputValidation},
		{name: "invalid domain", opts: initOptions{NonInteractive: true, License: valid.License, Domain: "bad_domain", Language: valid.Language}, wantCode: exitCodeInputValidation},
		{name: "invalid language", opts: initOptions{NonInteractive: true, License: valid.License, Domain: valid.Domain, Language: "fr"}, wantCode: exitCodeInputValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInitOptions(tt.opts)
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("validateInitOptions() error = %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("validateInitOptions() expected error")
			}
			if got := ExitCode(err); got != tt.wantCode {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestValidateInitOptionsDoesNotExposeLicense(t *testing.T) {
	licenseKey := "LICENSE-ABCDEF0123456789"
	err := validateInitOptions(initOptions{
		NonInteractive: true,
		License:        licenseKey,
		Domain:         "bad_domain",
		Language:       "en",
	})
	if err == nil {
		t.Fatal("validateInitOptions() expected error")
	}
	if strings.Contains(err.Error(), licenseKey) {
		t.Fatalf("validation error exposed full license: %q", err.Error())
	}
}

func TestResolveInitLicenseSource(t *testing.T) {
	licenseKey := "LICENSE-ABCDEF0123456789"
	dir := t.TempDir()
	licensePath := filepath.Join(dir, "license.tmp")
	if err := os.WriteFile(licensePath, []byte(" "+licenseKey+"\n"), 0600); err != nil {
		t.Fatalf("write license fixture: %v", err)
	}

	tests := []struct {
		name        string
		opts        initOptions
		stdin       string
		wantLicense string
		wantCode    int
	}{
		{
			name: "interactive mode ignores sources",
			opts: initOptions{LicenseFile: filepath.Join(dir, "missing")},
		},
		{
			name:     "missing source",
			opts:     initOptions{NonInteractive: true},
			wantCode: exitCodeInputValidation,
		},
		{
			name:     "multiple sources",
			opts:     initOptions{NonInteractive: true, License: licenseKey, LicenseFile: licensePath},
			wantCode: exitCodeInputValidation,
		},
		{
			name:        "legacy argv source",
			opts:        initOptions{NonInteractive: true, License: licenseKey},
			wantLicense: licenseKey,
		},
		{
			name:        "file source trims whitespace",
			opts:        initOptions{NonInteractive: true, LicenseFile: licensePath},
			wantLicense: licenseKey,
		},
		{
			name:        "stdin source trims whitespace",
			opts:        initOptions{NonInteractive: true, LicenseStdin: true},
			stdin:       "\n" + licenseKey + "\n",
			wantLicense: licenseKey,
		},
		{
			name:     "empty stdin source",
			opts:     initOptions{NonInteractive: true, LicenseStdin: true},
			stdin:    "\n",
			wantCode: exitCodeInputValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveInitLicenseSource(tt.opts, strings.NewReader(tt.stdin))
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("resolveInitLicenseSource() error = %v", err)
				}
				if got.License != tt.wantLicense {
					t.Fatalf("resolved license mismatch")
				}
				return
			}

			if err == nil {
				t.Fatal("resolveInitLicenseSource() expected error")
			}
			if code := ExitCode(err); code != tt.wantCode {
				t.Fatalf("ExitCode() = %d, want %d", code, tt.wantCode)
			}
			if strings.Contains(err.Error(), licenseKey) {
				t.Fatalf("source error exposed full license: %q", err.Error())
			}
		})
	}
}

func TestResolveInitLicenseFileErrors(t *testing.T) {
	licenseKey := "LICENSE-ABCDEF0123456789"
	dir := t.TempDir()
	emptyPath := filepath.Join(dir, "empty.tmp")
	badPath := filepath.Join(dir, "bad.tmp")
	if err := os.WriteFile(emptyPath, []byte("\n"), 0600); err != nil {
		t.Fatalf("write empty fixture: %v", err)
	}
	if err := os.WriteFile(badPath, []byte(licenseKey), 0600); err != nil {
		t.Fatalf("write bad fixture: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "missing file", path: filepath.Join(dir, "missing.tmp")},
		{name: "non regular file", path: dir},
		{name: "empty file", path: emptyPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveInitLicenseSource(initOptions{NonInteractive: true, LicenseFile: tt.path}, strings.NewReader(""))
			if err == nil {
				t.Fatal("resolveInitLicenseSource() expected error")
			}
			if code := ExitCode(err); code != exitCodeInputValidation {
				t.Fatalf("ExitCode() = %d, want %d", code, exitCodeInputValidation)
			}
			if strings.Contains(err.Error(), licenseKey) {
				t.Fatalf("file source error exposed full license: %q", err.Error())
			}
		})
	}

	unreadablePath := filepath.Join(dir, "unreadable.tmp")
	if err := os.WriteFile(unreadablePath, []byte(licenseKey), 0600); err != nil {
		t.Fatalf("write unreadable fixture: %v", err)
	}
	if err := os.Chmod(unreadablePath, 0000); err != nil {
		t.Fatalf("chmod unreadable fixture: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(unreadablePath, 0600); err != nil {
			t.Logf("restore unreadable fixture permissions: %v", err)
		}
	})
	_, err := resolveInitLicenseSource(initOptions{NonInteractive: true, LicenseFile: unreadablePath}, strings.NewReader(""))
	if err == nil {
		t.Skip("current user can read chmod 0000 files; skipping unreadable-file assertion")
	}
	if code := ExitCode(err); code != exitCodeInputValidation {
		t.Fatalf("ExitCode() = %d, want %d", code, exitCodeInputValidation)
	}
}

func TestResolveInitLicenseSourceRejectsLargeInputs(t *testing.T) {
	dir := t.TempDir()
	largeValue := strings.Repeat("A", maxInitLicenseSourceBytes+1)
	largePath := filepath.Join(dir, "large.tmp")
	if err := os.WriteFile(largePath, []byte(largeValue), 0600); err != nil {
		t.Fatalf("write large fixture: %v", err)
	}

	tests := []struct {
		name  string
		opts  initOptions
		stdin string
	}{
		{name: "large file", opts: initOptions{NonInteractive: true, LicenseFile: largePath}},
		{name: "large stdin", opts: initOptions{NonInteractive: true, LicenseStdin: true}, stdin: largeValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveInitLicenseSource(tt.opts, strings.NewReader(tt.stdin))
			if err == nil {
				t.Fatal("resolveInitLicenseSource() expected error")
			}
			if code := ExitCode(err); code != exitCodeInputValidation {
				t.Fatalf("ExitCode() = %d, want %d", code, exitCodeInputValidation)
			}
		})
	}
}

func TestLicenseFileInvalidFormatUsesInputExitCode(t *testing.T) {
	dir := t.TempDir()
	licensePath := filepath.Join(dir, "license.tmp")
	if err := os.WriteFile(licensePath, []byte("bad-license\n"), 0600); err != nil {
		t.Fatalf("write invalid fixture: %v", err)
	}

	opts, err := resolveInitLicenseSource(initOptions{
		NonInteractive: true,
		LicenseFile:    licensePath,
		Domain:         "example.com",
		Language:       "en",
	}, strings.NewReader(""))
	if err != nil {
		t.Fatalf("resolveInitLicenseSource() error = %v", err)
	}

	err = validateInitOptions(opts)
	if err == nil {
		t.Fatal("validateInitOptions() expected error")
	}
	if code := ExitCode(err); code != exitCodeInputValidation {
		t.Fatalf("ExitCode() = %d, want %d", code, exitCodeInputValidation)
	}
	if strings.Contains(err.Error(), "bad-license") {
		t.Fatalf("validation error exposed file content: %q", err.Error())
	}
}

func TestSanitizeLicenseErrorMasksLicense(t *testing.T) {
	licenseKey := "LICENSE-ABCDEF0123456789"
	message := sanitizeLicenseError("license LICENSE-ABCDEF0123456789 is invalid", licenseKey)
	if strings.Contains(message, licenseKey) {
		t.Fatalf("sanitizeLicenseError() exposed full license: %q", message)
	}
	if !strings.Contains(message, "LICENSE-************6789") {
		t.Fatalf("sanitizeLicenseError() = %q, want masked license", message)
	}
}

func TestBackupExistingConfigsSecuresEnvBackup(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("LICENSE_KEY=LICENSE-ABCDEF0123456789\n"), 0600); err != nil {
		t.Fatalf("write env fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("services: {}\n"), 0644); err != nil {
		t.Fatalf("write compose fixture: %v", err)
	}

	if err := backupExistingConfigs(dir); err != nil {
		t.Fatalf("backupExistingConfigs() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	var backupEnv string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "backup-") {
			backupEnv = filepath.Join(dir, entry.Name(), ".env")
			break
		}
	}
	if backupEnv == "" {
		t.Fatal("backup .env not found")
	}

	info, err := os.Stat(backupEnv)
	if err != nil {
		t.Fatalf("stat backup .env: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("backup .env mode = %o, want 0600", got)
	}
}

func TestCollectInitOptionsTrimsFlags(t *testing.T) {
	oldYes, oldForce := yesInit, forceInit
	oldLicense, oldLicenseFile, oldLicenseStdin := initLicense, initLicenseFile, initLicenseStdin
	oldDomain, oldLanguage := initDomain, initLanguage
	t.Cleanup(func() {
		yesInit, forceInit = oldYes, oldForce
		initLicense, initLicenseFile, initLicenseStdin = oldLicense, oldLicenseFile, oldLicenseStdin
		initDomain, initLanguage = oldDomain, oldLanguage
	})

	yesInit = true
	forceInit = true
	initLicense = " LICENSE-ABCDEF0123456789 "
	initLicenseFile = " /tmp/license.tmp "
	initLicenseStdin = true
	initDomain = " example.com "
	initLanguage = " en "

	opts := collectInitOptions()
	if !opts.NonInteractive || !opts.Force {
		t.Fatalf("collectInitOptions() booleans = %+v", opts)
	}
	if opts.License != "LICENSE-ABCDEF0123456789" || opts.LicenseFile != "/tmp/license.tmp" || !opts.LicenseStdin || opts.Domain != "example.com" || opts.Language != "en" {
		t.Fatalf("collectInitOptions() did not trim strings: %+v", opts)
	}
}

func TestExitCode(t *testing.T) {
	if got := ExitCode(errors.New("legacy")); got != exitCodeLegacy {
		t.Fatalf("ExitCode(legacy) = %d, want %d", got, exitCodeLegacy)
	}

	err := NewExitError(exitCodeDockerValidation, errors.New("docker failed"))
	if got := ExitCode(err); got != exitCodeDockerValidation {
		t.Fatalf("ExitCode(typed) = %d, want %d", got, exitCodeDockerValidation)
	}
}

func TestNewExitErrorNil(t *testing.T) {
	if err := NewExitError(exitCodeInputValidation, nil); err != nil {
		t.Fatalf("NewExitError(nil) = %v, want nil", err)
	}
}

func TestRunInitUnattendedExitCodeContracts(t *testing.T) {
	licenseKey := "LICENSE-ABCDEF0123456789"

	tests := []struct {
		name      string
		configure func(t *testing.T)
		wantCode  int
	}{
		{
			name: "input validation failure",
			configure: func(t *testing.T) {
				initLicense = ""
			},
			wantCode: exitCodeInputValidation,
		},
		{
			name: "license validation failure",
			configure: func(t *testing.T) {
				newLicenseClient = func() *license.LicenseClient {
					return &license.LicenseClient{
						BaseURL: "http://127.0.0.1",
						HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
							return nil, fmt.Errorf("license %s rejected", licenseKey)
						})},
					}
				}
			},
			wantCode: exitCodeLicenseValidation,
		},
		{
			name: "docker validation failure",
			configure: func(t *testing.T) {
				DockerValidatorInstance = &validator.DockerValidator{
					LookPath: func(string) (string, error) { return "", os.ErrNotExist },
				}
			},
			wantCode: exitCodeDockerValidation,
		},
		{
			name: "render failure",
			configure: func(t *testing.T) {
				DockerValidatorInstance = successfulDockerValidator(t)
				renderTemplates = func(templates.Config, string) error {
					return errors.New("render failed")
				}
			},
			wantCode: exitCodeRenderFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetInitTestGlobals(t)
			yesInit = true
			initLicense = licenseKey
			initDomain = "example.com"
			initLanguage = "en"
			newLicenseClient = successfulLicenseClient
			DockerValidatorInstance = successfulDockerValidator(t)
			renderTemplates = func(templates.Config, string) error { return nil }
			startInitSpinner = func(string) initSpinner { return noopInitSpinner{} }
			tt.configure(t)

			cwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Getwd() error = %v", err)
			}
			tmp := t.TempDir()
			if chdirErr := os.Chdir(tmp); chdirErr != nil {
				t.Fatalf("Chdir() error = %v", chdirErr)
			}
			t.Cleanup(func() {
				if chdirErr := os.Chdir(cwd); chdirErr != nil {
					t.Logf("restore working directory: %v", chdirErr)
				}
			})
			t.Setenv("HOME", t.TempDir())

			err = runInit(&cobra.Command{}, nil)
			if err == nil {
				t.Fatal("runInit() expected error")
			}
			if got := ExitCode(err); got != tt.wantCode {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.wantCode)
			}
			if strings.Contains(err.Error(), licenseKey) {
				t.Fatalf("runInit() error exposed full license: %q", err.Error())
			}
		})
	}
}

func resetInitTestGlobals(t *testing.T) {
	oldYes, oldForce := yesInit, forceInit
	oldLicense, oldLicenseFile, oldLicenseStdin := initLicense, initLicenseFile, initLicenseStdin
	oldDomain, oldLanguage := initDomain, initLanguage
	oldDockerValidator := DockerValidatorInstance
	oldNewLicenseClient := newLicenseClient
	oldRenderTemplates := renderTemplates
	oldStartInitSpinner := startInitSpinner
	t.Cleanup(func() {
		yesInit, forceInit = oldYes, oldForce
		initLicense, initLicenseFile, initLicenseStdin = oldLicense, oldLicenseFile, oldLicenseStdin
		initDomain, initLanguage = oldDomain, oldLanguage
		DockerValidatorInstance = oldDockerValidator
		newLicenseClient = oldNewLicenseClient
		renderTemplates = oldRenderTemplates
		startInitSpinner = oldStartInitSpinner
	})
}

type noopInitSpinner struct{}

func (noopInitSpinner) Fail(message ...any) {}

func (noopInitSpinner) Success(message ...any) {}

func successfulLicenseClient() *license.LicenseClient {
	return &license.LicenseClient{
		BaseURL: "http://127.0.0.1",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success","public_key":"TEST-PUBLIC-KEY"}`)),
			}, nil
		})},
	}
}

func successfulDockerValidator(t *testing.T) *validator.DockerValidator {
	t.Helper()
	return &validator.DockerValidator{
		LookPath: func(string) (string, error) { return "/usr/bin/docker", nil },
		CommandContext: func(context.Context, string, ...string) *exec.Cmd {
			return exec.Command("true")
		},
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
