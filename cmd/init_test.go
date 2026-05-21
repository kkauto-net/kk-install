package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	oldLicense, oldDomain, oldLanguage := initLicense, initDomain, initLanguage
	t.Cleanup(func() {
		yesInit, forceInit = oldYes, oldForce
		initLicense, initDomain, initLanguage = oldLicense, oldDomain, oldLanguage
	})

	yesInit = true
	forceInit = true
	initLicense = " LICENSE-ABCDEF0123456789 "
	initDomain = " example.com "
	initLanguage = " en "

	opts := collectInitOptions()
	if !opts.NonInteractive || !opts.Force {
		t.Fatalf("collectInitOptions() booleans = %+v", opts)
	}
	if opts.License != "LICENSE-ABCDEF0123456789" || opts.Domain != "example.com" || opts.Language != "en" {
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
