package n8n

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// RenderTemplate renders a single template file to the specified output path.
// The output directory must already exist.
func RenderTemplate(name string, cfg N8nConfig, outputPath string) error {
	tmplContent, err := templateFS.ReadFile("templates/" + name + ".tmpl")
	if err != nil {
		return fmt.Errorf("read template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", name, err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", outputPath, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, cfg); err != nil {
		return fmt.Errorf("execute template %s: %w", name, err)
	}
	return nil
}

// RenderAll renders all templates to the n8n directory.
// It validates the config, creates required directories, and renders all template files.
func RenderAll(cfg N8nConfig) error {
	// Validate config before rendering
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := EnsureDirectories(); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	files := map[string]string{
		"docker-compose.yml": "docker-compose.yml",
		"env":                ".env",
	}

	for tmplName, outputName := range files {
		outputPath := filepath.Join(N8nDir(), outputName)
		if err := RenderTemplate(tmplName, cfg, outputPath); err != nil {
			return err
		}
	}

	// Set .env permissions to 0600 (owner read/write only)
	envPath := filepath.Join(N8nDir(), ".env")
	if err := os.Chmod(envPath, 0600); err != nil {
		return fmt.Errorf("set .env permissions: %w", err)
	}
	return nil
}
