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

	"github.com/kkauto-net/kk-install/pkg/templates"
	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/validator"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Khoi tao kkengine Docker stack",
	Long:  `Tao docker-compose.yml va cac file config can thiet.`,
	RunE:  runInit,
}

var DockerValidatorInstance *validator.DockerValidator

func init() {
	rootCmd.AddCommand(initCmd)
	DockerValidatorInstance = validator.NewDockerValidator()
}

func runInit(cmd *cobra.Command, args []string) error {
	// Step 1: Check Docker
	ui.ShowInfo(ui.IconDocker + " " + ui.MsgCheckingDocker())
	if err := DockerValidatorInstance.CheckDockerInstalled(); err != nil {
		ui.ShowError(err.Error())
		return err
	}
	if err := DockerValidatorInstance.CheckDockerDaemon(); err != nil {
		ui.ShowError(err.Error())
		return err
	}
	if err := DockerValidatorInstance.CheckComposeVersion(); err != nil {
		ui.ShowError(err.Error())
		return err
	}
	ui.ShowSuccess(ui.IconCheck + " " + ui.MsgDockerOK())

	// Step 2: Language selection
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

	// Step 3: Get working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("\n%s %s\n\n", ui.IconFolder, ui.MsgF("init_in_dir", cwd))

	// Step 4: Check if already initialized
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

	// Step 5: Interactive prompts
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

	// Step 6: Generate passwords
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

	// Step 7: Render templates with spinner
	spinner, _ := pterm.DefaultSpinner.Start(ui.IconWrite + " " + ui.Msg("generating_files"))

	cfg := templates.Config{
		EnableSeaweedFS: enableSeaweedFS,
		EnableCaddy:     enableCaddy,
		DBPassword:      dbPass,
		DBRootPassword:  dbRootPass,
		RedisPassword:   redisPass,
		Domain:          domain,
	}

	if err := templates.RenderAll(cfg, cwd); err != nil {
		spinner.Fail(ui.MsgF("error_create_file", err.Error()))
		return fmt.Errorf("%s: %w", ui.Msg("error_create_file"), err)
	}

	spinner.Success(ui.IconCheck + " " + ui.Msg("files_generated"))

	// Step 8: Show success
	fmt.Println()
	ui.ShowSuccess(ui.IconCheck + " " + ui.MsgCreated("docker-compose.yml"))
	ui.ShowSuccess(ui.IconCheck + " " + ui.MsgCreated(".env"))
	ui.ShowSuccess(ui.IconCheck + " " + ui.MsgCreated("kkphp.conf"))
	if enableCaddy {
		ui.ShowSuccess(ui.IconCheck + " " + ui.MsgCreated("Caddyfile"))
	}
	if enableSeaweedFS {
		ui.ShowSuccess(ui.IconCheck + " " + ui.MsgCreated("kkfiler.toml"))
	}

	// Step 9: Show completion box
	fmt.Println()
	pterm.DefaultBox.
		WithTitle(ui.IconComplete + " " + ui.Msg("init_complete")).
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgGreen)).
		Println(ui.Msg("next_steps_box"))

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
