package ui

var messagesVI = map[string]string{
	// Docker validation
	"checking_docker":       "Đang kiểm tra Docker...",
	"docker_ok":             "Docker đã sẵn sàng",
	"docker_not_installed":  "Docker chưa được cài đặt",
	"docker_not_running":    "Docker daemon không chạy",
	"docker_not_found":      "Không tìm thấy Docker",
	"docker_daemon_stopped": "Docker không chạy",
	"docker_compose_issue":  "Vấn đề Docker Compose",
	"docker_required":       "Cần Docker để tiếp tục",

	// Docker auto-install
	"ask_install_docker":      "Tự động cài đặt Docker?",
	"ask_install_docker_desc": "Sẽ tải và cài Docker bằng script chính thức",
	"yes_install":             "Có, cài Docker",
	"no_manual":               "Không, tôi sẽ tự cài",
	"installing_docker":       "Đang cài đặt Docker...",
	"docker_installed":        "Đã cài đặt Docker thành công",
	"docker_install_failed":   "Cài đặt Docker thất bại",
	"ask_start_docker":        "Khởi động Docker daemon?",
	"starting_docker":         "Đang khởi động Docker...",
	"docker_started":          "Đã khởi động Docker daemon",
	"docker_start_failed":     "Không thể khởi động Docker",
	"docker_start_suggestion": "Khởi động Docker và chạy lại init",
	"yes":                     "Có",

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
	"preflight_failed":        "Kiểm tra thất bại. Vui lòng sửa lỗi trên",
	"preflight_checks_failed": "Một hoặc nhiều kiểm tra thất bại",
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
	"no_services_defined": "Không có dịch vụ nào được định nghĩa trong docker-compose.yml",
	"run_init":           "Chạy: kk init",
	"all_stopped":        "Tất cả dịch vụ đã dừng",
	"status_summary_stopped": "%d dịch vụ đã dừng\nĐể khởi động KKEngine, chạy: kk start",
	"all_running":             "Tất cả %d dịch vụ đang chạy",
	"some_running":            "%d/%d dịch vụ đang chạy",
	"get_status_failed":       "Không lấy được trạng thái",
	"start_summary_success":   "Tất cả %d dịch vụ đã khởi động thành công",
	"start_summary_partial":   "%d/%d dịch vụ đã khởi động\nKiểm tra logs: docker compose logs",
	"restart_summary_success": "Tất cả %d dịch vụ đã khởi động lại thành công",
	"restart_summary_partial": "%d/%d dịch vụ đã khởi động lại\nKiểm tra logs: docker compose logs",
	"update_summary_success":  "Tất cả %d dịch vụ đã cập nhật thành công",
	"update_summary_partial":  "%d/%d dịch vụ đã cập nhật\nKiểm tra logs: docker compose logs",

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
	"config_summary":   "Tóm tắt cấu hình",
	"created_files":    "Các file đã tạo",
	"enabled":          "Bật",
	"disabled":         "Tắt",
	"domain":           "Tên miền",
	"install_location": "Vị trí cài đặt",
	"data_directories": "Thư mục dữ liệu",

	// Status display
	"status_running":  "Đang chạy",
	"status_stopped":  "Đã dừng",
	"status_starting": "Đang khởi động",
	"summary":         "Tóm tắt",

	// Init wizard steps
	"step_license":      "Xác thực License",
	"step_docker_check": "Kiểm tra Docker",
	"step_language":     "Chọn ngôn ngữ",
	"step_options":      "Tùy chọn cấu hình",
	"step_domain":       "Cấu hình domain",
	"step_credentials":  "Cấu hình môi trường",
	"step_generate":     "Tạo file",
	"step_complete":     "Hoàn tất",

	// License verification
	"enter_license":            "Nhập license key của bạn",
	"license_required":         "License key là bắt buộc",
	"license_invalid_format":   "Định dạng license không hợp lệ. Mong đợi: LICENSE-XXXXXXXXXXXXXXXX",
	"validating_license":       "Đang xác thực license...",
	"license_validated":        "Xác thực license thành công",
	"license_validation_failed": "Xác thực license thất bại",
	"license_check_key":        "Vui lòng kiểm tra license key và thử lại",

	// Credentials / Environment Configuration
	"ask_use_random":      "Sử dụng mật khẩu tự động tạo?",
	"ask_use_random_desc": "Các mật khẩu ngẫu nhiên an toàn đã được tạo sẵn",
	"no_edit":             "Không, để tôi chỉnh sửa",
	"group_system":        "Cấu hình hệ thống",
	"group_db_secrets":    "Mật khẩu Database",
	"group_s3_secrets":    "Mật khẩu S3 Storage",
	"error_jwt_secret":    "Không thể tạo JWT secret",
	"error_s3_access_key": "Không thể tạo S3 access key",
	"error_s3_secret_key": "Không thể tạo S3 secret key",

	// Force mode messages
	"docker_not_installed_force_init":      "Docker chưa cài đặt (force mode - tiếp tục)",
	"docker_daemon_not_running_force_init": "Docker daemon không chạy (force mode - tiếp tục)",
	"docker_compose_issue_force_init":      "Phát hiện vấn đề Docker Compose (force mode - tiếp tục)",
	"compose_exists_force_init":            "docker-compose.yml đã tồn tại, ghi đè trong force mode",

	// Validation
	"error_invalid_domain": "Định dạng domain không hợp lệ (dùng example.com hoặc localhost)",

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

	// Stop command
	"step_stop_services": "Dừng dịch vụ",
	"stopping_services":  "Đang dừng dịch vụ...",
	"stop_failed":        "Dừng thất bại",
	"stop_complete":      "Tất cả dịch vụ đã dừng",
	"stop_desc":          "Dừng tất cả dịch vụ",

	// Remove command
	"step_remove_services": "Xóa dịch vụ",
	"removing_services":    "Đang xóa containers và networks...",
	"remove_failed":        "Xóa thất bại",
	"remove_complete":      "Đã xóa tất cả containers và networks",
	"removed_containers":   "Đã xóa containers và networks",
	"removed_with_volumes": "Đã xóa containers, networks và volumes",
	"remove_desc":          "Xóa tất cả containers",

	// Container conflict errors
	"container_conflict_suggestion": "Containers từ project khác đang sử dụng cùng tên. Hãy xóa chúng trước",

	// Command banners
	"status_desc":  "Trạng thái dịch vụ",
	"init_desc":    "Khởi tạo Docker Stack",
	"start_desc":   "Khởi động tất cả dịch vụ",
	"restart_desc": "Khởi động lại tất cả dịch vụ",
	"update_desc":  "Cập nhật & Khởi tạo lại",

	// Error box
	"to_fix":   "Để khắc phục",
	"then_run": "Sau đó chạy",

	// Table columns (new)
	"col_image":   "Image",
	"col_current": "Hiện tại",
	"col_new":     "Mới",
	"col_file":    "Tệp",

	// Progress
	"starting":         "đang khởi động...",
	"ready":            "sẵn sàng",
	"unhealthy":        "không khỏe mạnh",
	"services_started": "Đã khởi động dịch vụ",

	// Selfupdate
	"step_check_update":     "Kiểm tra cập nhật",
	"step_install_update":   "Cài đặt cập nhật",
	"checking_cli_update":   "Đang kiểm tra cập nhật...",
	"check_update_failed":   "Không thể kiểm tra cập nhật",
	"current_version":       "Phiên bản hiện tại",
	"latest_version":        "Phiên bản mới nhất",
	"cli_up_to_date":        "kk CLI đã là phiên bản mới nhất",
	"update_available":      "Có phiên bản mới!",
	"to_update_run":         "Để cập nhật, chạy",
	"confirm_cli_update":    "Cập nhật lên phiên bản mới nhất?",
	"selfupdate_cancelled":  "Đã hủy cập nhật",
	"downloading_update":    "Đang tải và cài đặt...",
	"update_install_failed": "Không thể cài đặt cập nhật",
	"selfupdate_complete":   "Cập nhật thành công!",
	"selfupdate_desc":       "Cập nhật kk CLI",

	// Project configuration
	"project_not_configured":  "Chưa cấu hình Project",
	"run_init_to_configure":   "Chạy 'kk init' trong thư mục project của bạn",
	"config_load_failed":      "Không thể tải cấu hình",
	"config_title":            "Cấu hình:",
	"config_language":         "Ngôn ngữ",
	"config_project_dir":      "Thư mục Project",
	"config_file_path":        "File cấu hình",
	"config_not_set":          "(chưa thiết lập)",
}
