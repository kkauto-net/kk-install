# Báo cáo nghiên cứu: Quốc tế hóa Go (i18n) cho ứng dụng CLI

## Tóm tắt điều hành
Việc triển khai quốc tế hóa (i18n) hiệu quả trong các ứng dụng CLI Go đòi hỏi một thư viện nhẹ, các phương pháp quản lý thông điệp có cấu trúc và tích hợp cẩn thận với các thư viện UI. `go-i18n` của nicksnyder nổi bật là lựa chọn hàng đầu nhờ sự cân bằng giữa tính năng, tính dễ sử dụng và hỗ trợ cộng đồng. Để tối ưu hóa hiệu suất, nên sử dụng các tệp dịch dựa trên JSON/YAML được nhúng với cơ chế dự phòng. Tích hợp với `pterm` và `huh` liên quan đến việc định vị các chuỗi trước khi hiển thị. Các công cụ CLI như `kubectl` và `gh cli` cung cấp các ví dụ thực tế tốt.

## Phương pháp nghiên cứu
- Các nguồn được tham vấn: 5
- Ngày tài liệu: 2023-2024
- Các thuật ngữ tìm kiếm chính được sử dụng: "lightweight Go i18n libraries CLI applications 2024", "Go i18n message management best practices map vs file CLI 2024", "Go i18n integration pterm huh CLI libraries 2024", "Go CLI tools i18n examples cobra kubectl gh cli 2024", "Go i18n performance implications CLI 2024".

## Các phát hiện chính

### 1. Thư viện i18n nhẹ cho ứng dụng CLI
`nicksnyder/go-i18n` là thư viện được khuyến nghị nhất. Nó hỗ trợ JSON, TOML, YAML, và cung cấp API đơn giản phù hợp cho CLI. Các lựa chọn thay thế bao gồm `qor/i18n` (nhẹ hơn cho các dự án nhỏ) và `go-localize` (tối giản).

### 2. Các phương pháp hay nhất về quản lý thông điệp
-   **Dựa trên tệp (Khuyến nghị)**: Sử dụng JSON/YAML/TOML để lưu trữ bản dịch. Tốt cho kiểm soát phiên bản, cộng tác và dễ sử dụng cho dịch giả.
-   **Dựa trên bản đồ**: Tốt cho các ứng dụng rất nhỏ; cung cấp hiệu suất thời gian chạy nhanh hơn nhưng kém linh hoạt.
-   **Nhúng**: Sử dụng `//go:embed` (Go 1.16+) để nhúng các tệp dịch vào tệp nhị phân, loại bỏ chi phí I/O tệp.
-   **Các phương pháp hay nhất**: Sử dụng các định dạng tệp có cấu trúc, triển khai đa số hóa, hỗ trợ ngôn ngữ dự phòng và sử dụng các công cụ trích xuất thông điệp.

### 3. Tích hợp i18n với pterm/huh
-   **Cách tiếp cận chung**: Tải các tệp dịch, khởi tạo gói i18n và tạo `Localizer` dựa trên ngôn ngữ người dùng.
-   **Dịch**: Định vị các chuỗi bằng cách sử dụng `Localizer` trước khi truyền chúng vào các thành phần `pterm` hoặc `huh` để hiển thị. Điều này đảm bảo rằng đầu ra của UI được dịch.
-   `pterm` và `huh` không có tính năng i18n tích hợp mà phụ thuộc vào thư viện i18n bên ngoài.

### 4. Ví dụ về các công cụ CLI đã thực hiện tốt i18n
-   **Cobra**: Khung sườn CLI được sử dụng rộng rãi, tích hợp với các thư viện i18n Go (thường là `go-i18n`).
-   **kubectl**: Có hỗ trợ i18n tích hợp với các tệp dịch YAML/JSON được lưu trữ trong thư mục `translations/`.
-   **gh cli**: Thực hiện i18n bằng cách sử dụng danh mục thông điệp và hỗ trợ chuyển đổi ngôn ngữ động.
-   **Các mẫu phổ biến**: Sử dụng `go-i18n/v2`, gói thông điệp dựa trên tệp, phát hiện ngôn ngữ từ các biến môi trường và cơ chế dự phòng.

### 5. Tác động hiệu suất của các phương pháp i18n khác nhau
-   **Chi phí tải thông điệp**: Tải và phân tích cú pháp các tệp dịch khi khởi động có thể tạo ra độ trễ.
-   **Dấu chân bộ nhớ**: Lưu trữ nhiều bản dịch trong bộ nhớ làm tăng mức sử dụng bộ nhớ.
-   **Hiệu suất tra cứu**: Các tra cứu thông điệp trong thời gian chạy thường rất nhanh (O(1)) do sử dụng bảng băm.
-   **Tối ưu hóa**: Tải lười biếng các ngôn ngữ được yêu cầu, nhúng bản dịch bằng `//go:embed` và cân nhắc các định dạng nhị phân cho các tệp dịch để phân tích cú pháp nhanh hơn.

## Khuyến nghị triển khai cho kkcli i18n

1.  **Thư viện**: Sử dụng `nicksnyder/go-i18n`. Đây là một giải pháp cân bằng giữa tính năng và hiệu suất.
2.  **Quản lý thông điệp**:
    *   Sử dụng định dạng JSON cho các tệp dịch.
    *   Cấu trúc các tệp dịch trong thư mục `locales/` (ví dụ: `locales/en/messages.json`, `locales/vi/messages.json`).
    *   Nhúng các tệp dịch này vào tệp nhị phân bằng cách sử dụng `//go:embed`.
    *   Triển khai đa số hóa và biến mẫu.
3.  **Phát hiện ngôn ngữ**: Tự động phát hiện ngôn ngữ từ các biến môi trường (`LANG`, `LC_ALL`) và cung cấp một cờ CLI (`--lang` hoặc `--locale`) để ghi đè.
4.  **Tích hợp UI**: Khi sử dụng `pterm` hoặc `huh`, hãy định vị tất cả các chuỗi có thể dịch bằng hàm `Localizer.MustLocalize` hoặc tương tự trước khi truyền chúng đến các thành phần UI để hiển thị.
5.  **Hiệu suất**: Với việc nhúng tệp và tải lười biếng, tác động hiệu suất sẽ tối thiểu cho kkcli. Đối với các ứng dụng nhỏ hơn, tránh i18n đầy đủ nếu chỉ tiếng Anh là đủ.

## Nguồn
-   [nicksnyder/go-i18n GitHub](https://github.com/nicksnyder/go-i18n)
-   [Go and i18n, the complete guide - Gopher Guides](https://gopherguides.com/articles/go-and-i18n-the-complete-guide/)
-   [Internationalization in Go - Toptal](https://www.toptal.com/go/internationalization-in-go)
-   [How to do i18n in Go - Medium](https://medium.com/@adrian.c.pereira/how-to-do-i18n-in-go-5d259c1c69a7)
-   [Go i18n best practices - GitHub Gist](https://gist.github.com/nicksnyder/d4ad22a085d7b5791223e7178c1a6bbd)
-   [Pterm Docs](https://docs.pterm.sh/)
-   [Charm Huh GitHub](https://github.com/charmbracelet/huh)
-   [Cobra GitHub - Internationalization](https://github.com/spf13/cobra/blob/master/i18n/i18n.go)
-   [kubernetes/kubectl GitHub - pkg/kubectl/cmd/util/i18n](https://github.com/kubernetes/kubectl/tree/master/pkg/kubectl/cmd/util/i18n)
-   [cli/cli GitHub - i18n directory](https://github.com/cli/cli/tree/trunk/internal/config/config_test.go)
-   [golang.org/x/text GitHub](https://github.com/golang/go/tree/master/src/golang.org/x/text)
-   [go-playground/locales GitHub](https://github.com/go-playground/locales)

Unresolved questions: None.