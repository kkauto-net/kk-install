package cmd

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/config"
	"github.com/kkauto-net/kk-install/pkg/license"
	"github.com/kkauto-net/kk-install/pkg/templates"
	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/validator"
)

var initCmd = &cobra.Command{
	Use:         "init",
	Short:       "Initialize Docker stack with interactive setup",
	Long:        `Create docker-compose.yml and required config files.`,
	Annotations: map[string]string{"group": "core"},
	RunE:        runInit,
}

var (
	forceInit               bool
	yesInit                 bool
	installDockerFlag       bool
	initLicense             string
	initLicenseFile         string
	initLicenseStdin        bool
	initDomain              string
	initLanguage            string
	DockerValidatorInstance *validator.DockerValidator
	newLicenseClient        = license.NewClient
	renderTemplates         = templates.RenderAll
	domainRegex             = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z]{2,}$`)
	startInitSpinner        = func(text string) initSpinner {
		return ui.StartPtermSpinner(text)
	}
)

type initSpinner interface {
	Fail(message ...any)
	Success(message ...any)
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Bỏ qua tất cả các lời nhắc tương tác và sử dụng các giá trị mặc định")
	initCmd.Flags().BoolVar(&yesInit, "yes", false, "Run init without interactive prompts")
	initCmd.Flags().BoolVar(&installDockerFlag, "install-docker", false, "Auto-install/start Docker during unattended init")
	initCmd.Flags().StringVar(&initLicense, "license", "", "License key for unattended init (discouraged for automation; prefer --license-file)")
	initCmd.Flags().StringVar(&initLicenseFile, "license-file", "", "Read license key from file for unattended init")
	initCmd.Flags().BoolVar(&initLicenseStdin, "license-stdin", false, "Read license key from stdin for unattended init")
	initCmd.Flags().StringVar(&initDomain, "domain", "", "Domain for unattended init")
	initCmd.Flags().StringVar(&initLanguage, "language", "", "Language for unattended init (en or vi)")
	DockerValidatorInstance = validator.NewDockerValidator()
}

func runInit(cmd *cobra.Command, args []string) error {
	var spinner initSpinner
	opts := collectInitOptions()
	var err error
	opts, err = resolveInitLicenseSource(opts, cmd.InOrStdin())
	if err != nil {
		showInitInputError(err)
		return err
	}
	err = validateInitOptions(opts)
	if err != nil {
		showInitInputError(err)
		return err
	}
	// Command banner
	ui.ShowCommandBanner("kk init", ui.Msg("init_desc"))

	// Get working directory early to load existing env
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Load existing .env for pre-filling form values
	existingEnv := loadExistingEnv(cwd)
	hasExistingEnv := len(existingEnv) > 0
	if hasExistingEnv {
		ui.ShowInfo(ui.Msg("loading_existing_env"))
	}

	// Step 0: License Verification
	ui.ShowStepHeader(1, 7, ui.Msg("step_license"))

	var licenseData struct {
		Key       string
		PublicKey string
	}

	reexecLicense, reexecPublicKey, reexecOK := consumeReexecLicenseEnv()

	// Pre-fill license from existing env
	licenseKey := existingEnv["LICENSE_KEY"]
	if reexecOK {
		licenseKey = reexecLicense
		ui.ShowInfo(ui.IconKey + " " + ui.Msg("license_already_validated"))
		licenseData.Key = reexecLicense
		licenseData.PublicKey = reexecPublicKey
	} else if opts.NonInteractive {
		licenseKey = opts.License
	} else {
		licenseForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(ui.IconKey + " " + ui.Msg("enter_license")).
					Value(&licenseKey).
					Placeholder("LICENSE-XXXXXXXXXXXXXXXX").
					Validate(func(s string) error {
						if s == "" {
							return errors.New(ui.Msg("license_required"))
						}
						if !license.ValidateFormat(s) {
							return errors.New(ui.Msg("license_invalid_format"))
						}
						return nil
					}),
			),
		)
		if formErr := licenseForm.Run(); formErr != nil {
			return formErr
		}
	}

	// Skip license validation in test environment
	if reexecOK {
		// License was validated before docker group re-exec.
	} else if os.Getenv("KK_TEST_SKIP_LICENSE_VALIDATION") == "true" {
		ui.ShowWarning(ui.Msg("warn_skipping_license"))
		licenseData.Key = licenseKey
		licenseData.PublicKey = "TEST-PUBLIC-KEY"
	} else {
		// Validate license against API
		spinner = startInitSpinner(ui.IconKey + " " + ui.Msg("validating_license"))
		client := newLicenseClient()
		licenseResp, validateErr := client.Validate(licenseKey)
		if validateErr != nil {
			safeMessage := sanitizeLicenseError(validateErr.Error(), licenseKey)
			spinner.Fail(ui.Msg("license_validation_failed"))
			ui.ShowBoxedError(ui.ErrorSuggestion{
				Title:      ui.Msg("license_validation_failed"),
				Message:    safeMessage,
				Suggestion: ui.Msg("license_check_key"),
			})
			return NewExitError(exitCodeLicenseValidation, errors.New(safeMessage))
		}
		spinner.Success(ui.IconCheck + " " + ui.Msg("license_validated"))

		// Store license data for later use
		licenseData.Key = licenseKey
		licenseData.PublicKey = licenseResp.PublicKey
	}

	// Step 1: Check Docker
	ui.ShowStepHeader(2, 7, ui.Msg("step_docker_check"))
	ui.ShowInfo(ui.IconDocker + " " + ui.MsgCheckingDocker())

	if err = ensureInitDocker(opts, licenseData.Key, licenseData.PublicKey); err != nil {
		return err
	}

	ui.ShowSuccess(ui.IconCheck + " " + ui.MsgDockerOK())

	// Step 2: Language selection
	ui.ShowStepHeader(3, 7, ui.Msg("step_language"))
	var langChoice string
	if opts.NonInteractive {
		langChoice = opts.Language
	} else if opts.Force {
		langChoice = "en" // Default to English in force mode
	} else {
		langForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(ui.IconLanguage+" "+ui.Msg("select_language")).
					Options(
						huh.NewOption(ui.Msg("lang_english"), "en"),
						huh.NewOption(ui.Msg("lang_vietnamese"), "vi"),
					).
					Value(&langChoice),
			),
		)
		if formErr := langForm.Run(); formErr != nil {
			return formErr
		}
		// Set default to English if no selection
		if langChoice == "" {
			langChoice = "en"
		}
	}
	ui.SetLanguage(ui.Language(langChoice))

	// Save language preference to config
	cfg, err := config.Load()
	if err != nil {
		ui.ShowWarningf(ui.Msg("warn_cannot_load_config"), err)
		cfg = &config.Config{Language: langChoice}
	} else {
		cfg.Language = langChoice
	}
	if saveErr := cfg.Save(); saveErr != nil {
		ui.ShowWarning(fmt.Sprintf("Cannot save config: %v", saveErr))
	}

	// cwd already obtained earlier for loading existing env
	fmt.Printf("\n%s %s\n\n", ui.IconFolder, ui.MsgF("init_in_dir", cwd))

	// Check if already initialized
	composePath := filepath.Join(cwd, "docker-compose.yml")
	if _, err := os.Stat(composePath); err == nil {
		var overwrite bool
		if opts.NonInteractive || opts.Force {
			overwrite = true // Auto-overwrite in force mode
			ui.ShowInfo(ui.Msg("compose_exists_force_init"))
		} else {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(ui.Msg("compose_exists")).
						Value(&overwrite),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
		}
		if !overwrite {
			return errors.New(ui.Msg("init_cancelled"))
		}

		// Backup existing config files before overwrite
		if err := backupExistingConfigs(cwd); err != nil {
			ui.ShowWarning(fmt.Sprintf("Cannot backup existing files: %v", err))
		}
	}

	// Step 3: Service Selection (SeaweedFS, Caddy only)
	ui.ShowStepHeader(4, 7, ui.Msg("step_options"))
	enableSeaweedFS := true // Default: enabled (recommended)
	enableCaddy := true     // Default: enabled (recommended)

	if !opts.NonInteractive && !opts.Force {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(ui.IconStorage+" "+ui.Msg("enable_seaweedfs")).
					Description(ui.Msg("seaweedfs_desc")).
					Affirmative(ui.Msg("yes_recommended")).
					Negative(ui.Msg("no")).
					Value(&enableSeaweedFS),

				huh.NewConfirm().
					Title(ui.IconWeb+" "+ui.Msg("enable_caddy")).
					Description(ui.Msg("caddy_desc")).
					Affirmative(ui.Msg("yes_recommended")).
					Negative(ui.Msg("no")).
					Value(&enableCaddy),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}
	}

	// Step 4: Domain Configuration
	ui.ShowStepHeader(5, 7, ui.Msg("step_domain"))
	// Pre-fill domain from existing env or use localhost
	domain := existingEnv["SYSTEM_DOMAIN"]
	if opts.NonInteractive {
		domain = opts.Domain
	} else if domain == "" {
		domain = "localhost"
	}
	if !opts.NonInteractive && !opts.Force {
		domainForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(ui.IconLink + " " + ui.Msg("enter_domain")).
					Value(&domain).
					Placeholder("localhost").
					Validate(validateDomain),
			),
		)
		if err := domainForm.Run(); err != nil {
			return err
		}
		if domain == "" {
			domain = "localhost"
		}
	}

	// Timezone detection and prompt (within domain step)
	timezone := existingEnv["TZ"]
	if timezone == "" {
		timezone = getSystemTimezone()
	}
	if !opts.NonInteractive && !opts.Force {
		tzForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(ui.IconClock + " " + ui.Msg("enter_timezone")).
					Value(&timezone).
					Placeholder("Asia/Ho_Chi_Minh"),
			),
		)
		if err := tzForm.Run(); err != nil {
			return err
		}
		if timezone == "" {
			timezone = "Asia/Ho_Chi_Minh"
		}
	}

	// Step 5: Environment Configuration
	ui.ShowStepHeader(6, 7, ui.Msg("step_credentials"))

	// Load secrets from existing env or generate new ones
	// Only use existing values if they meet minimum length requirements
	jwtSecret := existingEnv["JWT_SECRET"]
	if len(jwtSecret) < 32 {
		var err error
		jwtSecret, err = generatePasswordWithRetry(32)
		if err != nil {
			return fmt.Errorf("%s: %w", ui.Msg("error_jwt_secret"), err)
		}
	}

	dbPass := existingEnv["DB_PASSWORD"]
	if len(dbPass) < 16 {
		var err error
		dbPass, err = generatePasswordWithRetry(24)
		if err != nil {
			return fmt.Errorf("%s: %w", ui.Msg("error_db_password"), err)
		}
	}

	dbRootPass := existingEnv["DB_ROOT_PASSWORD"]
	if len(dbRootPass) < 16 {
		var err error
		dbRootPass, err = generatePasswordWithRetry(24)
		if err != nil {
			return fmt.Errorf("%s: %w", ui.Msg("error_db_root_pass"), err)
		}
	}

	redisPass := existingEnv["REDIS_PASSWORD"]
	if len(redisPass) < 16 {
		var err error
		redisPass, err = generatePasswordWithRetry(24)
		if err != nil {
			return fmt.Errorf("%s: %w", ui.Msg("error_redis_pass"), err)
		}
	}

	s3AccessKey := existingEnv["S3_ACCESS_KEY"]
	if len(s3AccessKey) < 16 {
		var err error
		s3AccessKey, err = generateS3AccessKeyWithRetry(20)
		if err != nil {
			return fmt.Errorf("%s: %w", ui.Msg("error_s3_access_key"), err)
		}
	}

	s3SecretKey := existingEnv["S3_SECRET_KEY"]
	if len(s3SecretKey) < 32 {
		var err error
		s3SecretKey, err = generatePasswordWithRetry(40)
		if err != nil {
			return fmt.Errorf("%s: %w", ui.Msg("error_s3_secret_key"), err)
		}
	}

	// Ask: Use random secrets?
	useRandom := true // Always use random secrets in force mode
	if !opts.NonInteractive && !opts.Force {
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(ui.Msg("ask_use_random")).
					Description(ui.Msg("ask_use_random_desc")).
					Affirmative(ui.Msg("yes")).
					Negative(ui.Msg("no_edit")).
					Value(&useRandom),
			),
		)
		if err := confirmForm.Run(); err != nil {
			return err
		}
	}

	// If No -> Show grouped edit form
	if !useRandom && !opts.NonInteractive && !opts.Force {
		groups := []*huh.Group{}

		// Group 1: System Configuration
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("JWT_SECRET").
				Value(&jwtSecret).
				Validate(validateMinLength(32, "JWT_SECRET")),
		).Title(ui.Msg("group_system")))

		// Group 2: Database Secrets
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("DB_PASSWORD").
				Value(&dbPass).
				Validate(validateMinLength(16, "DB_PASSWORD")),
			huh.NewInput().
				Title("DB_ROOT_PASSWORD").
				Value(&dbRootPass).
				Validate(validateMinLength(16, "DB_ROOT_PASSWORD")),
			huh.NewInput().
				Title("REDIS_PASSWORD").
				Value(&redisPass).
				Validate(validateMinLength(16, "REDIS_PASSWORD")),
		).Title(ui.Msg("group_db_secrets")))

		// Group 3: S3 Secrets (only if SeaweedFS enabled)
		if enableSeaweedFS {
			groups = append(groups, huh.NewGroup(
				huh.NewInput().
					Title("S3_ACCESS_KEY").
					Value(&s3AccessKey).
					Validate(validateMinLength(16, "S3_ACCESS_KEY")),
				huh.NewInput().
					Title("S3_SECRET_KEY").
					Value(&s3SecretKey).
					Validate(validateMinLength(32, "S3_SECRET_KEY")),
			).Title(ui.Msg("group_s3_secrets")))
		}

		editForm := huh.NewForm(groups...)
		if err := editForm.Run(); err != nil {
			return err
		}
	}

	// Step 6: Generate Files + Complete
	ui.ShowStepHeader(7, 7, ui.Msg("step_generate"))

	// Render templates with spinner
	spinner = startInitSpinner(ui.IconWrite + " " + ui.Msg("generating_files"))

	tmplCfg := templates.Config{
		EnableSeaweedFS: enableSeaweedFS,
		EnableCaddy:     enableCaddy,
		Domain:          domain,
		Timezone:        timezone,
		JWTSecret:       jwtSecret,
		LicenseKey:      licenseData.Key,
		ServerPublicKey: licenseData.PublicKey,
		DBPassword:      dbPass,
		DBRootPassword:  dbRootPass,
		RedisPassword:   redisPass,
		S3AccessKey:     s3AccessKey,
		S3SecretKey:     s3SecretKey,
	}

	if err := renderTemplates(tmplCfg, cwd); err != nil {
		spinner.Fail(fmt.Sprintf("%s: %v", ui.Msg("error_create_file"), err))
		return NewExitError(exitCodeRenderFailure, fmt.Errorf("%s: %w", ui.Msg("error_create_file"), err))
	}

	spinner.Success(ui.IconCheck + " " + ui.Msg("files_generated"))

	// Save project directory to config
	cfg.ProjectDir = cwd
	if saveErr := cfg.Save(); saveErr != nil {
		ui.ShowWarning(fmt.Sprintf("Cannot save config: %v", saveErr))
	}

	// Show completion summary
	// Collect created files
	createdFiles := []string{"docker-compose.yml", ".env", "kkphp.conf"}
	if enableCaddy {
		createdFiles = append(createdFiles, "Caddyfile")
	}
	if enableSeaweedFS {
		createdFiles = append(createdFiles, "kkfiler.toml")
	}

	// Show summary table
	ui.PrintInitSummary(enableSeaweedFS, enableCaddy, domain, createdFiles, cwd)

	// Show completion banner
	fmt.Println()
	ui.ShowCompletionBanner(true, ui.IconComplete+" "+ui.Msg("init_complete"), ui.Msg("next_steps_box"))

	return nil
}

func showInitInputError(err error) {
	if err == nil || ExitCode(err) != exitCodeInputValidation {
		return
	}
	ui.ShowBoxedError(ui.ErrorSuggestion{
		Title:      ui.Msg("err_invalid_init_input"),
		Message:    ui.SanitizeError(err),
		Suggestion: ui.Msg("err_invalid_init_input_suggestion"),
	})
}

// backupExistingConfigs creates a timestamped backup folder and copies existing config files into it
func backupExistingConfigs(dir string) error {
	configFiles := []string{
		"docker-compose.yml",
		".env",
		"Caddyfile",
		"kkfiler.toml",
		"kkphp.conf",
	}

	timestamp := time.Now().Format("20060102150405")
	backupDirName := "backup-" + timestamp
	backupDir := filepath.Join(dir, backupDirName)

	// First pass: check which files exist
	var toBackup []string
	for _, filename := range configFiles {
		srcPath := filepath.Join(dir, filename)
		if _, err := os.Stat(srcPath); err == nil {
			toBackup = append(toBackup, filename)
		}
	}

	if len(toBackup) == 0 {
		return nil
	}

	// Create backup folder
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("cannot create backup dir: %w", err)
	}

	// Copy each file into backup folder
	var backedUp []string
	for _, filename := range toBackup {
		srcPath := filepath.Join(dir, filename)
		data, err := os.ReadFile(srcPath)
		if err != nil {
			continue // Skip on error
		}

		dstPath := filepath.Join(backupDir, filename)
		mode := os.FileMode(0644)
		if filename == ".env" {
			mode = 0600
		}
		if err := os.WriteFile(dstPath, data, mode); err != nil {
			continue // Skip on error
		}

		backedUp = append(backedUp, filename)
	}

	if len(backedUp) > 0 {
		ui.ShowInfo(fmt.Sprintf("Backed up to: %s/", backupDirName))
	}

	return nil
}

// generateS3AccessKey generates alphanumeric uppercase key for S3 access
func generateS3AccessKey(length int) (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[idx.Int64()]
	}
	return string(result), nil
}

// generatePasswordWithRetry generates password with retry logic (max 3 attempts)
func generatePasswordWithRetry(length int) (string, error) {
	const maxRetries = 3
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		pass, err := ui.GeneratePassword(length)
		if err == nil {
			return pass, nil
		}
		lastErr = err
	}
	return "", lastErr
}

// generateS3AccessKeyWithRetry generates S3 access key with retry logic
func generateS3AccessKeyWithRetry(length int) (string, error) {
	const maxRetries = 3
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		key, err := generateS3AccessKey(length)
		if err == nil {
			return key, nil
		}
		lastErr = err
	}
	return "", lastErr
}

func ensureInitDocker(opts initOptions, licenseKey, licensePublicKey string) error {
	if opts.Force {
		if err := DockerValidatorInstance.CheckDockerInstalled(); err != nil {
			ui.ShowWarning(ui.Msg("docker_not_installed_force_init"))
		} else if err := DockerValidatorInstance.CheckDockerDaemon(); err != nil {
			ui.ShowWarning(ui.Msg("docker_daemon_not_running_force_init"))
		} else if err := DockerValidatorInstance.CheckComposeVersion(); err != nil {
			ui.ShowWarning(ui.Msg("docker_compose_issue_force_init"))
		}
		return nil
	}

	ensureOpts := validator.EnsureDockerOptions{
		AutoFix: opts.NonInteractive && opts.InstallDocker,
	}
	if opts.NonInteractive && opts.InstallDocker {
		ensureOpts.MaxRetries = 1
	}

	if !opts.NonInteractive {
		ensureOpts.ConfirmInstall = func() (bool, error) {
			ui.ShowInfo(ui.IconDocker + " " + ui.Msg("docker_not_installed"))
			var installDocker bool
			installForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(ui.IconDocker + " " + ui.Msg("ask_install_docker")).
						Description(ui.Msg("ask_install_docker_desc")).
						Affirmative(ui.Msg("yes_install")).
						Negative(ui.Msg("no_manual")).
						Value(&installDocker),
				),
			)
			if formErr := installForm.Run(); formErr != nil {
				return false, formErr
			}
			return installDocker, nil
		}
		ensureOpts.ConfirmStart = func() (bool, error) {
			ui.ShowInfo(ui.IconDocker + " " + ui.Msg("docker_not_running"))
			var startDocker bool
			startForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(ui.IconDocker + " " + ui.Msg("ask_start_docker")).
						Affirmative(ui.Msg("yes")).
						Negative(ui.Msg("no")).
						Value(&startDocker),
				),
			)
			if formErr := startForm.Run(); formErr != nil {
				return false, formErr
			}
			return startDocker, nil
		}
		ensureOpts.Install = func() error {
			ui.ShowInfo(ui.IconDocker + " " + ui.Msg("installing_docker"))
			ui.ShowNote(ui.Msg("docker_install_in_progress_note"))
			if validator.IsInteractiveTTY() {
				ui.ShowNote(ui.Msg("docker_sudo_password_hint"))
			}
			err := DockerValidatorInstance.InstallDocker()
			if err != nil {
				return err
			}
			if DockerValidatorInstance.CheckDockerDaemon() == nil {
				ui.ShowSuccess(ui.IconCheck + " " + ui.Msg("docker_installed"))
			} else if DockerValidatorInstance.IsDockerDaemonRunningPrivileged() {
				ui.ShowSuccess(ui.IconCheck + " " + ui.Msg("docker_installed_daemon_running"))
				ui.ShowNote(ui.Msg("docker_group_activate_note"))
			} else {
				ui.ShowSuccess(ui.IconCheck + " " + ui.Msg("docker_installed"))
			}
			return nil
		}
		ensureOpts.Start = func() error {
			ui.ShowInfo(ui.IconDocker + " " + ui.Msg("starting_docker"))
			err := DockerValidatorInstance.StartDockerDaemon()
			if err != nil {
				return err
			}
			ui.ShowSuccess(ui.IconCheck + " " + ui.Msg("docker_started"))
			return nil
		}
	}

	err := DockerValidatorInstance.EnsureDockerReady(ensureOpts)
	if err == nil {
		return nil
	}

	if reexecErr := tryReexecInitWithDockerGroup(err, licenseKey, licensePublicKey); reexecErr != nil {
		return formatInitDockerError(opts, reexecErr)
	}

	return nil
}

func formatInitDockerError(opts initOptions, err error) error {
	key := validator.UserErrorKey(err)
	switch {
	case key == "docker_not_installed":
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("docker_not_found"),
			Message:    ui.Msg("docker_required"),
			Suggestion: ui.Msg("err_docker_manual_install_suggestion"),
		})
		if opts.NonInteractive {
			return NewExitError(exitCodeDockerValidation, err)
		}
		return errors.New(ui.Msg("docker_required"))
	case key == "docker_permission_not_effective":
		title := ui.Msg("docker_permission_pending_title")
		suggestion := validator.UserErrorSuggestion(err)
		command := dockerGroupReexecHint()
		if !DockerValidatorInstance.HasDockerGroupRunner() {
			title = ui.Msg("docker_session_relogin_title")
			suggestion = ui.Msg("docker_session_relogin_suggestion")
			command = ui.Msg("docker_session_relogin_command")
		}
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      title,
			Message:    validator.FormatUserErrorForBox(err),
			Suggestion: suggestion,
			Command:    command,
		})
		if opts.NonInteractive {
			return NewExitError(exitCodeDockerValidation, err)
		}
		return err
	case key == "docker_not_running", key == "docker_permission_denied":
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("docker_daemon_stopped"),
			Message:    validator.FormatUserErrorForBox(err),
			Suggestion: validator.UserErrorSuggestion(err),
			Command:    "sudo systemctl start docker && kk init --yes --install-docker ...",
		})
		if opts.NonInteractive {
			return NewExitError(exitCodeDockerValidation, err)
		}
		return err
	case validator.IsDockerInstallError(err):
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("docker_install_failed"),
			Message:    validator.FormatUserErrorForBox(err),
			Suggestion: validator.UserErrorSuggestion(err),
			Command:    ui.Msg("docker_check_network_command"),
		})
		return err
	case key == "docker_start_failed", key == "docker_daemon_wait_timeout":
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("docker_daemon_stopped"),
			Message:    validator.FormatUserErrorForBox(err),
			Suggestion: validator.UserErrorSuggestion(err),
			Command:    "sudo systemctl start docker && kk init",
		})
		return err
	case key == "compose_not_found", key == "compose_version_old":
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("docker_compose_issue"),
			Message:    validator.FormatUserErrorForBox(err),
			Suggestion: ui.Msg("err_docker_compose_update_suggestion"),
		})
		if opts.NonInteractive {
			return NewExitError(exitCodeDockerValidation, err)
		}
		return err
	default:
		if opts.NonInteractive {
			return NewExitError(exitCodeDockerValidation, err)
		}
		return err
	}
}

// validateDomain validates domain format (RFC 1123 hostname or localhost)
func validateDomain(s string) error {
	if s == "" {
		return nil // Empty allowed, defaults to localhost
	}
	if s == "localhost" {
		return nil
	}
	if !domainRegex.MatchString(s) {
		return errors.New(ui.Msg("error_invalid_domain"))
	}
	return nil
}

// validateMinLength returns a validator function for minimum length
func validateMinLength(minLen int, fieldName string) func(string) error {
	return func(s string) error {
		if len(s) < minLen {
			return fmt.Errorf("%s must be at least %d characters", fieldName, minLen)
		}
		return nil
	}
}

// getSystemTimezone detects the system timezone on Linux.
// Tries: timedatectl → /etc/timezone → /etc/localtime symlink → fallback.
func getSystemTimezone() string {
	const fallback = "Asia/Ho_Chi_Minh"

	// Method 1: timedatectl (systemd)
	out, err := exec.Command("timedatectl", "show", "--property=Timezone", "--value").Output()
	if err == nil {
		if tz := strings.TrimSpace(string(out)); tz != "" {
			return tz
		}
	}

	// Method 2: /etc/timezone (Debian/Ubuntu)
	data, err := os.ReadFile("/etc/timezone")
	if err == nil {
		if tz := strings.TrimSpace(string(data)); tz != "" {
			return tz
		}
	}

	// Method 3: /etc/localtime symlink (RHEL/Fedora/Arch)
	link, err := os.Readlink("/etc/localtime")
	if err == nil {
		// e.g. /usr/share/zoneinfo/Asia/Ho_Chi_Minh → Asia/Ho_Chi_Minh
		const prefix = "zoneinfo/"
		if idx := strings.LastIndex(link, prefix); idx != -1 {
			return link[idx+len(prefix):]
		}
	}

	return fallback
}

// loadExistingEnv parses existing .env file and returns key-value map
func loadExistingEnv(dir string) map[string]string {
	result := make(map[string]string)
	envPath := filepath.Join(dir, ".env")

	data, err := os.ReadFile(envPath)
	if err != nil {
		return result // File doesn't exist or unreadable
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes (single or double)
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
					(value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}
			result[key] = value
		}
	}
	return result
}
