package cmd

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/config"
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
	forceInit bool
	DockerValidatorInstance *validator.DockerValidator
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Bỏ qua tất cả các lời nhắc tương tác và sử dụng các giá trị mặc định")
	DockerValidatorInstance = validator.NewDockerValidator()
}

func runInit(cmd *cobra.Command, args []string) error {
	// Command banner
	ui.ShowCommandBanner("kk init", ui.Msg("init_desc"))

	// Step 1: Check Docker
	ui.ShowStepHeader(1, 6, ui.Msg("step_docker_check"))
	ui.ShowInfo(ui.IconDocker + " " + ui.MsgCheckingDocker())

	// Check Docker installation
	dockerInstalled := true
	if err := DockerValidatorInstance.CheckDockerInstalled(); err != nil {
		if forceInit {
			ui.ShowWarning(ui.Msg("docker_not_installed_force_init"))
			// In force mode, assume Docker will be handled externally or allow to proceed with potential issues
			dockerInstalled = true
		} else {
			dockerInstalled = false
			ui.ShowWarning(ui.Msg("docker_not_installed"))

			// Ask user if they want to install Docker
			var installDocker bool
			installForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(ui.IconDocker+" "+ui.Msg("ask_install_docker")).
						Description(ui.Msg("ask_install_docker_desc")).
						Affirmative(ui.Msg("yes_install")).
						Negative(ui.Msg("no_manual")).
						Value(&installDocker),
				),
			)
			if err := installForm.Run(); err != nil {
				return err
			}

			if installDocker {
				// Install Docker with spinner
				spinner, _ := pterm.DefaultSpinner.Start(ui.IconDocker + " " + ui.Msg("installing_docker"))
				if err := DockerValidatorInstance.InstallDocker(); err != nil {
					spinner.Fail(ui.Msg("docker_install_failed"))
					ui.ShowBoxedError(ui.ErrorSuggestion{
						Title:      ui.Msg("docker_install_failed"),
						Message:    err.Error(),
						Suggestion: "Install manually: https://docs.docker.com/get-docker/",
					})
					return err
				}
				spinner.Success(ui.IconCheck + " " + ui.Msg("docker_installed"))
				dockerInstalled = true
			} else {
				ui.ShowBoxedError(ui.ErrorSuggestion{
					Title:      ui.Msg("docker_not_found"),
					Message:    ui.Msg("docker_required"),
					Suggestion: "Install Docker from https://docs.docker.com/get-docker/",
				})
				return errors.New(ui.Msg("docker_required"))
			}
		}
	}

	// Check Docker daemon if Docker is installed
	if dockerInstalled {
		if err := DockerValidatorInstance.CheckDockerDaemon(); err != nil {
			if forceInit {
				ui.ShowWarning(ui.Msg("docker_daemon_not_running_force_init"))
				// In force mode, assume daemon will be started externally or allow to proceed
			} else {
				ui.ShowWarning(ui.Msg("docker_not_running"))

				// Ask to start Docker daemon
				var startDocker bool
				startForm := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(ui.IconDocker+" "+ui.Msg("ask_start_docker")).
							Affirmative(ui.Msg("yes")).
							Negative(ui.Msg("no")).
							Value(&startDocker),
					),
				)
				if err := startForm.Run(); err != nil {
					return err
				}

				if startDocker {
					spinner, _ := pterm.DefaultSpinner.Start(ui.IconDocker + " " + ui.Msg("starting_docker"))
					if err := DockerValidatorInstance.StartDockerDaemon(); err != nil {
						spinner.Fail(ui.Msg("docker_start_failed"))
						ui.ShowBoxedError(ui.ErrorSuggestion{
							Title:      ui.Msg("docker_daemon_stopped"),
							Message:    err.Error(),
							Suggestion: "Start Docker daemon",
							Command:    "systemctl start docker",
						})
						return err
					}
					spinner.Success(ui.IconCheck + " " + ui.Msg("docker_started"))
				} else {
					ui.ShowBoxedError(ui.ErrorSuggestion{
						Title:      ui.Msg("docker_daemon_stopped"),
						Message:    ui.Msg("docker_required"),
						Suggestion: "Start Docker daemon",
						Command:    "systemctl start docker",
					})
					return errors.New(ui.Msg("docker_required"))
				}
			}
		}

		// Check Docker Compose version
		if err := DockerValidatorInstance.CheckComposeVersion(); err != nil {
			if forceInit {
				ui.ShowWarning(ui.Msg("docker_compose_issue_force_init"))
			} else {
				ui.ShowBoxedError(ui.ErrorSuggestion{
					Title:      ui.Msg("docker_compose_issue"),
					Message:    err.Error(),
					Suggestion: "Update Docker to latest version",
				})
				return err
			}
		}
	}

	ui.ShowSuccess(ui.IconCheck + " " + ui.MsgDockerOK())

	// Step 2: Language selection
	ui.ShowStepHeader(2, 6, ui.Msg("step_language"))
	var langChoice string
	if forceInit {
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
		if err := langForm.Run(); err != nil {
			return err
		}
		// Set default to English if no selection
		if langChoice == "" {
			langChoice = "en"
		}
	}
	ui.SetLanguage(ui.Language(langChoice))

	// Save language preference to config
	cfg, _ := config.Load()
	cfg.Language = langChoice
	_ = cfg.Save() // Best effort, don't fail init if config save fails

	// Get working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("\n%s %s\n\n", ui.IconFolder, ui.MsgF("init_in_dir", cwd))

	// Check if already initialized
	composePath := filepath.Join(cwd, "docker-compose.yml")
	if _, err := os.Stat(composePath); err == nil {
		var overwrite bool
		if forceInit {
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
	ui.ShowStepHeader(3, 6, ui.Msg("step_options"))
	enableSeaweedFS := true // Default: enabled (recommended)
	enableCaddy := true     // Default: enabled (recommended)

	if !forceInit {
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
	ui.ShowStepHeader(4, 6, ui.Msg("step_domain"))
	domain := "localhost"
	if !forceInit {
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

	// Step 5: Environment Configuration
	ui.ShowStepHeader(5, 6, ui.Msg("step_credentials"))

	// Pre-generate all secrets with retry logic
	jwtSecret, err := generatePasswordWithRetry(32)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_jwt_secret"), err)
	}
	dbPass, err := generatePasswordWithRetry(24)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_db_password"), err)
	}
	dbRootPass, err := generatePasswordWithRetry(24)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_db_root_pass"), err)
	}
	redisPass, err := generatePasswordWithRetry(24)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_redis_pass"), err)
	}
	s3AccessKey, err := generateS3AccessKeyWithRetry(20)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_s3_access_key"), err)
	}
	s3SecretKey, err := generatePasswordWithRetry(40)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_s3_secret_key"), err)
	}

	// Ask: Use random secrets?
	useRandom := true // Always use random secrets in force mode
	if !forceInit {
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
	if !useRandom && !forceInit {
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
	ui.ShowStepHeader(6, 6, ui.Msg("step_generate"))

	// Render templates with spinner
	spinner, _ := pterm.DefaultSpinner.Start(ui.IconWrite + " " + ui.Msg("generating_files"))

	tmplCfg := templates.Config{
		EnableSeaweedFS: enableSeaweedFS,
		EnableCaddy:     enableCaddy,
		Domain:          domain,
		JWTSecret:       jwtSecret,
		DBPassword:      dbPass,
		DBRootPassword:  dbRootPass,
		RedisPassword:   redisPass,
		S3AccessKey:     s3AccessKey,
		S3SecretKey:     s3SecretKey,
	}

	if err := templates.RenderAll(tmplCfg, cwd); err != nil {
		spinner.Fail(ui.MsgF("error_create_file", err.Error()))
		return fmt.Errorf("%s: %w", ui.Msg("error_create_file"), err)
	}

	spinner.Success(ui.IconCheck + " " + ui.Msg("files_generated"))

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
	ui.PrintInitSummary(enableSeaweedFS, enableCaddy, domain, createdFiles)

	// Show completion banner
	fmt.Println()
	ui.ShowCompletionBanner(true, ui.IconComplete+" "+ui.Msg("init_complete"), ui.Msg("next_steps_box"))

	return nil
}

// backupExistingConfigs creates .bak backups of existing config files
func backupExistingConfigs(dir string) error {
	configFiles := []string{
		"docker-compose.yml",
		".env",
		"Caddyfile",
		"kkfiler.toml",
		"kkphp.conf",
	}

	var backedUp []string
	for _, filename := range configFiles {
		srcPath := filepath.Join(dir, filename)
		if _, err := os.Stat(srcPath); err == nil {
			// File exists, create backup
			bakPath := srcPath + ".bak"

			// Read source
			data, err := os.ReadFile(srcPath)
			if err != nil {
				continue // Skip on error
			}

			// Write backup
			if err := os.WriteFile(bakPath, data, 0644); err != nil {
				continue // Skip on error
			}

			backedUp = append(backedUp, filename)
		}
	}

	if len(backedUp) > 0 {
		ui.ShowInfo(fmt.Sprintf("Backed up: %s", strings.Join(backedUp, ", ")))
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

// validateDomain validates domain format (RFC 1123 hostname or localhost)
func validateDomain(s string) error {
	if s == "" {
		return nil // Empty allowed, defaults to localhost
	}
	if s == "localhost" {
		return nil
	}
	// RFC 1123 hostname pattern
	pattern := `^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, s)
	if !matched {
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
