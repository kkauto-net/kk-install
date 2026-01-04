# System Architecture - kkcli

## 1. High-Level Architecture

The `kkcli` is a command-line interface (CLI) tool designed to streamline the management of Docker Compose applications. Its architecture is organized into several key layers, promoting modularity, maintainability, and extensibility.

```mermaid
graph TD
    User --> CLI_Interface;

    CLI_Interface --> Command_Handler;

    Command_Handler --> Project_Initializer;
    Command_Handler --> Validation_Layer;
    Command_Handler --> Configuration_Manager;
    Command_Handler --> Template_Engine;
    Command_Handler --> Docker_Compose_Wrapper;
    Command_Handler --> Monitoring_Service;

    Project_Initializer --> Template_Engine;
    Project_Initializer --> Configuration_Manager;

    Validation_Layer --> Config_Validator;
    Validation_Layer --> Env_Validator;
    Validation_Layer --> Disk_Validator;
    Validation_Layer --> Port_Validator;

    Configuration_Manager --> File_System;
    Template_Engine --> File_System;
    Docker_Compose_Wrapper --> Docker_Daemon;
    Monitoring_Service --> Docker_Daemon;

    subgraph Core Components
        Command_Handler
        Project_Initializer
        Configuration_Manager
        Template_Engine
        Docker_Compose_Wrapper
        Monitoring_Service
    end

    subgraph User Interaction
        CLI_Interface
    end

    subgraph External Services
        Docker_Daemon
        File_System
    end

    subgraph Validation Subsystem
        Validation_Layer
        Config_Validator
        Env_Validator
        Disk_Validator
        Port_Validator
    end
```

## 2. Architectural Layers

### 2.1. CLI Interface (`cmd/`, `pkg/ui/`)

-   **Responsibility**: Handles user input, parses commands and arguments, and presents output.
-   **Components**:
    -   `main.go`: Application entry point.
    -   `cmd/`: Defines the available CLI commands (e.g., `init`, `up`).
    -   `pkg/ui/`: Provides utilities for interactive prompts, password handling, and formatted output.

### 2.2. Command Handler (`cmd/`, `main.go`)

-   **Responsibility**: Routes parsed commands to the appropriate business logic components.
-   **Interaction**: Acts as an orchestrator, coordinating calls to various services based on the executed command.

### 2.3. Project Initializer (`pkg/init/` - conceptual)

-   **Responsibility**: Sets up new `kkcli` projects, generating necessary configuration files and directory structures.
-   **Interaction**: Utilizes the Template Engine and Configuration Manager.

### 2.4. Validation Layer (`pkg/validator/`)

-   **Responsibility**: Ensures the operating environment and project configurations meet predefined requirements before critical operations are executed.
-   **Components**:
    -   `preflight.go`: Orchestrates and executes various validation checks.
    -   `config.go`: Validates `kkcli` and Docker Compose configuration files.
    -   `env.go`: Checks for required environment variables.
    -   `disk.go`: Verifies disk space and file system permissions.
    -   `ports.go`: Detects port conflicts.
    -   `errors.go`: Custom error handling for validation failures.

### 2.5. Configuration Manager (`pkg/config/` - conceptual)

-   **Responsibility**: Manages loading, parsing, and persisting `kkcli`'s configuration settings.
-   **Interaction**: Provides structured access to configuration data for other components.

### 2.6. Template Engine (`pkg/templates/`)

-   **Responsibility**: Generates various configuration files (e.g., `docker-compose.yml`, Caddyfile) from templates.
-   **Components**:
    -   `embed.go`: Embeds template files into the Go binary.
    -   Specific `.tmpl` files: Define the structure and content of generated configurations.

### 2.7. Docker Compose Wrapper (`pkg/compose/`)

-   **Responsibility**: Provides an abstraction layer for interacting with the underlying `docker-compose` CLI tool.
-   **Interaction**: Executes Docker Compose commands and handles their output.

### 2.8. Monitoring Service (`pkg/monitor/` - conceptual)

-   **Responsibility**: (Future) Collects and displays metrics or logs from Docker Compose services.
-   **Interaction**: Potentially interacts directly with the Docker Daemon API.

## 3. Data Flow

1.  **User Input**: A user executes a `kkcli` command in the terminal.
2.  **CLI Parsing**: The `CLI Interface` (using `main.go` and `cmd/`) parses the command and its arguments.
3.  **Command Execution**: The `Command Handler` receives the parsed command.
4.  **Pre-flight Validation**: For critical operations, the `Command Handler` invokes the `Validation Layer`.
    -   The `Validation Layer` runs checks (config, env, disk, ports) using its internal components.
    -   Validation results (success/failure, detailed errors) are returned.
5.  **Logic Execution**: If validation passes (or if the command doesn't require validation), the `Command Handler` calls the relevant core components:
    -   `Project Initializer`: for `init` commands.
    -   `Docker Compose Wrapper`: for `up`, `down`, `logs`, etc.
    -   `Configuration Manager`: for reading/writing `kkcli` settings.
    -   `Template Engine`: for generating files.
6.  **External Interaction**: The `Docker Compose Wrapper` interacts with the `Docker Daemon` to manage containers. The `Configuration Manager` and `Template Engine` interact with the `File System`.
7.  **Output**: Results, status messages, or errors are formatted by `pkg/ui/` and presented back to the user via the `CLI Interface`.

## 4. Key Design Principles

-   **Modularity**: Clear separation of concerns into distinct packages.
-   **Extensibility**: Easy to add new commands, validation checks, or template types.
-   **User-Centricity**: Prioritizing clear feedback and ease of use.
-   **Robustness**: Emphasizing validation and comprehensive error handling.
-   **Idempotency**: Operations should produce the same result if applied multiple times (where applicable).