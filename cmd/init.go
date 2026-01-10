package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

var DockerValidatorInstance *validator.DockerValidator

func init() {
	rootCmd.AddCommand(initCmd)
	DockerValidatorInstance = validator.NewDockerValidator()
}

func runInit(cmd *cobra.Command, args []string) error {
	// Command banner
	ui.ShowCommandBanner("kk init", ui.Msg("init_desc"))

	// Step 1: Check Docker
	ui.ShowStepHeader(1, 5, ui.Msg("step_docker_check"))
	ui.ShowInfo(ui.IconDocker + " " + ui.MsgCheckingDocker())

	// Check Docker installation
	dockerInstalled := true
	if err := DockerValidatorInstance.CheckDockerInstalled(); err != nil {
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

	// Check Docker daemon if Docker is installed
	if dockerInstalled {
		if err := DockerValidatorInstance.CheckDockerDaemon(); err != nil {
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

		// Check Docker Compose version
		if err := DockerValidatorInstance.CheckComposeVersion(); err != nil {
			ui.ShowBoxedError(ui.ErrorSuggestion{
				Title:      ui.Msg("docker_compose_issue"),
				Message:    err.Error(),
				Suggestion: "Update Docker to latest version",
			})
			return err
		}
	}

	ui.ShowSuccess(ui.IconCheck + " " + ui.MsgDockerOK())

	// Step 2: Language selection
	ui.ShowStepHeader(2, 5, ui.Msg("step_language"))
	var langChoice string
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
		if !overwrite {
			return errors.New(ui.Msg("init_cancelled"))
		}

		// Backup existing config files before overwrite
		if err := backupExistingConfigs(cwd); err != nil {
			ui.ShowWarning(fmt.Sprintf("Cannot backup existing files: %v", err))
		}
	}

	// Step 3: Configuration options
	ui.ShowStepHeader(3, 5, ui.Msg("step_options"))
	enableSeaweedFS := true // Default: enabled (recommended)
	enableCaddy := true     // Default: enabled (recommended)
	var domain string

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

	// If Caddy enabled, ask for domain
	if enableCaddy {
		domainForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(ui.IconLink + " " + ui.Msg("enter_domain")).
					Value(&domain).
					Placeholder("localhost"),
			),
		)
		if err := domainForm.Run(); err != nil {
			return err
		}
		if domain == "" {
			domain = "localhost"
		}
	}

	// Step 4: Generate files
	ui.ShowStepHeader(4, 5, ui.Msg("step_generate"))
	dbPass, err := ui.GeneratePassword(24)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_db_password"), err)
	}
	dbRootPass, err := ui.GeneratePassword(24)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_db_root_pass"), err)
	}
	redisPass, err := ui.GeneratePassword(24)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_redis_pass"), err)
	}

	// Render templates with spinner
	spinner, _ := pterm.DefaultSpinner.Start(ui.IconWrite + " " + ui.Msg("generating_files"))

	tmplCfg := templates.Config{
		EnableSeaweedFS: enableSeaweedFS,
		EnableCaddy:     enableCaddy,
		DBPassword:      dbPass,
		DBRootPassword:  dbRootPass,
		RedisPassword:   redisPass,
		Domain:          domain,
	}

	if err := templates.RenderAll(tmplCfg, cwd); err != nil {
		spinner.Fail(ui.MsgF("error_create_file", err.Error()))
		return fmt.Errorf("%s: %w", ui.Msg("error_create_file"), err)
	}

	spinner.Success(ui.IconCheck + " " + ui.Msg("files_generated"))

	// Step 5: Complete - show summary
	ui.ShowStepHeader(5, 5, ui.Msg("step_complete"))

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
