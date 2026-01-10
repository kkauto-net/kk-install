package ui

var messagesVI = map[string]string{
	// Docker validation
	"checking_docker":      "Đang kiểm tra Docker...",
	"docker_ok":            "Docker đã sẵn sàng",
	"docker_not_installed": "Docker chưa được cài đặt",
	"docker_not_running":   "Docker daemon không chạy",

	// Init flow
	"init_in_dir":    "Khởi tạo trong: %s",
	"compose_exists": "docker-compose.yml đã tồn tại. Ghi đè?",
	"init_cancelled": "Hủy khởi tạo",

	// Prompts
	"enable_seaweedfs": "Bật SeaweedFS file storage?",
	"seaweedfs_desc":   "SeaweedFS là hệ thống lưu trữ file phân tán",
	"enable_caddy":     "Bật Caddy web server?",
	"caddy_desc":       "Caddy là reverse proxy với tự động HTTPS",
	"enter_domain":     "Nhập domain (vd: example.com):",
	"yes_recommended":  "Có (khuyến nghị)",
	"no":               "Không",

	// Errors
	"error_db_password":  "Không thể tạo mật khẩu DB",
	"error_db_root_pass": "Không thể tạo mật khẩu DB root",
	"error_redis_pass":   "Không thể tạo mật khẩu Redis",
	"error_create_file":  "Lỗi khi tạo file",

	// File generation
	"generating_files": "Đang tạo các file cấu hình...",
	"files_generated":  "Các file cấu hình đã được tạo",

	// Success
	"created":       "Đã tạo: %s",
	"init_complete": "Khởi tạo hoàn tất!",

	// Next steps
	"next_steps": `
Bước tiếp theo:
  1. Kiểm tra và chỉnh sửa .env nếu cần
  2. Chạy: kk start
`,

	// Next steps for box
	"next_steps_box": `Bước tiếp theo:
  1. Kiểm tra và chỉnh sửa .env nếu cần
  2. Chạy: kk start`,

	// Language selection
	"select_language": "Chọn ngôn ngữ / Select language",
	"lang_english":    "English",
	"lang_vietnamese": "Tiếng Việt",

	// Runtime messages (start, restart, update, status)
	"stopping":           "Đang dừng lại...",
	"preflight_checking": "Kiểm tra trước khi chạy...",
	"preflight_failed":   "Kiểm tra thất bại. Vui lòng sửa lỗi trên",
	"starting_services":  "Khởi động services...",
	"start_failed":       "Khởi động thất bại",
	"health_checking":    "Đang kiểm tra sức khỏe dịch vụ...",
	"health_failed":      "Không thể theo dõi health",
	"some_not_ready":     "Một số dịch vụ chưa sẵn sàng. Kiểm tra: kk status",
	"start_complete":     "Khởi động hoàn tất!",
	"restarting":         "Đang khởi động lại dịch vụ...",
	"restart_failed":     "Khởi động lại thất bại",
	"restart_complete":   "Đã khởi động lại",
	"checking_updates":   "Đang kiểm tra cập nhật...",
	"pulling_images":     "Đang tải images...",
	"pull_failed":        "Không tải được images",
	"images_up_to_date":  "Tất cả images đã là phiên bản mới nhất",
	"updates_available":  "Có cập nhật:",
	"confirm_restart":    "Khởi động lại services với images mới?",
	"update_cancelled":   "Hủy cập nhật. Images đã được tải, chạy 'kk restart' để áp dụng",
	"recreating":         "Đang khởi động lại với images mới...",
	"recreate_failed":    "Recreate thất bại",
	"update_complete":    "Cập nhật hoàn tất!",
	"no_services":        "Không có dịch vụ nào đang chạy",
	"run_start":          "Chạy: kk start",
	"all_running":        "Tất cả %d dịch vụ đang chạy",
	"some_running":       "%d/%d dịch vụ đang chạy",
	"get_status_failed":  "Không lấy được trạng thái",

	// Table columns
	"service_status": "Trạng thái dịch vụ",
	"access_info":    "Thông tin truy cập",
	"col_service":    "Dịch vụ",
	"col_status":     "Trạng thái",
	"col_health":     "Sức khỏe",
	"col_ports":      "Cổng",
	"col_url":        "URL",
	"col_setting":    "Cài đặt",
	"col_value":      "Giá trị",

	// Init summary
	"config_summary": "Tóm tắt cấu hình",
	"created_files":  "Các file đã tạo",
	"enabled":        "Bật",
	"disabled":       "Tắt",
	"domain":         "Tên miền",

	// Status display
	"status_running": "Đang chạy",
	"status_stopped": "Đã dừng",

	// Init wizard steps
	"step_docker_check": "Kiểm tra Docker",
	"step_language":     "Chọn ngôn ngữ",
	"step_options":      "Tùy chọn cấu hình",
	"step_generate":     "Tạo file",
	"step_complete":     "Hoàn tất",

	// Preflight
	"check":  "Kiểm tra",
	"result": "Kết quả",

	// Start/Restart/Update steps
	"step_preflight":      "Kiểm tra trước",
	"step_start_services": "Khởi động dịch vụ",
	"step_health_check":   "Kiểm tra sức khỏe",
	"step_status":         "Trạng thái",
	"step_pull_images":    "Tải images",
	"step_recreate":       "Tạo lại containers",
}
