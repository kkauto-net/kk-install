package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kkauto-net/kk-install/pkg/license"
)

type initOptions struct {
	NonInteractive bool
	Force          bool
	License        string
	Domain         string
	Language       string
}

func collectInitOptions() initOptions {
	return initOptions{
		NonInteractive: yesInit,
		Force:          forceInit,
		License:        strings.TrimSpace(initLicense),
		Domain:         strings.TrimSpace(initDomain),
		Language:       strings.TrimSpace(initLanguage),
	}
}

func validateInitOptions(opts initOptions) error {
	if !opts.NonInteractive {
		return nil
	}

	if opts.License == "" {
		return NewExitError(exitCodeInputValidation, errors.New("--license is required when --yes is set"))
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
