package templates

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed *.tmpl
var templateFS embed.FS // Force recompile

type Config struct {
	EnableSeaweedFS bool
	EnableCaddy     bool
	DBPassword      string
	DBRootPassword  string
	RedisPassword   string
	Domain          string
}

// RenderTemplate renders a single template file
func RenderTemplate(name string, cfg Config, outputPath string) error {
	tmplContent, err := templateFS.ReadFile(name + ".tmpl")
	if err != nil {
		return err
	}

	tmpl, err := template.New(name).Parse(string(tmplContent))
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Backup existing file if it exists
	if _, err := os.Stat(outputPath); err == nil {
		backupPath := outputPath + ".bak"
		if err := os.Rename(outputPath, backupPath); err != nil {
			return err
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, cfg)
}

// RenderAll renders all templates to the target directory
func RenderAll(cfg Config, targetDir string) error {
	files := map[string]string{
		"docker-compose.yml": "docker-compose.yml",
		"env":                ".env",
		"kkphp.conf":         "kkphp.conf",
	}

	if cfg.EnableCaddy {
		files["Caddyfile"] = "Caddyfile"
	}
	if cfg.EnableSeaweedFS {
		files["kkfiler.toml"] = "kkfiler.toml"
	}

	for tmplName, outputName := range files {
		outputPath := filepath.Join(targetDir, outputName)
		if err := RenderTemplate(tmplName, cfg, outputPath); err != nil {
			return err
		}
	}

	// Set .env permissions to 0600 (owner read/write only)
	envPath := filepath.Join(targetDir, ".env")
	if err := os.Chmod(envPath, 0600); err != nil {
		return err
	}

	return nil
}
