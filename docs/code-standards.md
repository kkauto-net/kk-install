# Code Standards and Structure - kkcli

## 1. Codebase Structure

The `kkcli` project adheres to a clear and maintainable Go project structure to promote consistency, scalability, and ease of understanding.

### Top-Level Directories:

-   `cmd/`: Contains application-specific commands. Each top-level command should have its own package within this directory (e.g., `cmd/init`, `cmd/up`).
-   `pkg/`: Contains reusable library code that can be imported by external applications. This directory is further subdivided into functional areas.
    -   `pkg/compose/`: Logic for interacting with Docker Compose, abstracting common operations.
    -   `pkg/validator/`: Implements pre-flight validation checks.
    -   `pkg/templates/`: Manages embedded template files for configuration generation.
    -   `pkg/ui/`: Contains code related to user interface elements, prompts, and output formatting.
    -   `pkg/monitor/`: (Future) Functionality for monitoring Docker Compose services.
-   `docs/`: All project documentation, including PDRs, architecture, and code standards.
-   `example/`: Sample configuration files and usage examples.
-   `plans/`: Project plans, reports, and research documents.
-   `templates/`: Raw template files used by `pkg/templates`.
-   `go.mod`, `go.sum`: Go module definition and dependency manifests.
-   `main.go`: The main entry point of the `kkcli` application.
-   `README.md`: Project overview and quick start guide.

## 2. Code Standards and Conventions

### Go Language Best Practices:

-   **Formatting**: Use `go fmt` to automatically format code.
-   **Linting**: Adhere to `go vet` and `golangci-lint` recommendations.
-   **Naming Conventions**:
    -   **Packages**: Lowercase, single-word names (e.g., `validator`, `compose`). Avoid underscores.
    -   **Variables**: CamelCase for exported variables, `_` for unexported local variables. Short, descriptive names are preferred.
    -   **Functions/Methods**: CamelCase. Exported functions start with an uppercase letter, unexported with a lowercase.
    -   **Constants**: ALL_CAPS for global constants, CamelCase for package-level constants.
-   **Error Handling**:
    -   Return errors as the last return value.
    -   Use `fmt.Errorf` for wrapping errors with additional context.
    -   Avoid `panic` for recoverable errors.
    -   Handle errors explicitly; do not ignore them.
-   **Documentation**:
    -   All exported types, functions, and variables must have clear doc comments.
    -   Package-level comments should explain the purpose of the package.
-   **Concurrency**:
    -   Use `sync.WaitGroup` for managing goroutines.
    -   Prefer channels for communication between goroutines.
    -   Avoid shared memory by communicating; do not communicate by sharing memory.
-   **Testing**:
    -   Write unit tests for all significant logic.
    -   Test files should be named `_test.go` (e.g., `config_test.go`).
    -   Use `go test -cover` to check test coverage.

### Specific to `kkcli` Project:

-   **Modularity**: Keep packages focused on a single responsibility.
-   **Dependency Injection**: Where appropriate, use dependency injection to facilitate testing and flexibility.
-   **Configuration**: Centralize configuration management and validation.
-   **User Feedback**: Provide clear, actionable feedback to users, especially during validation and error scenarios.

## 3. Security Protocols and Compliance

-   **Input Validation**: All user inputs and external data must be thoroughly validated to prevent injection attacks and unexpected behavior.
-   **Secrets Management**: Avoid hardcoding sensitive information. Use environment variables or secure configuration mechanisms.
-   **Least Privilege**: Components should operate with the minimum necessary permissions.
-   **Dependency Audits**: Regularly audit third-party dependencies for known vulnerabilities.
-   **Error Masking**: Avoid exposing sensitive internal details in error messages returned to users. Log full errors internally if necessary.

## 4. API Design Guidelines (Internal APIs)

-   **Clarity**: API names should clearly indicate their purpose.
-   **Consistency**: Follow consistent naming and parameter conventions across the API.
-   **Simplicity**: Keep API surface areas minimal and easy to use.
-   **Backward Compatibility**: Strive to maintain backward compatibility for public APIs.
-   **Idiomatic Go**: Design APIs that feel natural to Go developers.
