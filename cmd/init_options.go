package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kkauto-net/kk-install/pkg/license"
)

const maxInitLicenseSourceBytes = 4096

type initOptions struct {
	NonInteractive bool
	Force          bool
	InstallDocker  bool
	License        string
	LicenseFile    string
	LicenseStdin   bool
	Domain         string
	Language       string
}

func collectInitOptions() initOptions {
	return initOptions{
		NonInteractive: yesInit,
		Force:          forceInit,
		InstallDocker:  installDockerFlag,
		License:        strings.TrimSpace(initLicense),
		LicenseFile:    strings.TrimSpace(initLicenseFile),
		LicenseStdin:   initLicenseStdin,
		Domain:         strings.TrimSpace(initDomain),
		Language:       strings.TrimSpace(initLanguage),
	}
}

func resolveInitLicenseSource(opts initOptions, stdin io.Reader) (initOptions, error) {
	if !opts.NonInteractive {
		return opts, nil
	}

	sourceCount := 0
	if opts.License != "" {
		sourceCount++
	}
	if opts.LicenseFile != "" {
		sourceCount++
	}
	if opts.LicenseStdin {
		sourceCount++
	}

	if sourceCount == 0 {
		return opts, NewExitError(exitCodeInputValidation, errors.New("one license source is required when --yes is set"))
	}
	if sourceCount > 1 {
		return opts, NewExitError(exitCodeInputValidation, errors.New("only one license source can be used when --yes is set"))
	}

	if opts.LicenseFile != "" {
		licenseKey, err := readInitLicenseFile(opts.LicenseFile)
		if err != nil {
			return opts, err
		}
		opts.License = licenseKey
	}

	if opts.LicenseStdin {
		licenseKey, err := readInitLicenseStdin(stdin)
		if err != nil {
			return opts, err
		}
		opts.License = licenseKey
	}

	return opts, nil
}

func readInitLicenseFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", NewExitError(exitCodeInputValidation, fmt.Errorf("cannot read --license-file: %w", err))
	}
	if !info.Mode().IsRegular() {
		return "", NewExitError(exitCodeInputValidation, errors.New("--license-file must be a regular file"))
	}
	if info.Size() > maxInitLicenseSourceBytes {
		return "", NewExitError(exitCodeInputValidation, errors.New("--license-file is too large"))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", NewExitError(exitCodeInputValidation, fmt.Errorf("cannot read --license-file: %w", err))
	}
	return normalizeInitLicenseSource("--license-file", string(data))
}

func readInitLicenseStdin(stdin io.Reader) (string, error) {
	if stdin == nil {
		return "", NewExitError(exitCodeInputValidation, errors.New("cannot read --license-stdin"))
	}
	if file, ok := stdin.(*os.File); ok {
		info, err := file.Stat()
		if err == nil && info.Mode()&os.ModeCharDevice != 0 {
			return "", NewExitError(exitCodeInputValidation, errors.New("--license-stdin requires piped or redirected input"))
		}
	}
	data, err := io.ReadAll(io.LimitReader(stdin, maxInitLicenseSourceBytes+1))
	if err != nil {
		return "", NewExitError(exitCodeInputValidation, fmt.Errorf("cannot read --license-stdin: %w", err))
	}
	if len(data) > maxInitLicenseSourceBytes {
		return "", NewExitError(exitCodeInputValidation, errors.New("--license-stdin is too large"))
	}
	return normalizeInitLicenseSource("--license-stdin", string(data))
}

func normalizeInitLicenseSource(source, value string) (string, error) {
	if len(value) > maxInitLicenseSourceBytes {
		return "", NewExitError(exitCodeInputValidation, fmt.Errorf("%s is too large", source))
	}
	licenseKey := strings.TrimSpace(value)
	if licenseKey == "" {
		return "", NewExitError(exitCodeInputValidation, fmt.Errorf("%s is empty", source))
	}
	return licenseKey, nil
}

func validateInitOptions(opts initOptions) error {
	if !opts.NonInteractive {
		return nil
	}

	if opts.License == "" {
		return NewExitError(exitCodeInputValidation, errors.New("resolved license is required when --yes is set"))
	}
	if opts.Domain == "" {
		return NewExitError(exitCodeInputValidation, errors.New("--domain is required when --yes is set"))
	}
	if opts.Language == "" {
		return NewExitError(exitCodeInputValidation, errors.New("--language is required when --yes is set"))
	}
	if !license.ValidateFormat(opts.License) {
		return NewExitError(exitCodeInputValidation, errors.New("--license has invalid format"))
	}
	if err := validateDomain(opts.Domain); err != nil {
		return NewExitError(exitCodeInputValidation, fmt.Errorf("--domain is invalid: %w", err))
	}
	if opts.Language != "en" && opts.Language != "vi" {
		return NewExitError(exitCodeInputValidation, errors.New("--language must be en or vi"))
	}

	return nil
}

func sanitizeLicenseError(message, licenseKey string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "license validation failed"
	}
	if licenseKey == "" {
		return message
	}
	return strings.ReplaceAll(message, licenseKey, maskLicense(licenseKey))
}

func maskLicense(licenseKey string) string {
	if len(licenseKey) <= 4 {
		return "****"
	}
	return "LICENSE-************" + licenseKey[len(licenseKey)-4:]
}
