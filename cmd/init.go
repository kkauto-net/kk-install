package cmd

import (
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

	// Step 2: Get working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("\nKhoi tao trong: %s\n\n", cwd)

	// Step 3: Check if already initialized
	composePath := filepath.Join(cwd, "docker-compose.yml")
	if _, err := os.Stat(composePath); err == nil {
		var overwrite bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("docker-compose.yml da ton tai. Ghi de?").
					Value(&overwrite),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		if !overwrite {
			return fmt.Errorf("huy khoi tao")
		}
	}

	// Step 4: Interactive prompts
	enableSeaweedFS := true // Default: enabled (recommended)
	enableCaddy := true     // Default: enabled (recommended)
	var domain string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Bat SeaweedFS file storage?").
				Description("SeaweedFS la he thong luu tru file phan tan").
				Affirmative("Yes (recommended)").
				Negative("No").
				Value(&enableSeaweedFS),

			huh.NewConfirm().
				Title("Bat Caddy web server?").
				Description("Caddy la reverse proxy voi tu dong HTTPS").
				Affirmative("Yes (recommended)").
				Negative("No").
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
					Title("Nhap domain (vd: example.com):").
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

	// Step 5: Generate passwords
	dbPass, err := ui.GeneratePassword(24)
	if err != nil {
		return fmt.Errorf("khong the tao password DB: %w", err)
	}
	dbRootPass, err := ui.GeneratePassword(24)
	if err != nil {
		return fmt.Errorf("khong the tao password DB root: %w", err)
	}
	redisPass, err := ui.GeneratePassword(24)
	if err != nil {
		return fmt.Errorf("khong the tao password Redis: %w", err)
	}

	// Step 6: Render templates
	cfg := templates.Config{
		EnableSeaweedFS: enableSeaweedFS,
		EnableCaddy:     enableCaddy,
		DBPassword:      dbPass,
		DBRootPassword:  dbRootPass,
		RedisPassword:   redisPass,
		Domain:          domain,
	}

	if err := templates.RenderAll(cfg, cwd); err != nil {
		return fmt.Errorf("loi khi tao file: %w", err)
	}

	// Step 7: Show success
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
