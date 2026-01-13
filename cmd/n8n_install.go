package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/validator"
)

var n8nInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install n8n with interactive setup",
	Long:  `Set up n8n with PostgreSQL database in Docker containers.`,
	RunE:  runN8nInstall,
}

var forceN8nInstall bool

func init() {
	n8nInstallCmd.Flags().BoolVarP(&forceN8nInstall, "force", "f", false,
		"Skip prompts, use defaults")
	n8nCmd.AddCommand(n8nInstallCmd)
}

func runN8nInstall(cmd *cobra.Command, args []string) error {
	ui.ShowCommandBanner("kk n8n install", ui.Msg("n8n_install_desc"))

	// Step 1: Check Docker
	ui.ShowStepHeader(1, 6, ui.Msg("step_docker_check"))
	dv := validator.NewDockerValidator()
	if err := dv.CheckDockerInstalled(); err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      "Docker Required",
			Message:    err.Error(),
			Suggestion: "Install Docker first",
		})
		return err
	}
	if err := dv.CheckDockerDaemon(); err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      "Docker Not Running",
			Message:    err.Error(),
			Suggestion: "Start Docker daemon",
			Command:    "sudo systemctl start docker",
		})
		return err
	}
	ui.ShowSuccess(ui.Msg("docker_ok"))

	// Step 2: Check existing installation
	ui.ShowStepHeader(2, 6, ui.Msg("n8n_check_existing"))
	if n8n.IsInstalled() {
		ui.ShowInfo(ui.Msg("n8n_already_installed"))
		var overwrite bool
		if !forceN8nInstall {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(ui.Msg("n8n_overwrite_prompt")).
						Description(ui.Msg("n8n_overwrite_desc")).
						Value(&overwrite),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
			if !overwrite {
				return errors.New(ui.Msg("init_cancelled"))
			}
		} else {
			overwrite = true
		}
	}

	// Step 3: Domain configuration
	ui.ShowStepHeader(3, 6, ui.Msg("n8n_domain_config"))
	domain := "localhost"
	if !forceN8nInstall {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(ui.Msg("n8n_enter_domain")).
					Description(ui.Msg("n8n_domain_desc")).
					Value(&domain).
					Placeholder("localhost").
					Validate(validateN8nDomain),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
	}
	if domain == "" {
		domain = "localhost"
	}

	// Step 4: Generate credentials
	ui.ShowStepHeader(4, 6, ui.Msg("n8n_generate_credentials"))

	dbPassword, err := ui.GeneratePassword(24)
	if err != nil {
		return fmt.Errorf("failed to generate DB password: %w", err)
	}
	encryptionKey, err := ui.GeneratePassword(32)
	if err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	cfg := n8n.N8nConfig{
		Domain:        domain,
		N8nHost:       domain,
		DBUser:        "n8n",
		DBPassword:    dbPassword,
		EncryptionKey: encryptionKey,
		Timezone:      "Asia/Ho_Chi_Minh",
	}

	// Check kkengine network and ask about connection
	cfg.ConnectKKEngine = false
	if !forceN8nInstall {
		kkengineNetExists := checkKKEngineNetwork()
		if kkengineNetExists {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(ui.Msg("n8n_connect_kkengine")).
						Description(ui.Msg("n8n_connect_kkengine_desc")).
						Value(&cfg.ConnectKKEngine),
				),
			)
			_ = form.Run()
		}
	}

	// Show encryption key warning and get confirmation
	if !forceN8nInstall {
		pterm.Println()
		pterm.DefaultBox.
			WithTitle(pterm.Yellow("⚠ " + ui.Msg("n8n_encryption_key_warning"))).
			WithTitleTopLeft().
			WithBoxStyle(pterm.NewStyle(pterm.FgYellow)).
			Println(fmt.Sprintf("%s\n\n%s", ui.Msg("n8n_encryption_key_desc"), cfg.EncryptionKey))

		var confirmed bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(ui.Msg("n8n_encryption_key_confirm")).
					Affirmative(ui.Msg("yes")).
					Negative(ui.Msg("no")).
					Value(&confirmed),
			),
		)
		if err := confirmForm.Run(); err != nil {
			return err
		}
		if !confirmed {
			return errors.New(ui.Msg("init_cancelled"))
		}
	}

	// Optional: Edit credentials
	useDefaults := true
	if !forceN8nInstall {
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(ui.Msg("ask_use_random")).
					Description(ui.Msg("n8n_credentials_desc")).
					Affirmative(ui.Msg("yes")).
					Negative(ui.Msg("no_edit")).
					Value(&useDefaults),
			),
		)
		if err := confirmForm.Run(); err != nil {
			return err
		}
	}

	if !useDefaults && !forceN8nInstall {
		editForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(ui.Msg("n8n_db_password")).
					Value(&cfg.DBPassword),
				huh.NewInput().
					Title(ui.Msg("n8n_encryption_key")).
					Value(&cfg.EncryptionKey),
			),
		)
		if err := editForm.Run(); err != nil {
			return err
		}
	}

	// Step 5: Generate files
	ui.ShowStepHeader(5, 6, ui.Msg("n8n_generate_files"))
	spinner := ui.StartPtermSpinner(ui.Msg("n8n_generating"))

	if err := n8n.RenderAll(cfg); err != nil {
		spinner.Fail(ui.Msg("n8n_generate_failed"))
		return err
	}
	spinner.Success(ui.Msg("n8n_files_generated"))

	// Show summary
	pterm.Println()
	pterm.DefaultSection.Println(ui.Msg("n8n_install_summary"))
	pterm.DefaultTable.
		WithHasHeader(false).
		WithBoxed(true).
		WithData(pterm.TableData{
			{ui.Msg("install_location"), n8n.N8nDir()},
			{ui.Msg("domain"), cfg.Domain},
			{"Database", "PostgreSQL (n8n-postgres)"},
			{"Port", "5678"},
		}).
		Render()

	// Step 6: Ask to start
	ui.ShowStepHeader(6, 6, ui.Msg("n8n_ready_to_start"))
	var startNow bool
	if !forceN8nInstall {
		startForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(ui.Msg("n8n_start_now")).
					Value(&startNow),
			),
		)
		_ = startForm.Run()
	} else {
		startNow = true
	}

	if startNow {
		return runN8nStartInternal()
	}

	pterm.Println()
	ui.ShowSuccess(ui.Msg("n8n_install_complete"))
	pterm.Println("  " + ui.Msg("n8n_run_start"))
	return nil
}

// checkKKEngineNetwork checks if kkengine_net docker network exists
func checkKKEngineNetwork() bool {
	cmd := exec.Command("docker", "network", "inspect", "kkengine_net")
	return cmd.Run() == nil
}

// runN8nStartInternal is called to start n8n after install
// This is a forward declaration - actual implementation in n8n_start.go
func runN8nStartInternal() error {
	// Placeholder - will be implemented in Phase 3
	pterm.Println()
	ui.ShowSuccess(ui.Msg("n8n_install_complete"))
	pterm.Println("  " + ui.Msg("n8n_run_start"))
	return nil
}

// validateN8nDomain validates domain format
func validateN8nDomain(s string) error {
	if s == "" {
		return nil // Empty allowed, defaults to localhost
	}
	if s == "localhost" {
		return nil
	}
	// Basic validation - must contain at least one dot for non-localhost
	if !strings.Contains(s, ".") && s != "localhost" {
		return errors.New(ui.Msg("error_invalid_domain"))
	}
	return nil
}
