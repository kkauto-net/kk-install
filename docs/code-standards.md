# Code Standards and Structure

## Codebase Structure

The `kkcli` project follows a standard Go project layout, emphasizing modularity, testability, and maintainability.

```
kkcli/
├── .github/              # GitHub specific configurations (e.g., workflows)
│   └── workflows/
│       └── ci.yml        # CI/CD pipeline definition using GitHub Actions
├── cmd/                  # Main application commands
│   ├── kkcli.go          # Main entry point for the CLI
│   ├── update.go         # Implementation for the `kk update` command
│   └── completion.go     # Implementation for the `kk completion` command
├── docs/                 # Project documentation
│   ├── project-overview-pdr.md
│   ├── code-standards.md
│   ├── codebase-summary.md
│   └── system-architecture.md
├── pkg/                  # Reusable packages/libraries
│   └── updater/          # Package for handling application updates
│       ├── updater.go    # Core logic for checking and applying updates
│       └── updater_test.go # Unit tests for the updater package
├── scripts/              # Helper scripts
│   └── install.sh        # Secure installation script
├── vendor/               # Go module dependencies
├── Makefile              # Build automation script
├── .goreleaser.yml       # GoReleaser configuration for release automation
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
└── README.md             # Project README
```

### Key Directories and Their Purpose:

*   **`.github/workflows/`**: Contains YAML files defining GitHub Actions workflows for Continuous Integration and Continuous Deployment.
*   **`cmd/`**: Houses the main entry points for the `kkcli` application's commands. Each command typically has its own file for better organization.
*   **`docs/`**: Stores all project documentation, including design documents, architecture overviews, and code standards.
*   **`pkg/`**: Contains internal and external reusable packages. The `updater` package, for instance, encapsulates the logic for application updates.
*   **`scripts/`**: Holds utility scripts that aid in development, installation, or deployment processes. The `install.sh` script is a crucial part of the secure distribution strategy.

## Code Standards

### Go Naming Conventions

*   **Packages**: Use lowercase, single-word names. Avoid underscores or hyphens.
*   **Variables**:
    *   **Local**: camelCase.
    *   **Global/Exported**: PascalCase. Acronyms (like `HTTP`, `URL`, `ID`) should be all caps when exported (e.g., `HTTPRequest`, `AppID`).
*   **Functions**:
    *   **Internal**: camelCase.
    *   **Exported**: PascalCase. Should have a clear, descriptive name reflecting their action.
*   **Constants**: PascalCase for exported constants, camelCase for internal ones. Group related constants using `const ()`.
*   **Structs**: PascalCase for exported structs. Fields also use PascalCase if exported, camelCase if internal.
*   **Interfaces**: PascalCase, often ending with `-er` (e.g., `Reader`, `Writer`) if they define behavior.

### Error Handling

*   Errors are returned as the last return value and should be checked immediately.
*   Use `fmt.Errorf` for simple error messages.
*   Wrap errors using `fmt.Errorf("...: %w", err)` for context and traceability.
*   Avoid panics for recoverable errors; reserve panics for truly unrecoverable states.

### Testing

*   All packages and significant functions should have corresponding unit tests.
*   Test files are named `_test.go` (e.g., `updater_test.go`).
*   Test functions start with `Test` and use PascalCase (e.g., `TestUpdateCheck`).
*   Utilize Go's built-in `testing` package. Consider table-driven tests for multiple test cases.

### Documentation (GoDoc)

*   All exported types, functions, methods, and constants must have GoDoc comments.
*   Comments should be concise, clear, and explain the purpose, parameters, and return values.

### Dependency Management

*   Use Go Modules for dependency management.
*   Ensure `go.mod` and `go.sum` are committed to version control.
*   Avoid unnecessary dependencies.

## Build and Release Automation Specifics

### Makefile

*   The `Makefile` should define common development tasks: `build`, `test`, `clean`, `install`, etc.
*   Commands within the Makefile should be idempotent where possible.
*   Utilize variables for paths and flags to ensure easy configuration.

### GoReleaser

*   The `.goreleaser.yml` file configures the release process.
*   It specifies targets for building binaries (OS, architecture), packaging (archives, installers), and publishing.
*   Checksum generation and signing are mandatory parts of the release process to ensure artifact integrity.

### GitHub Actions (CI/CD)

*   Workflows (`.github/workflows/ci.yml`) define automated processes for:
    *   **Build**: Compiling the application for various platforms.
    *   **Test**: Running unit and integration tests.
    *   **Lint/Static Analysis**: Ensuring code quality and adherence to standards.
    *   **Release**: Triggered on tag pushes, leveraging GoReleaser to publish new versions.
*   Environment variables, especially for secrets (e.g., `GITHUB_TOKEN`), should be managed securely within GitHub Secrets.
