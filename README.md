<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-light.svg">
    <img src="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-light.svg" alt="TelemetryFlow Logo" width="80%">
  </picture>

  <h3>TelemetryFlow GO MCP Server (TFO-GO-MCP)</h3>

[![Version](https://img.shields.io/badge/Version-1.1.2-orange.svg)](CHANGELOG.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![MCP Protocol](https://img.shields.io/badge/MCP-2024--11--05-purple?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMTIgMkM2LjQ4IDIgMiA2LjQ4IDIgMTJzNC40OCAxMCAxMCAxMCAxMC00LjQ4IDEwLTEwUzE3LjUyIDIgMTIgMnoiIGZpbGw9IiNmZmYiLz48L3N2Zz4=)](https://modelcontextprotocol.io/)
[![Claude API](https://img.shields.io/badge/Claude-Opus%204%20%7C%20Sonnet%204-E1BEE7?logo=anthropic)](https://anthropic.com)
[![OTEL SDK](https://img.shields.io/badge/OpenTelemetry_SDK-1.39.0-blueviolet)](https://opentelemetry.io/)
[![Architecture](https://img.shields.io/badge/Architecture-DDD%2FCQRS-success)](docs/ARCHITECTURE.md)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql)](https://www.postgresql.org/)
[![ClickHouse](https://img.shields.io/badge/ClickHouse-23+-FFCC00?logo=clickhouse)](https://clickhouse.com/)

</div>

---

**Enterprise-Grade Model Context Protocol Server with Claude AI Integration**

A comprehensive MCP server implementation built using Go and following Domain-Driven Design (DDD) patterns, providing seamless integration between the Model Context Protocol and Anthropic's Claude AI.

This server works as the **AI integration layer** for the TelemetryFlow Platform, providing:

- Claude AI conversation capabilities via MCP
- Tool execution with built-in and custom tools
- Resource management and prompt templates
- TelemetryFlow SDK observability integration

---

## TelemetryFlow Ecosystem

```mermaid
graph LR
    subgraph "TelemetryFlow Ecosystem v1.1.2"
        subgraph "Instrumentation"
            SDK_GO[TFO-Go-SDK<br/>OTEL SDK v1.39.0]
            SDK_PY[TFO-Python-SDK<br/>OTEL SDK v1.28.0]
            SDK_OTHER[TFO-AnyStacks-SDK<br/>OTEL AnyStacks SDK]
        end
        subgraph "Collection"
            AGENT[TFO-Agent<br/>OTEL SDK v1.39.0]
        end
        subgraph "Processing"
            COLLECTOR[TFO-Collector<br/>OTEL v0.142.0]
        end
        subgraph "AI Integration"
            MCP_GO[TFO-Go-MCP<br/>Claude API + MCP]
            MCP_PY[TFO-Python-MCP<br/>Claude API + MCP]
        end
        subgraph "Platform"
            CORE[TFO-Core<br/>NestJS IAM v1.1.4]
        end
    end

    SDK_GO --> AGENT
    SDK_PY --> AGENT
    SDK_OTHER --> AGENT
    AGENT --> COLLECTOR
    COLLECTOR --> CORE
    MCP_GO --> CORE
    MCP_PY --> CORE
    MCP_GO -.-> |AI Capabilities| COLLECTOR
    MCP_PY -.-> |AI Capabilities| COLLECTOR

    style MCP_GO fill:#FFA1E1,stroke:#C989B4,stroke-width:5px
    style MCP_PY fill:#E1BEE7,stroke:#7B1FA2
    style SDK_GO fill:#C8E6C9,stroke:#388E3C
    style SDK_PY fill:#C8E6C9,stroke:#388E3C
    style SDK_OTHER fill:#DFDFDF,stroke:#0F0F0F
    style AGENT fill:#BBDEFB,stroke:#1976D2
    style COLLECTOR fill:#FFE0B2,stroke:#F57C00
    style CORE fill:#B3E5FC,stroke:#0288D1
```

| Component      | Version    | OTEL Base          | Role                          |
| -------------- | ---------- | ------------------ | ----------------------------- |
| TFO-Core       | v1.1.4     | -                  | Identity & Access Management  |
| TFO-Agent      | v1.1.2     | SDK v1.39.0        | Telemetry Collection Agent    |
| TFO-Collector  | v1.1.2     | Collector v0.142.0 | Central Telemetry Processing  |
| TFO-Go-SDK     | v1.1.2     | SDK v1.39.0        | Go Instrumentation            |
| TFO-Python-SDK | v1.1.2     | SDK v1.28.0        | Python Instrumentation        |
| **TFO-Go-MCP** | **v1.1.2** | **SDK v1.39.0**    | **GO MCP Server + Claude AI** |
| TFO-Python-MCP | v1.1.2     | SDK v1.28.0        | Python MCP Server + Claude AI |

---

## Quick Facts

| Property             | Value                                                   |
| -------------------- | ------------------------------------------------------- |
| **Version**          | 1.1.2                                                   |
| **Language**         | Go 1.24+                                                |
| **MCP Protocol**     | 2024-11-05                                              |
| **Claude SDK**       | anthropic-sdk-go v0.2.0-beta.3                          |
| **OTEL SDK**         | v1.39.0                                                 |
| **Architecture**     | DDD/CQRS                                                |
| **Transport**        | stdio, SSE (planned), WebSocket (planned)               |
| **Built-in Tools**   | 8 tools                                                 |
| **Supported Models** | Claude 4 Opus, Claude 4 Sonnet, Claude 3.5 Sonnet/Haiku |
| **Databases**        | PostgreSQL (GORM), ClickHouse, Redis (Cache)            |
| **Queue**            | NATS JetStream                                          |

---

## System Architecture

```mermaid
graph TB
    subgraph "Client Applications"
        CC[Claude Code]
        IDE[IDE Extensions]
        CLI[CLI Tools]
        CUSTOM[Custom MCP Clients]
    end

    subgraph "TFO-GO-MCP Server"
        subgraph "Presentation Layer"
            SERVER[MCP Server<br/>JSON-RPC 2.0]
            TOOLS[Built-in Tools]
            RESOURCES[Resources]
            PROMPTS[Prompts]
        end

        subgraph "Application Layer - CQRS"
            CMD[Commands]
            QRY[Queries]
            HANDLERS[Handlers]
        end

        subgraph "Domain Layer - DDD"
            AGG[Aggregates<br/>Session, Conversation]
            ENT[Entities<br/>Message, Tool, Resource]
            VO[Value Objects<br/>IDs, Content, Types]
            EVT[Domain Events]
            SVC[Domain Services]
        end

        subgraph "Infrastructure Layer"
            CLAUDE[Claude API Client]
            CONFIG[Configuration]
            REPO[Repositories]
            LOG[Logging]
            OTEL[OpenTelemetry]
        end
    end

    subgraph "External Services"
        ANTHROPIC[Anthropic Claude API]
        OTLP[OTLP Collector]
    end

    CC --> SERVER
    IDE --> SERVER
    CLI --> SERVER
    CUSTOM --> SERVER

    SERVER --> CMD
    SERVER --> QRY
    TOOLS --> HANDLERS
    RESOURCES --> HANDLERS
    PROMPTS --> HANDLERS

    HANDLERS --> AGG
    HANDLERS --> SVC
    AGG --> ENT
    AGG --> VO
    AGG --> EVT

    SVC --> CLAUDE
    HANDLERS --> REPO
    CONFIG --> SERVER
    LOG --> SERVER
    OTEL --> OTLP

    CLAUDE --> ANTHROPIC

    style SERVER fill:#E1BEE7,stroke:#7B1FA2,stroke-width:2px
    style CLAUDE fill:#FFCDD2,stroke:#C62828
    style ANTHROPIC fill:#FFCDD2,stroke:#C62828
    style AGG fill:#C8E6C9,stroke:#388E3C
    style HANDLERS fill:#BBDEFB,stroke:#1976D2
```

---

## MCP Protocol Data Flow

```mermaid
sequenceDiagram
    participant Client as MCP Client
    participant Server as TFO-GO-MCP Server
    participant Handler as Request Handler
    participant Domain as Domain Layer
    participant Claude as Claude API

    Note over Client,Claude: Session Initialization
    Client->>Server: initialize (JSON-RPC 2.0)
    Server->>Handler: InitializeSessionCommand
    Handler->>Domain: Create Session Aggregate
    Domain-->>Handler: Session Created + Events
    Handler-->>Server: Initialize Result
    Server-->>Client: capabilities, serverInfo

    Client->>Server: notifications/initialized
    Note over Server: Session Ready

    Note over Client,Claude: Tool Execution Flow
    Client->>Server: tools/list
    Server->>Handler: ListToolsQuery
    Handler->>Domain: Get Tools from Session
    Domain-->>Handler: Tool List
    Handler-->>Server: Tools Array
    Server-->>Client: {tools: [...]}

    Client->>Server: tools/call (claude_conversation)
    Server->>Handler: ExecuteToolCommand
    Handler->>Domain: Validate & Execute
    Domain->>Claude: CreateMessage Request
    Claude-->>Domain: Claude Response
    Domain-->>Handler: Tool Result
    Handler-->>Server: Content Array
    Server-->>Client: {content: [...]}
```

---

## Domain-Driven Design Architecture

```mermaid
graph TB
    subgraph "Domain Layer"
        subgraph "Aggregates"
            SESSION[Session Aggregate<br/>├─ ID, State, Capabilities<br/>├─ Tools, Resources, Prompts<br/>└─ Conversations]
            CONV[Conversation Aggregate<br/>├─ ID, Model, Status<br/>├─ Messages, Tools<br/>└─ Settings]
        end

        subgraph "Entities"
            MSG[Message<br/>├─ ID, Role, Content<br/>└─ Metadata]
            TOOL[Tool<br/>├─ Name, Description<br/>├─ InputSchema<br/>└─ Handler]
            RES[Resource<br/>├─ URI, Name, MimeType<br/>└─ Reader]
            PROMPT[Prompt<br/>├─ Name, Arguments<br/>└─ Generator]
        end

        subgraph "Value Objects"
            IDS[Identifiers<br/>SessionID, ConversationID<br/>MessageID, ToolID]
            CONTENT[Content Types<br/>TextContent, Role<br/>Model, MimeType]
            MCP_VO[MCP Types<br/>Method, Capability<br/>LogLevel, ErrorCode]
        end

        subgraph "Domain Events"
            SESS_EVT[Session Events<br/>Created, Initialized<br/>Closed]
            CONV_EVT[Conversation Events<br/>Created, MessageAdded<br/>Closed]
            TOOL_EVT[Tool Events<br/>Registered, Executed]
        end
    end

    SESSION --> CONV
    CONV --> MSG
    SESSION --> TOOL
    SESSION --> RES
    SESSION --> PROMPT

    SESSION --> IDS
    CONV --> IDS
    MSG --> CONTENT
    TOOL --> MCP_VO

    SESSION --> SESS_EVT
    CONV --> CONV_EVT
    TOOL --> TOOL_EVT

    style SESSION fill:#C8E6C9,stroke:#388E3C,stroke-width:2px
    style CONV fill:#C8E6C9,stroke:#388E3C,stroke-width:2px
    style MSG fill:#BBDEFB,stroke:#1976D2
    style TOOL fill:#BBDEFB,stroke:#1976D2
    style IDS fill:#FFF9C4,stroke:#F9A825
    style CONTENT fill:#FFF9C4,stroke:#F9A825
```

---

## CQRS Pattern Implementation

```mermaid
flowchart LR
    subgraph "Commands - Write Side"
        C1[InitializeSession]
        C2[SendMessage]
        C3[ExecuteTool]
        C4[RegisterTool]
        C5[CloseSession]
    end

    subgraph "Command Handlers"
        CH1[SessionHandler]
        CH2[ConversationHandler]
        CH3[ToolHandler]
    end

    subgraph "Domain"
        AGG[Aggregates]
        REPO[Repositories]
        EVT[Events]
    end

    subgraph "Query Handlers"
        QH1[SessionHandler]
        QH2[ToolHandler]
        QH3[ResourceHandler]
    end

    subgraph "Queries - Read Side"
        Q1[GetSession]
        Q2[ListTools]
        Q3[ListResources]
        Q4[GetPrompt]
        Q5[ListConversations]
    end

    C1 --> CH1
    C2 --> CH2
    C3 --> CH3
    C4 --> CH3
    C5 --> CH1

    CH1 --> AGG
    CH2 --> AGG
    CH3 --> AGG
    AGG --> REPO
    AGG --> EVT

    Q1 --> QH1
    Q2 --> QH2
    Q3 --> QH3
    Q4 --> QH3
    Q5 --> QH1

    QH1 --> REPO
    QH2 --> REPO
    QH3 --> REPO

    style C1 fill:#FFCDD2,stroke:#C62828
    style C2 fill:#FFCDD2,stroke:#C62828
    style C3 fill:#FFCDD2,stroke:#C62828
    style Q1 fill:#C8E6C9,stroke:#388E3C
    style Q2 fill:#C8E6C9,stroke:#388E3C
    style Q3 fill:#C8E6C9,stroke:#388E3C
```

---

## Built-in Tools Architecture

```mermaid
graph TB
    subgraph "Tool Registry"
        REG[Tool Registry<br/>Manages all tools]
    end

    subgraph "AI Tools"
        T1[claude_conversation<br/>AI-powered chat]
    end

    subgraph "File Tools"
        T2[read_file<br/>Read file contents]
        T3[write_file<br/>Write to files]
        T4[list_directory<br/>List directory]
        T5[search_files<br/>Search by pattern]
    end

    subgraph "System Tools"
        T6[execute_command<br/>Run shell commands]
        T7[system_info<br/>System information]
    end

    subgraph "Utility Tools"
        T8[echo<br/>Testing utility]
    end

    REG --> T1
    REG --> T2
    REG --> T3
    REG --> T4
    REG --> T5
    REG --> T6
    REG --> T7
    REG --> T8

    subgraph "Execution Flow"
        INPUT[Tool Input]
        VALIDATE[Validate Schema]
        EXEC[Execute Handler]
        RESULT[Tool Result]
    end

    INPUT --> VALIDATE
    VALIDATE --> EXEC
    EXEC --> RESULT

    style T1 fill:#E1BEE7,stroke:#7B1FA2,stroke-width:2px
    style REG fill:#FFE0B2,stroke:#F57C00
```

### Tool Reference

| Tool                  | Category | Description                | Key Parameters                      |
| --------------------- | -------- | -------------------------- | ----------------------------------- |
| `claude_conversation` | AI       | Send messages to Claude AI | `message`, `model`, `system_prompt` |
| `read_file`           | File     | Read file contents         | `path`, `encoding`                  |
| `write_file`          | File     | Write content to file      | `path`, `content`, `create_dirs`    |
| `list_directory`      | File     | List directory contents    | `path`, `recursive`                 |
| `search_files`        | File     | Search files by pattern    | `path`, `pattern`                   |
| `execute_command`     | System   | Execute shell commands     | `command`, `working_dir`, `timeout` |
| `system_info`         | System   | Get system information     | -                                   |
| `echo`                | Utility  | Echo input (testing)       | `message`                           |

---

## Claude AI Integration

```mermaid
sequenceDiagram
    participant Tool as claude_conversation Tool
    participant Service as Claude Service
    participant SDK as anthropic-sdk-go
    participant API as Anthropic API

    Tool->>Service: CreateMessage Request
    Note over Service: Build ClaudeRequest<br/>Model, Messages, Tools

    Service->>SDK: client.Messages.New()
    SDK->>API: POST /v1/messages

    alt Success
        API-->>SDK: Message Response
        SDK-->>Service: *anthropic.Message
        Service->>Service: Convert to Domain Response
        Service-->>Tool: ClaudeResponse
    else Rate Limited
        API-->>SDK: 429 Error
        SDK-->>Service: Error
        Service->>Service: Retry with backoff
        Service->>SDK: Retry request
    else API Error
        API-->>SDK: Error Response
        SDK-->>Service: Error
        Service-->>Tool: Error Result
    end

    Note over Tool: Return ToolResult
```

### Supported Models

| Model             | ID                           | Use Case                       |
| ----------------- | ---------------------------- | ------------------------------ |
| Claude 4 Opus     | `claude-opus-4-20250514`     | Complex reasoning, analysis    |
| Claude 4 Sonnet   | `claude-sonnet-4-20250514`   | Balanced performance (default) |
| Claude 3.7 Sonnet | `claude-3-7-sonnet-20250219` | Extended thinking              |
| Claude 3.5 Sonnet | `claude-3-5-sonnet-20241022` | Fast, capable                  |
| Claude 3.5 Haiku  | `claude-3-5-haiku-20241022`  | Quick responses                |

---

## Session Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Created: New Session
    Created --> Initializing: Initialize Request
    Initializing --> Ready: Initialized Notification
    Ready --> Ready: Tool/Resource/Prompt Operations
    Ready --> Closing: Shutdown Request
    Closing --> Closed: Cleanup Complete
    Closed --> [*]

    note right of Created
        Session aggregate created
        Default capabilities set
    end note

    note right of Ready
        Full MCP operations available
        Tools, Resources, Prompts
    end note

    note right of Closed
        All conversations closed
        Resources released
    end note
```

---

## Configuration Architecture

```mermaid
graph TD
    subgraph "Configuration Sources"
        ENV[Environment Variables<br/>ANTHROPIC_API_KEY<br/>TELEMETRYFLOW_MCP_*]
        FILE[Config File<br/>config.yaml]
        DEFAULT[Default Values]
    end

    subgraph "Viper Configuration"
        VIPER[Viper Manager]
    end

    subgraph "Configuration Sections"
        SRV[Server Config<br/>name, port, transport]
        CLAUDE[Claude Config<br/>api_key, model, tokens]
        MCP_CFG[MCP Config<br/>capabilities, limits]
        LOG[Logging Config<br/>level, format]
        TEL[Telemetry Config<br/>OTEL settings]
        SEC[Security Config<br/>rate limits, CORS]
    end

    ENV --> VIPER
    FILE --> VIPER
    DEFAULT --> VIPER

    VIPER --> SRV
    VIPER --> CLAUDE
    VIPER --> MCP_CFG
    VIPER --> LOG
    VIPER --> TEL
    VIPER --> SEC

    style VIPER fill:#FFE0B2,stroke:#F57C00,stroke-width:2px
    style ENV fill:#C8E6C9,stroke:#388E3C
```

### Environment Variables

| Variable                                 | Description               | Default                    |
| ---------------------------------------- | ------------------------- | -------------------------- |
| `ANTHROPIC_API_KEY`                      | Claude API key (required) | -                          |
| `TELEMETRYFLOW_MCP_SERVER_TRANSPORT`     | Transport type            | `stdio`                    |
| `TELEMETRYFLOW_MCP_SERVER_PORT`          | Server port (SSE/WS)      | `8080`                     |
| `TELEMETRYFLOW_MCP_LOG_LEVEL`            | Log level                 | `info`                     |
| `TELEMETRYFLOW_MCP_LOG_FORMAT`           | Log format                | `json`                     |
| `TELEMETRYFLOW_MCP_DEBUG`                | Debug mode                | `false`                    |
| `TELEMETRYFLOW_MCP_CLAUDE_DEFAULT_MODEL` | Default Claude model      | `claude-sonnet-4-20250514` |
| `TELEMETRYFLOW_MCP_OTLP_ENDPOINT`        | OTEL collector endpoint   | `localhost:4317`           |

---

## Installation

### Prerequisites

- Go 1.24 or later
- Anthropic API key

### From Source

```bash
# Clone the repository
git clone https://github.com/telemetryflow/telemetryflow-go-mcp.git
cd telemetryflow/telemetryflow-go-mcp

# Download dependencies
make deps

# Build
make build

# Install to GOPATH/bin
make install
```

### Using Go Install

```bash
go install github.com/telemetryflow/telemetryflow-go-mcp/cmd/mcp@latest
```

### Docker

```bash
# Build image
docker build -t telemetryflow-go-mcp:1.1.2 .

# Run container
docker run --rm -it \
  -e ANTHROPIC_API_KEY="your-api-key" \
  telemetryflow-go-mcp:1.1.2
```

---

## Configuration

### Configuration File

Create `tfo-mcp.yaml` or use `configs/tfo-mcp.yaml`:

```yaml
# =============================================================================
# TelemetryFlow GO MCP Server Configuration
# Version: 1.1.2
# =============================================================================

server:
  name: "TelemetryFlow-MCP"
  version: "1.1.2"
  transport: "stdio" # stdio, sse, websocket
  debug: false

claude:
  # api_key: Set via ANTHROPIC_API_KEY env var
  default_model: "claude-sonnet-4-20250514"
  max_tokens: 4096
  temperature: 1.0
  timeout: "120s"
  max_retries: 3

mcp:
  protocol_version: "2024-11-05"
  enable_tools: true
  enable_resources: true
  enable_prompts: true
  enable_logging: true
  tool_timeout: "30s"

logging:
  level: "info" # debug, info, warn, error
  format: "json" # json, text
  output: "stderr"

telemetry:
  enabled: true
  service_name: "telemetryflow-go-mcp"
  otlp_endpoint: "localhost:4317"
  trace_sample_rate: 1.0
```

---

## Usage

### Running the Server

```bash
# Run with default config
tfo-mcp

# Run with custom config
tfo-mcp --config /path/to/config.yaml

# Run in debug mode
tfo-mcp --debug

# Show version
tfo-mcp version

# Validate configuration
tfo-mcp validate
```

### Integration with Claude Code

Add to your Claude Code MCP settings (`~/.config/claude-code/mcp_settings.json`):

```json
{
  "mcpServers": {
    "telemetryflow": {
      "command": "tfo-mcp",
      "args": [],
      "env": {
        "ANTHROPIC_API_KEY": "your-api-key"
      }
    }
  }
}
```

### MCP Protocol Examples

#### Initialize Session

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "my-client",
      "version": "1.0.0"
    }
  }
}
```

#### Response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "serverInfo": {
      "name": "TelemetryFlow-MCP",
      "version": "1.1.2"
    },
    "capabilities": {
      "tools": { "listChanged": true },
      "resources": { "subscribe": true, "listChanged": true },
      "prompts": { "listChanged": true },
      "logging": {}
    }
  }
}
```

#### Call Claude Conversation Tool

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "claude_conversation",
    "arguments": {
      "message": "Explain the MCP protocol in simple terms",
      "model": "claude-sonnet-4-20250514",
      "max_tokens": 1024
    }
  }
}
```

---

## Project Structure

```
telemetryflow-go-mcp/
├── cmd/
│   └── mcp/
│       └── main.go                     # Application entry point
├── internal/
│   ├── domain/                         # Domain Layer (DDD)
│   │   ├── aggregates/                 # Session, Conversation aggregates
│   │   │   ├── session.go
│   │   │   └── conversation.go
│   │   ├── entities/                   # Message, Tool, Resource, Prompt
│   │   │   ├── message.go
│   │   │   ├── tool.go
│   │   │   ├── resource.go
│   │   │   └── prompt.go
│   │   ├── valueobjects/               # Immutable value objects
│   │   │   ├── identifiers.go
│   │   │   ├── content.go
│   │   │   └── mcp.go
│   │   ├── events/                     # Domain events
│   │   │   └── events.go
│   │   ├── repositories/               # Repository interfaces
│   │   │   └── repositories.go
│   │   └── services/                   # Domain service interfaces
│   │       └── claude_service.go
│   ├── application/                    # Application Layer (CQRS)
│   │   ├── commands/                   # Write operations
│   │   │   └── commands.go
│   │   ├── queries/                    # Read operations
│   │   │   └── queries.go
│   │   └── handlers/                   # Command/Query handlers
│   │       ├── session_handler.go
│   │       ├── tool_handler.go
│   │       └── conversation_handler.go
│   ├── infrastructure/                 # Infrastructure Layer
│   │   ├── claude/                     # Claude API client
│   │   │   └── client.go
│   │   ├── config/                     # Configuration management
│   │   │   └── config.go
│   │   ├── cache/                      # Redis cache implementation
│   │   │   └── redis.go
│   │   ├── queue/                      # NATS JetStream queue
│   │   │   ├── nats.go
│   │   │   └── tasks.go
│   │   └── persistence/                # Repository implementations
│   │       ├── memory_repositories.go
│   │       ├── clickhouse.go           # ClickHouse analytics
│   │       ├── analytics_repository.go # Analytics queries
│   │       ├── migrator.go             # Database migrations
│   │       ├── seeder.go               # Database seeding
│   │       └── models/                 # GORM models
│   │           └── models.go
│   └── presentation/                   # Presentation Layer
│       ├── server/                     # MCP server implementation
│       │   └── server.go
│       └── tools/                      # Built-in tools
│           └── builtin_tools.go
├── migrations/                         # Database migrations
│   ├── postgres/                       # PostgreSQL migrations
│   │   ├── 000001_init_schema.up.sql
│   │   └── 000001_init_schema.down.sql
│   └── clickhouse/                     # ClickHouse migrations
│       ├── 000001_init_analytics.up.sql
│       └── 000001_init_analytics.down.sql
├── scripts/                            # Initialization scripts
│   ├── init-db.sql                     # PostgreSQL Docker init
│   └── init-clickhouse.sql             # ClickHouse Docker init
├── tests/                              # Test suites
│   ├── unit/                           # Unit tests
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── presentation/
│   └── integration/                    # Integration tests
├── configs/
│   └── config.yaml                     # Default configuration
├── docs/                               # Documentation
│   ├── README.md
│   ├── ARCHITECTURE.md
│   ├── CONFIGURATION.md
│   ├── COMMANDS.md
│   ├── ERD.md                          # Entity relationship diagrams
│   └── DEVELOPMENT.md
├── .kiro/                              # Specifications and steering
│   └── steering/
│       ├── tech.md
│       └── development-patterns.md
├── Makefile                            # Build automation
├── Dockerfile                          # Container build
├── docker-compose.yml                  # Local development stack
├── go.mod                              # Go module
├── .env.example                        # Environment template
└── .gitignore
```

---

## Development

### Make Commands

```bash
# Development
make build              # Build binary
make build-release      # Build optimized release binary
make run                # Build and run
make run-debug          # Run in debug mode
make install            # Install to GOPATH/bin
make clean              # Clean build artifacts

# Dependencies
make deps               # Download dependencies
make deps-update        # Update and tidy dependencies
make deps-refresh       # Refresh all dependencies (clean + download)
make deps-vendor        # Vendor dependencies
make deps-check         # Check for vulnerabilities (requires govulncheck)
make deps-graph         # Show dependency graph
make deps-why DEP=...   # Explain why a dependency is needed

# Code Quality
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint
make lint-fix           # Auto-fix lint issues

# Testing
make test               # Run tests
make test-cover         # Tests with coverage
make test-bench         # Run benchmarks
make test-short         # Run short tests only
make test-all           # Run all tests (unit, integration, e2e)

# Cross-compilation
make build-all          # Build for all platforms
make build-linux        # Build for Linux
make build-darwin       # Build for macOS
make build-windows      # Build for Windows

# Docker
make docker-build       # Build Docker image
make docker-run         # Run Docker container

# CI/CD
make ci                 # Full CI pipeline
make ci-test            # CI pipeline (format, vet, lint, test)
make release            # Create release artifacts

# CI-Specific (GitHub Actions)
make test-unit-ci       # Unit tests with coverage output
make test-integration-ci # Integration tests with coverage
make test-e2e-ci        # End-to-end tests
make ci-build           # Cross-platform CI build
make deps-verify        # Verify dependencies
make staticcheck        # Run staticcheck
make govulncheck        # Vulnerability scanning
make coverage-report    # Generate merged coverage report
```

### Testing

```bash
# Run all tests
make test

# Run all test types (unit, integration, e2e)
make test-all

# Run tests with coverage
make test-cover

# View coverage report
open build/coverage.html

# Run benchmarks
make test-bench

# Run CI test pipeline (format + vet + lint + test)
make ci-test
```

---

## OpenTelemetry Integration

```mermaid
graph LR
    subgraph "TFO-GO-MCP"
        APP[Application]
        TRACER[Tracer Provider]
        METER[Meter Provider]
    end

    subgraph "Export"
        OTLP_EXP[OTLP Exporter]
    end

    subgraph "TelemetryFlow Stack"
        COLLECTOR[TFO-Collector<br/>:4317 gRPC<br/>:4318 HTTP]
        BACKEND[TelemetryFlow Backend]
    end

    APP --> TRACER
    APP --> METER
    TRACER --> OTLP_EXP
    METER --> OTLP_EXP
    OTLP_EXP --> COLLECTOR
    COLLECTOR --> BACKEND

    style COLLECTOR fill:#FFE0B2,stroke:#F57C00
    style OTLP_EXP fill:#BBDEFB,stroke:#1976D2
```

### Telemetry Configuration

```yaml
telemetry:
  enabled: true
  service_name: "telemetryflow-go-mcp"
  environment: "production"
  otlp_endpoint: "localhost:4317"
  otlp_insecure: false
  trace_sample_rate: 1.0
  metrics_enabled: true
  metrics_interval: "30s"
```

---

## MCP Capabilities Matrix

| Capability              | Status | Description                   |
| ----------------------- | ------ | ----------------------------- |
| `tools`                 | ✅     | Tool listing and execution    |
| `tools.listChanged`     | ✅     | Dynamic tool registration     |
| `resources`             | ✅     | Resource listing and reading  |
| `resources.subscribe`   | ✅     | Resource change subscriptions |
| `resources.listChanged` | ✅     | Dynamic resource registration |
| `prompts`               | ✅     | Prompt templates              |
| `prompts.listChanged`   | ✅     | Dynamic prompt registration   |
| `logging`               | ✅     | Log level management          |
| `sampling`              | 🔜     | LLM sampling (planned)        |

---

## Error Handling

```mermaid
graph TD
    subgraph "JSON-RPC Errors"
        E1[Parse Error<br/>-32700]
        E2[Invalid Request<br/>-32600]
        E3[Method Not Found<br/>-32601]
        E4[Invalid Params<br/>-32602]
        E5[Internal Error<br/>-32603]
    end

    subgraph "MCP Errors"
        M1[Tool Not Found<br/>-32001]
        M2[Resource Not Found<br/>-32002]
        M3[Prompt Not Found<br/>-32003]
        M4[Tool Execution Error<br/>-32004]
        M5[Rate Limited<br/>-32007]
    end

    REQ[Request] --> PARSE{Parse JSON}
    PARSE -->|Error| E1
    PARSE -->|OK| VALIDATE{Validate}
    VALIDATE -->|Invalid| E2
    VALIDATE -->|OK| ROUTE{Route Method}
    ROUTE -->|Not Found| E3
    ROUTE -->|OK| EXEC{Execute}
    EXEC -->|Tool Error| M4
    EXEC -->|Not Found| M1
    EXEC -->|OK| RESP[Response]

    style E1 fill:#FFCDD2,stroke:#C62828
    style E2 fill:#FFCDD2,stroke:#C62828
    style M4 fill:#FFE0B2,stroke:#F57C00
```

---

## Security Considerations

| Aspect                | Implementation                           |
| --------------------- | ---------------------------------------- |
| **API Key Storage**   | Environment variables only               |
| **Command Execution** | Configurable timeout, sandboxing planned |
| **File Access**       | Path validation, no traversal            |
| **Rate Limiting**     | Configurable per-minute limits           |
| **CORS**              | Configurable for SSE transport           |
| **Input Validation**  | JSON Schema validation for tools         |

---

## Documentation Index

| Document                                           | Description                         |
| -------------------------------------------------- | ----------------------------------- |
| [README.md](README.md)                             | Project overview and quick start    |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)       | Detailed architecture documentation |
| [docs/CONFIGURATION.md](docs/CONFIGURATION.md)     | Configuration reference             |
| [docs/COMMANDS.md](docs/COMMANDS.md)               | CLI commands reference              |
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)         | Development guide                   |
| [docs/INSTALLATION.md](docs/INSTALLATION.md)       | Installation guide                  |
| [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) | Troubleshooting guide               |

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Use DDD patterns for domain logic
- Write unit tests for all handlers
- Document public APIs
- Keep commits atomic and well-described

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Support

- **Documentation**: [TelemetryFlow Docs](https://docs.telemetryflow.id)
- **Issues**: [GitHub Issues](https://github.com/telemetryflow/telemetryflow-go-mcp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/telemetryflow/telemetryflow-go-mcp/discussions)

---

<p align="center">
  <strong>Built with Go and Claude AI integration for the TelemetryFlow Platform</strong>
  <br/>
  <sub>Copyright © 2024-2026 TelemetryFlow. All rights reserved.</sub>
</p>
