package ui

var messagesVI = map[string]string{
	// Docker validation
	"checking_docker":      "Dang kiem tra Docker...",
	"docker_ok":            "Docker da san sang",
	"docker_not_installed": "Docker chua cai dat",
	"docker_not_running":   "Docker daemon khong chay",

	// Init flow
	"init_in_dir":    "Khoi tao trong: %s",
	"compose_exists": "docker-compose.yml da ton tai. Ghi de?",
	"init_cancelled": "Huy khoi tao",

	// Prompts
	"enable_seaweedfs": "Bat SeaweedFS file storage?",
	"seaweedfs_desc":   "SeaweedFS la he thong luu tru file phan tan",
	"enable_caddy":     "Bat Caddy web server?",
	"caddy_desc":       "Caddy la reverse proxy voi tu dong HTTPS",
	"enter_domain":     "Nhap domain (vd: example.com):",
	"yes_recommended":  "Yes (recommended)",
	"no":               "No",

	// Errors
	"error_db_password":   "Khong the tao password DB",
	"error_db_root_pass":  "Khong the tao password DB root",
	"error_redis_pass":    "Khong the tao password Redis",
	"error_create_file":   "Loi khi tao file",

	// Success
	"created":       "Da tao: %s",
	"init_complete": "Khoi tao hoan tat!",

	// Next steps
	"next_steps": `
Buoc tiep theo:
  1. Kiem tra va chinh sua .env neu can
  2. Chay: kk start
`,

	// Language selection
	"select_language": "Chon ngon ngu / Select language",
	"lang_english":    "English",
	"lang_vietnamese": "Tieng Viet",
}
