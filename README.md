# kkcli
Muốn tạo CLI để giúp đỡ user tùy chỉnh docker-compose và quản lý
- kk init: Khởi tạo, cấu hình, tùy chỉnh docker-compose.yml và .env
- kk start: chạy docker-compose, sau đó hiển thị kết quả (Docker chạy nền)
- kk restart: restart docker-compose
- kk update: update docker kkengine chính và các docker khác nếu server có update
- kk status: xem trạng thái dịch vụ đang chạy



Các docker-compose.yml gồm cách thành phần
- kkengine: Docker chứa các container chính để chạy dịch vụ chính của KK
- Mariadb: Database chính kkengine sử dụng
- Redis: Cache và lưu trữ các session
- seaweedfs: SeaweedFS để lưu trữ các file (Optional)
- caddy: Caddy để chạy web server (Optional)



network:  kkengine_net
- bridge: Docker sẽ tạo một network bridge và các container sẽ được kết nối vào network này
- nếu db, redis cấu hình riêng thì không cần cấu hình network



Các file config sẵn ở /example
- docker-compose.yml
- .env
- Caddyfile
- kkfiler.toml
- kkphp.conf

## Yêu cầu

Để sử dụng `kkcli`, bạn cần cài đặt Docker và Docker Compose trên hệ thống của mình.

-   **Docker**: Đảm bảo Docker đã được cài đặt và đang chạy.
-   **Docker Compose**: `kkcli` sử dụng Docker Compose để quản lý các dịch vụ.

## Cài đặt

1.  Tải về `kkcli` executable:
    ```bash
    # Ví dụ: Tải về phiên bản mới nhất từ GitHub releases
    # curl -L "https://github.com/kkauto-net/kk-install/releases/latest/download/kkcli-linux-amd64" -o /usr/local/bin/kkcli
    # chmod +x /usr/local/bin/kkcli
    ```
    (Thay đổi URL và tên file phù hợp với hệ điều hành của bạn)

2.  Đảm bảo `kkcli` có quyền thực thi và nằm trong PATH của bạn.

## Sử dụng lệnh `kk init`

Lệnh `kk init` giúp bạn khởi tạo và cấu hình môi trường Docker Compose cho dự án của mình.

1.  **Chạy lệnh khởi tạo**:
    ```bash
    kk init
    ```

2.  **Trả lời các câu hỏi cấu hình**: `kk init` sẽ hỏi bạn một số thông tin để tạo file `docker-compose.yml` và `.env` phù hợp:
    -   Tên dịch vụ chính (ví dụ: `kkengine`)
    -   Cấu hình cơ sở dữ liệu (MySQL/MariaDB)
    -   Cấu hình Redis
    -   Có muốn sử dụng SeaweedFS không?
    -   Có muốn sử dụng Caddy làm web server không?
    -   Các cổng (ports) bạn muốn ánh xạ
    -   ... và các cấu hình khác.

3.  **Kiểm tra file cấu hình**: Sau khi hoàn tất, `kk init` sẽ tạo hoặc cập nhật các file `docker-compose.yml` và `.env` trong thư mục hiện tại. Hãy xem lại các file này để đảm bảo chúng đúng với mong muốn của bạn.

4.  **Khởi động dịch vụ**: Sau khi cấu hình xong, bạn có thể khởi động các dịch vụ bằng lệnh:
    ```bash
    kk start
    ```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 kkauto-net