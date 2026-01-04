### Tích hợp Docker và Chiến lược xác thực cho KK CLI (Go)

#### 1. Mẫu sử dụng Docker SDK cho Go
*   **Xác thực Daemon:** Sử dụng `client.NewClientWithOpts` với các tùy chọn như `client.WithHostFromEnv()` và `client.WithAPIVersionFromEnv()` để kết nối với Docker daemon. Xác minh kết nối bằng `cli.Ping()`.
*   **Kiểm tra Container:** `cli.ContainerInspect()` để lấy trạng thái chi tiết của container (ID, tên, trạng thái, port bindings).
*   **Health Checks:**
    *   Sử dụng `cli.ContainerList()` với bộ lọc để tìm các container có nhãn health check cụ thể.
    *   Theo dõi `State.Health.Status` từ `ContainerInspect` để kiểm tra trạng thái sức khỏe của container.
    *   Triển khai logic đợi/thử lại với timeout.
```go
import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
)

func validateDockerDaemon(ctx context.Context) (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.WithFromEnv(), client.WithAPIVersionFromEnv())
	if err != nil {
		return nil, fmt.Errorf("tạo client Docker thất bại: %w", err)
	}
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("kết nối đến Docker daemon thất bại: %w", err)
	}
	return cli, nil
}
```
**Nguồn:** [Docker SDK for Go documentation](https://docs.docker.com/engine/api/sdk/examples/)

#### 2. Phát hiện xung đột cổng
*   **Go `net.Listen`:** Cách đáng tin cậy và đa nền tảng nhất. Thử lắng nghe trên một cổng, nếu lỗi, cổng đó đang được sử dụng.
*   **Docker API:** `cli.ContainerList()` và `cli.ContainerInspect()` để kiểm tra port mappings của các container đang chạy.
*   **`lsof` (Unix/macOS) / `netstat` (Windows):** Thực thi các lệnh hệ thống này để kiểm tra, nhưng kém tin cậy hơn và không đa nền tảng.
```go
import (
	"fmt"
	"net"
	"time"
)

func isPortInUse(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return true // Port đang được sử dụng
	}
	defer conn.Close()
	return false
}
```
**Nguồn:** [Stack Overflow discussions on port checking in Go](https://stackoverflow.com/questions/39668101/how-to-check-if-a-port-is-listening-in-go)

#### 3. Các thực hành tốt nhất về Giám sát Health Check
*   **Chiến lược thử lại (Retry Strategies):** Sử dụng hàm thử lại với backoff theo cấp số nhân và jitter.
*   **Xử lý Timeout:** Luôn sử dụng `context.WithTimeout` cho các hoạt động health check.
*   **Phát hiện lỗi:** Phân biệt `liveness` (daemon đang chạy) và `readiness` (sẵn sàng phục vụ yêu cầu).
*   **Triển khai:** Health check nên nhanh, nhẹ. Tránh các hoạt động tốn kém.
```go
import (
	"context"
	"errors"
	"time"
)

func healthCheckWithRetry(ctx context.Context, checkFunc func(context.Context) error, retries int, delay time.Duration) error {
	for i := 0; i < retries; i++ {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // Timeout cho mỗi lần thử
		err := checkFunc(ctx)
		cancel()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("health check bị hủy hoặc hết thời gian chờ: %w", ctx.Err())
		case <-time.After(delay):
			delay *= 2 // Backoff theo cấp số nhân
		}
	}
	return errors.New("health check thất bại sau nhiều lần thử")
}
```
**Nguồn:** [Go health check monitoring best practices](https://www.youtube.com/watch?v=1FhG6BqW-vQ), [Google Cloud's health check guidelines](https://cloud.google.com/load-balancing/docs/health-checks)

#### 4. Các mẫu dịch lỗi trong Go
*   **Tách biệt:** Tách lỗi kỹ thuật (developer-facing) và thông báo thân thiện với người dùng (user-facing).
*   **Keys dịch:** Sử dụng các key thay vì trực tiếp các chuỗi lỗi để cho phép dịch.
*   **Thư viện `go-i18n`:** Hỗ trợ dịch tin nhắn, pluralization, định dạng.
```go
// Ví dụ về lỗi có thể dịch (translatable error)
type UserError struct {
	Key    string
	Params map[string]interface{}
}

func (e *UserError) Error() string {
	return e.Key // Trong thực tế, sẽ được dịch ở lớp trình bày
}

// Giả định có một hàm dịch
func translate(key string, params map[string]interface{}) string {
	// Logic dịch thực tế, ví dụ với go-i18n
	return fmt.Sprintf("Đã xảy ra lỗi: %s (params: %v)", key, params)
}

func handleError(err error) {
	var userErr *UserError
	if errors.As(err, &userErr) {
		fmt.Println("Thông báo người dùng:", translate(userErr.Key, userErr.Params))
	} else {
		fmt.Println("Lỗi nội bộ:", err.Error())
	}
}
```
**Nguồn:** [go-i18n GitHub repository](https://github.com/nicksnyder/go-i18n), [Internationalization in Go](https://phrase.com/blog/posts/internationalization-i18n-in-go/)

#### 5. Chiến lược phân phối
*   **GitHub Releases:** Phổ biến nhất. Tạo bản phát hành với các static binary cho nhiều kiến trúc và hệ điều hành.
*   **Install Scripts:** Cung cấp script tải xuống và cài đặt (ví dụ: `curl ... | bash`). Cần cẩn thận về bảo mật.
*   **Static Binary Builds:** Go tạo ra các binary độc lập không có phụ thuộc runtime, làm cho việc phân phối đơn giản.
*   **Homebrew/APT/RPM:** Để phân phối chuyên nghiệp hơn.
```bash
# Ví dụ tạo static binary cho Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kkcli-linux-amd64 .

# Ví dụ tạo static binary cho macOS
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o kkcli-darwin-amd64 .
```
**Nguồn:** [Go documentation on cross-compilation](https://go.dev/doc/install/source#environment), [GitHub Actions for Go releases](https://docs.github.com/en/actions/publishing-packages-to-github-packages/publishing-go-packages)

**Các câu hỏi chưa được giải quyết:**
*   Yêu cầu cụ thể về ngôn ngữ I18n nào cần được hỗ trợ?
*   Mức độ chi tiết của thông báo lỗi cho người dùng cần được xác định.
*   Có cần hỗ trợ phân phối qua các trình quản lý gói cụ thể nào ngoài GitHub Releases không?
