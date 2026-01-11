package ui

var messagesEN = map[string]string{
	// Docker validation
	"checking_docker":       "Checking Docker...",
	"docker_ok":             "Docker is ready",
	"docker_not_installed":  "Docker is not installed",
	"docker_not_running":    "Docker daemon is not running",
	"docker_not_found":      "Docker Not Found",
	"docker_daemon_stopped": "Docker Not Running",
	"docker_compose_issue":  "Docker Compose Issue",
	"docker_required":       "Docker is required to continue",

	// Docker auto-install
	"ask_install_docker":      "Install Docker automatically?",
	"ask_install_docker_desc": "This will download and install Docker using the official script",
	"yes_install":             "Yes, install Docker",
	"no_manual":               "No, I'll install manually",
	"installing_docker":       "Installing Docker...",
	"docker_installed":        "Docker installed successfully",
	"docker_install_failed":   "Docker installation failed",
	"ask_start_docker":        "Start Docker daemon?",
	"starting_docker":         "Starting Docker daemon...",
	"docker_started":          "Docker daemon started",
	"docker_start_failed":     "Failed to start Docker",
	"yes":                     "Yes",

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
	"preflight_failed":        "Preflight checks failed. Please fix the errors above",
	"preflight_checks_failed": "One or more preflight checks failed",
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
	"no_services_defined": "No services defined in docker-compose.yml",
	"run_init":           "Run: kk init",
	"all_stopped":        "All services stopped",
	"status_summary_stopped": "%d services stopped\nTo start KKEngine, run: kk start",
	"all_running":             "All %d services running",
	"some_running":            "%d/%d services running",
	"get_status_failed":       "Failed to get status",
	"start_summary_success":   "All %d services started successfully",
	"start_summary_partial":   "%d/%d services started\nCheck logs: docker compose logs",
	"restart_summary_success": "All %d services restarted successfully",
	"restart_summary_partial": "%d/%d services restarted\nCheck logs: docker compose logs",
	"update_summary_success":  "All %d services updated successfully",
	"update_summary_partial":  "%d/%d services updated\nCheck logs: docker compose logs",

	// Table columns
	"service_status": "Service Status",
	"access_info":    "Access Information",
	"col_service":    "Service",
	"col_status":     "Status",
	"col_health":     "Health",
	"col_ports":      "Ports",
	"col_url":        "URL",
	"col_setting":    "Setting",
	"col_value":      "Value",

	// Init summary
	"config_summary": "Configuration Summary",
	"created_files":  "Created Files",
	"enabled":        "Enabled",
	"disabled":       "Disabled",
	"domain":         "Domain",

	// Status display
	"status_running":  "Running",
	"status_stopped":  "Stopped",
	"status_starting": "Starting",
	"summary":         "Summary",

	// Init wizard steps
	"step_docker_check": "Docker Check",
	"step_language":     "Language Selection",
	"step_options":      "Configuration Options",
	"step_domain":       "Domain Configuration",
	"step_credentials":  "Environment Configuration",
	"step_generate":     "Generate Files",
	"step_complete":     "Complete",

	// Credentials / Environment Configuration
	"ask_use_random":      "Use auto-generated secrets?",
	"ask_use_random_desc": "Secure random secrets have been generated for all fields",
	"no_edit":             "No, let me edit",
	"group_system":        "System Configuration",
	"group_db_secrets":    "Database Secrets",
	"group_s3_secrets":    "S3 Storage Secrets",
	"error_jwt_secret":    "Failed to generate JWT secret",
	"error_s3_access_key": "Failed to generate S3 access key",
	"error_s3_secret_key": "Failed to generate S3 secret key",

	// Force mode messages
	"docker_not_installed_force_init":      "Docker not installed (force mode - continuing)",
	"docker_daemon_not_running_force_init": "Docker daemon not running (force mode - continuing)",
	"docker_compose_issue_force_init":      "Docker Compose issue detected (force mode - continuing)",
	"compose_exists_force_init":            "docker-compose.yml exists, overwriting in force mode",

	// Validation
	"error_invalid_domain": "Invalid domain format (use example.com or localhost)",

	// Preflight
	"check":  "Check",
	"result": "Result",

	// Start/Restart/Update steps
	"step_preflight":      "Preflight Checks",
	"step_start_services": "Start Services",
	"step_health_check":   "Health Check",
	"step_status":         "Status",
	"step_pull_images":    "Pull Images",
	"step_recreate":       "Recreate Containers",

	// Command banners
	"status_desc":  "Service Status",
	"init_desc":    "Docker Stack Initialization",
	"start_desc":   "Start All Services",
	"restart_desc": "Restart All Services",
	"update_desc":  "Pull & Recreate",

	// Error box
	"to_fix":   "To fix",
	"then_run": "Then run",

	// Table columns (new)
	"col_image":   "Image",
	"col_current": "Current",
	"col_new":     "New",
	"col_file":    "File",

	// Progress
	"starting":         "starting...",
	"ready":            "ready",
	"unhealthy":        "unhealthy",
	"services_started": "Services started",
}
