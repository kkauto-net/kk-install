# Phase 2: Fix Docs Command Descriptions

## Context Links
- Parent: [plan.md](./plan.md)
- Phase 1: [phase-01-rewrite-readme.md](./phase-01-rewrite-readme.md)

## Overview
- **Priority**: P2
- **Status**: Pending
- **Description**: Replace Vietnamese command descriptions with English in docs files

## Key Insights
- 3 files contain `kk` command descriptions in Vietnamese
- Only command descriptions need English translation
- Rest of docs stay Vietnamese (as per user request)

## Requirements

### Scope
- Only translate `kk` command descriptions
- Keep other content in Vietnamese
- Maintain consistent command description style

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `/home/kkdev/kkcli/docs/project-overview-pdr.md` | Modify | Lines 15-20 |
| `/home/kkdev/kkcli/docs/codebase-summary.md` | Modify | Lines 13-18 |
| `/home/kkdev/kkcli/docs/system-architecture.md` | Modify | Lines 50, 87-90 |

## Implementation Steps

### Step 1: Edit `project-overview-pdr.md` (Lines 15-20)

**Before:**
```markdown
*   **`kk init`**: Khởi tạo ngăn xếp Docker bằng cách tạo các tệp `docker-compose.yml` và `.env` thông qua lời nhắc tương tác.
*   **`kk start`**: Thực hiện kiểm tra trước khi chạy, khởi động các dịch vụ Docker Compose và theo dõi tình trạng của chúng.
*   **`kk restart`**: Khởi động lại tất cả các dịch vụ Docker Compose.
*   **`kk status`**: Hiển thị trạng thái hiện tại của các container Docker.
*   **`kk update`**: Kéo các hình ảnh Docker mới nhất và tạo lại các container để cập nhật dịch vụ.
*   **`kk completion`**: Tạo tập lệnh tự động hoàn thành shell cho `kkcli`.
```

**After:**
```markdown
*   **`kk init`**: Initialize Docker stack by creating `docker-compose.yml` and `.env` files through interactive prompts.
*   **`kk start`**: Run preflight checks, start Docker Compose services, and monitor their health.
*   **`kk restart`**: Restart all Docker Compose services.
*   **`kk status`**: Display current status of Docker containers.
*   **`kk update`**: Pull latest Docker images and recreate containers to update services.
*   **`kk completion`**: Generate shell completion script for `kkcli`.
```

### Step 2: Edit `codebase-summary.md` (Lines 13-18)

**Before:**
```markdown
    *   `kk init`: Tương tác với người dùng để tạo các tệp cấu hình (ví dụ: `docker-compose.yml`, `.env`) bằng cách sử dụng `pkg/templates`.
    *   `kk start`: Thực hiện các kiểm tra trước khi chạy bằng `pkg/validator`, sau đó khởi động các dịch vụ Docker Compose thông qua `pkg/compose` và theo dõi trạng thái bằng `pkg/monitor`.
    *   `kk restart`: Khởi động lại các dịch vụ Docker Compose bằng `pkg/compose`.
    *   `kk status`: Hiển thị trạng thái của các container bằng cách sử dụng `pkg/compose` và `pkg/ui`.
    *   `kk update`: Kéo các hình ảnh Docker mới nhất bằng `pkg/compose` và có thể sử dụng `pkg/updater`.
    *   `kk completion`: Tạo script hoàn thành shell.
```

**After:**
```markdown
    *   `kk init`: Interact with user to create config files (`docker-compose.yml`, `.env`) using `pkg/templates`.
    *   `kk start`: Run preflight checks via `pkg/validator`, start Docker Compose services via `pkg/compose`, monitor status via `pkg/monitor`.
    *   `kk restart`: Restart Docker Compose services via `pkg/compose`.
    *   `kk status`: Display container status using `pkg/compose` and `pkg/ui`.
    *   `kk update`: Pull latest Docker images via `pkg/compose`, may use `pkg/updater`.
    *   `kk completion`: Generate shell completion script.
```

### Step 3: Edit `system-architecture.md` (Lines 50, 87-90)

**Line 50 - Before:**
```markdown
*   **`pkg/templates`** (Độc lập): Chủ yếu được sử dụng bởi lệnh `kk init` để tạo cấu hình mà không có sự phụ thuộc chặt chẽ vào các gói khác.
```

**Line 50 - After:**
```markdown
*   **`pkg/templates`** (Độc lập): Primarily used by `kk init` command to generate configs without tight coupling to other packages.
```

**Lines 87-90 - Before:**
```markdown
## 3. Luồng dữ liệu chính (Ví dụ: `kk start`)

1.  **Lệnh `kk start` được thực thi**: Người dùng gọi `kk start`.
2.  **Xác thực ban đầu**: Lệnh `start` gọi `pkg/validator` để thực hiện kiểm tra trước khi chạy (ví dụ: Docker đã cài đặt và chạy, không xung đột cổng).
```

**Lines 87-90 - After:**
```markdown
## 3. Luồng dữ liệu chính (Example: `kk start`)

1.  **`kk start` command executed**: User invokes `kk start`.
2.  **Xác thực ban đầu**: `start` command calls `pkg/validator` for preflight checks (e.g., Docker installed and running, no port conflicts).
```

## Todo List

- [ ] Edit `/home/kkdev/kkcli/docs/project-overview-pdr.md` - 6 command descriptions
- [ ] Edit `/home/kkdev/kkcli/docs/codebase-summary.md` - 6 command descriptions
- [ ] Edit `/home/kkdev/kkcli/docs/system-architecture.md` - 3 command references

## Success Criteria

- [ ] All `kk` command descriptions in English
- [ ] Descriptions match README command table
- [ ] No grammatical errors in translations

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Inconsistent wording | Low | Use exact same descriptions as README |
| Missing some occurrences | Low | Grep verified all occurrences |

## Security Considerations
- None - documentation only

## Next Steps
- Verify changes on GitHub
- Submit repo to goreportcard.com if not indexed
