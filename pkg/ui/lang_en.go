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
	"error_db_password":   "Failed to generate DB password",
	"error_db_root_pass":  "Failed to generate DB root password",
	"error_redis_pass":    "Failed to generate Redis password",
	"error_create_file":   "Failed to create file",

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
	"select_language": "Select language / Chon ngon ngu",
	"lang_english":    "English",
	"lang_vietnamese": "Tieng Viet",
}
