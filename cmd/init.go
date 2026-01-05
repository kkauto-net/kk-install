package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
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
	ui.ShowInfo(ui.MsgCheckingDocker())
	if err := DockerValidatorInstance.CheckDockerInstalled(); err != nil {
		ui.ShowError(err.Error())
		return err
	}
	if err := DockerValidatorInstance.CheckDockerDaemon(); err != nil {
		ui.ShowError(err.Error())
		return err
	}
	ui.ShowSuccess(ui.MsgDockerOK())

	// Step 2: Language selection
	var langChoice string
	langForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(ui.Msg("select_language")).
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
	fmt.Printf("\n%s\n\n", ui.MsgF("init_in_dir", cwd))

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
	}

	// Step 5: Interactive prompts
	enableSeaweedFS := true // Default: enabled (recommended)
	enableCaddy := true     // Default: enabled (recommended)
	var domain string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(ui.Msg("enable_seaweedfs")).
				Description(ui.Msg("seaweedfs_desc")).
				Affirmative(ui.Msg("yes_recommended")).
				Negative(ui.Msg("no")).
				Value(&enableSeaweedFS),

			huh.NewConfirm().
				Title(ui.Msg("enable_caddy")).
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
					Title(ui.Msg("enter_domain")).
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

	// Step 7: Render templates
	cfg := templates.Config{
		EnableSeaweedFS: enableSeaweedFS,
		EnableCaddy:     enableCaddy,
		DBPassword:      dbPass,
		DBRootPassword:  dbRootPass,
		RedisPassword:   redisPass,
		Domain:          domain,
	}

	if err := templates.RenderAll(cfg, cwd); err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("error_create_file"), err)
	}

	// Step 8: Show success
	fmt.Println()
	ui.ShowSuccess(ui.MsgCreated("docker-compose.yml"))
	ui.ShowSuccess(ui.MsgCreated(".env"))
	ui.ShowSuccess(ui.MsgCreated("kkphp.conf"))
	if enableCaddy {
		ui.ShowSuccess(ui.MsgCreated("Caddyfile"))
	}
	if enableSeaweedFS {
		ui.ShowSuccess(ui.MsgCreated("kkfiler.toml"))
	}

	fmt.Println()
	ui.ShowSuccess(ui.MsgInitComplete())
	fmt.Println(ui.MsgNextSteps())

	return nil
}
