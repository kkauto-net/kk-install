package ui

var messagesEN = map[string]string{
	// Docker validation
	"checking_docker":      "Checking Docker...",
	"docker_ok":            "Docker is ready",
	"docker_not_installed": "Docker is not installed",
	"docker_not_running":   "Docker daemon is not running",

	// Init flow
	"init_in_dir":    "Initializing in: %s",
	"compose_exists": "docker-compose.yml already exists. Overwrite?",
	"init_cancelled": "Initialization cancelled",

	// Prompts
	"enable_seaweedfs": "Enable SeaweedFS file storage?",
	"seaweedfs_desc":   "SeaweedFS is a distributed file storage system",
	"enable_caddy":     "Enable Caddy web server?",
	"caddy_desc":       "Caddy is a reverse proxy with automatic HTTPS",
	"enter_domain":     "Enter domain (e.g. example.com):",
	"yes_recommended":  "Yes (recommended)",
	"no":               "No",

	// Errors
	"error_db_password":  "Failed to generate DB password",
	"error_db_root_pass": "Failed to generate DB root password",
	"error_redis_pass":   "Failed to generate Redis password",
	"error_create_file":  "Failed to create file",

	// File generation
	"generating_files": "Generating configuration files...",
	"files_generated":  "Configuration files generated",

	// Success
	"created":       "Created: %s",
	"init_complete": "Initialization complete!",

	// Next steps
	"next_steps": `
Next steps:
  1. Review and edit .env if needed
  2. Run: kk start
`,

	// Next steps for box
	"next_steps_box": `Next steps:
  1. Review and edit .env if needed
  2. Run: kk start`,

	// Language selection
	"select_language": "Select language / Chọn ngôn ngữ",
	"lang_english":    "English",
	"lang_vietnamese": "Tiếng Việt",

	// Runtime messages
	"stopping":           "Stopping...",
	"preflight_checking": "Running preflight checks...",
	"preflight_failed":   "Preflight checks failed. Please fix the errors above",
	"starting_services":  "Starting services...",
	"start_failed":       "Start failed",
	"health_checking":    "Checking service health...",
	"health_failed":      "Cannot monitor health",
	"some_not_ready":     "Some services not ready. Check: kk status",
	"start_complete":     "Start complete!",
	"restarting":         "Restarting services...",
	"restart_failed":     "Restart failed",
	"restart_complete":   "Restart complete",
	"checking_updates":   "Checking for updates...",
	"pulling_images":     "Pulling images...",
	"pull_failed":        "Failed to pull images",
	"images_up_to_date":  "All images are up to date",
	"updates_available":  "Updates available:",
	"confirm_restart":    "Restart services with new images?",
	"update_cancelled":   "Update cancelled. Images downloaded, run 'kk restart' to apply",
	"recreating":         "Recreating with new images...",
	"recreate_failed":    "Recreate failed",
	"update_complete":    "Update complete!",
	"no_services":        "No services running",
	"run_start":          "Run: kk start",
	"all_running":        "All %d services running",
	"some_running":       "%d/%d services running",
	"get_status_failed":  "Failed to get status",
}
