# GitHub Actions Workflows

Dự án này sử dụng GitHub Actions để tự động hóa CI/CD pipeline.

## Workflows

### 1. CI (`ci.yml`)
**Trigger:** Push/PR đến branch `main`

**Jobs:**
- **test**: Chạy unit tests và build binary
- **lint**: Chạy golangci-lint để kiểm tra code quality

**Sử dụng:**
- Tự động chạy khi có push hoặc PR
- Đảm bảo code quality trước khi merge

---

### 2. Release (`release.yml`)
**Trigger:** Push tag theo pattern `v*.*.*` (ví dụ: `v0.1.0`)

**Jobs:**
- **goreleaser**: Build cross-platform binaries, tạo checksums, publish GitHub Release

**Sử dụng:**
```bash
# Tạo và push tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

**Output:**
- Multi-platform binaries (Linux/Darwin, amd64/arm64)
- Checksums file
- GitHub Release với artifacts

---

### 3. Draft Release (`draft-release.yml`)
**Trigger:** Manual workflow dispatch

**Jobs:**
- **draft-release**: Tạo draft release với changelog tự động

**Sử dụng:**
1. Vào tab "Actions" trên GitHub
2. Chọn "Draft Release" workflow
3. Click "Run workflow"
4. Nhập version (ví dụ: `v0.1.0`)
5. Review draft release và publish khi sẵn sàng

**Output:**
- Draft release với auto-generated changelog
- Installation instructions
- Full changelog link

---

### 4. Auto Version Bump (`auto-version.yml`)
**Trigger:** Khi PR được merge vào `main`

**Jobs:**
- **bump-version**: Tự động tạo tag dựa trên PR title

**Version Bump Rules:**
- **Major** (v1.0.0 → v2.0.0): PR title có `feat!:`, `feature!:`, hoặc `breaking:`
- **Minor** (v0.1.0 → v0.2.0): PR title có `feat:` hoặc `feature:`
- **Patch** (v0.1.0 → v0.1.1): Các PR khác (fix:, docs:, chore:, etc.)

**Ví dụ PR Titles:**
```
feat: add new Docker Compose manager      → v0.1.0 → v0.2.0
feat!: redesign CLI interface             → v0.1.0 → v1.0.0
fix: resolve port conflict issue          → v0.1.0 → v0.1.1
```

---

## Release Process

### Automatic Release (Recommended)
1. Tạo PR với conventional commit title
2. Merge PR → Auto version bump → Auto release

### Manual Release
1. Tạo draft release:
   ```bash
   # Via GitHub Actions UI
   Actions → Draft Release → Run workflow
   ```

2. Review và edit draft release

3. Tạo tag và publish:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

4. Release workflow sẽ tự động build và publish

---

## Conventional Commits

Để auto version bump hoạt động tốt, sử dụng conventional commits:

- `feat:` - New feature (minor bump)
- `fix:` - Bug fix (patch bump)
- `feat!:` - Breaking change (major bump)
- `docs:` - Documentation only
- `chore:` - Maintenance tasks
- `test:` - Test updates
- `refactor:` - Code refactoring

---

## Secrets Required

| Secret | Description | Required For |
|--------|-------------|--------------|
| `GITHUB_TOKEN` | Auto-provided by GitHub | All workflows |

---

## Troubleshooting

### Release workflow fails
- Check GoReleaser configuration (`.goreleaser.yml`)
- Verify tag format: `v{major}.{minor}.{patch}`
- Check Go version compatibility

### Auto version bump doesn't work
- Verify PR is merged (not closed)
- Check PR title follows conventional commits
- Ensure workflow has write permissions

### CI fails on tests
- Run tests locally: `go test ./...`
- Check Docker daemon is running (for integration tests)
- Verify Go version matches `go.mod`
