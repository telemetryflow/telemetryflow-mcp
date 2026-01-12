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

# Security Policy

## Table of Contents

- [Supported Versions](#supported-versions)
- [Reporting a Vulnerability](#reporting-a-vulnerability)
- [Security Measures](#security-measures)
- [Security Best Practices](#security-best-practices)
- [Dependency Security](#dependency-security)
- [API Key Security](#api-key-security)
- [Security Checklist](#security-checklist)

---

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.1.x   | :white_check_mark: |
| 1.0.x   | :x:                |
| < 1.0   | :x:                |

We recommend always using the latest version for the best security.

---

## Reporting a Vulnerability

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **Email**: security@devopscorner.id
2. **GitHub Security Advisories**: [Create a security advisory](https://github.com/telemetryflow/telemetryflow-go-mcp/security/advisories/new)

### What to Include

Please include the following information in your report:

- Type of vulnerability (e.g., XSS, SQL injection, buffer overflow)
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### Response Timeline

| Stage                      | Timeline              |
| -------------------------- | --------------------- |
| Initial Response           | Within 48 hours       |
| Vulnerability Confirmation | Within 7 days         |
| Fix Development            | Within 30 days        |
| Public Disclosure          | After fix is released |

### What to Expect

1. **Acknowledgment**: We will acknowledge receipt of your report within 48 hours.
2. **Assessment**: We will investigate and validate the reported vulnerability.
3. **Updates**: We will keep you informed of our progress.
4. **Fix**: We will develop and test a fix.
5. **Disclosure**: We will coordinate disclosure with you.
6. **Credit**: We will credit you in the security advisory (if desired).

---

## Security Measures

### Authentication & Authorization

```mermaid
flowchart TB
    subgraph Security["Security Layers"]
        API["API Key Validation"]
        RATE["Rate Limiting"]
        INPUT["Input Validation"]
        SANITIZE["Output Sanitization"]
    end

    REQUEST["Incoming Request"] --> API
    API --> RATE
    RATE --> INPUT
    INPUT --> SANITIZE
    SANITIZE --> HANDLER["Request Handler"]

    style Security fill:#ffcdd2,stroke:#f44336
```

### API Key Management

- API keys are never logged
- Keys are stored in environment variables
- Keys are not included in configuration files committed to version control
- Keys are validated on every request

### Input Validation

- All inputs are validated before processing
- JSON schema validation for tool arguments
- Path traversal prevention for file operations
- Command injection prevention for shell operations

### Rate Limiting

- Configurable request limits per minute
- Burst allowance for temporary spikes
- Per-client rate tracking

### Secure Defaults

- TLS for external API calls
- Minimal file permissions
- Secure default configuration

---

## Security Best Practices

### For Users

#### API Key Security

```bash
# DO: Use environment variables
export TELEMETRYFLOW_MCP_CLAUDE_API_KEY="your-api-key"

# DO: Use secret managers in production
TELEMETRYFLOW_MCP_CLAUDE_API_KEY=$(aws secretsmanager get-secret-value --secret-id telemetryflow-go-mcp/api-key)

# DON'T: Put keys in config files
# DON'T: Commit keys to version control
# DON'T: Share keys in logs or error messages
```

#### Configuration Security

```yaml
# config.yaml - DO NOT include secrets here
claude:
  api_key: "" # Use environment variable instead

security:
  rate_limit:
    enabled: true
    requests_per_minute: 60
```

#### File Operations

- Use absolute paths
- Validate file paths
- Restrict operations to allowed directories
- Set appropriate file permissions

#### Command Execution

- Validate all command inputs
- Use allowlists for permitted commands
- Set execution timeouts
- Run with minimal privileges

### For Developers

#### Code Security

```go
// DO: Validate inputs
func (h *Handler) Execute(ctx context.Context, input string) error {
    if err := validate(input); err != nil {
        return fmt.Errorf("invalid input: %w", err)
    }
    // ...
}

// DO: Use parameterized queries (for database operations)
// DO: Sanitize outputs
// DO: Handle errors without leaking information
```

#### Dependency Management

```bash
# Regularly update dependencies
go get -u ./...
go mod tidy

# Check for vulnerabilities
go list -m all | nancy sleuth
govulncheck ./...
```

#### Secrets in Code

```go
// DON'T: Hardcode secrets
const apiKey = "sk-ant-api..."  // NEVER DO THIS

// DO: Use environment variables
apiKey := os.Getenv("TELEMETRYFLOW_MCP_CLAUDE_API_KEY")
```

---

## Dependency Security

### Vulnerability Scanning

We use multiple tools to scan for vulnerabilities:

| Tool          | Purpose                   |
| ------------- | ------------------------- |
| `govulncheck` | Go vulnerability database |
| `nancy`       | Sonatype OSS Index        |
| `trivy`       | Container image scanning  |
| `grype`       | Container/SBOM scanning   |

### Running Security Scans

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for vulnerabilities
govulncheck ./...

# Using nancy
go list -json -m all | nancy sleuth

# Docker image scanning
trivy image telemetryflow-go-mcp:latest
```

### Dependency Update Policy

- Critical vulnerabilities: Patch within 24 hours
- High vulnerabilities: Patch within 7 days
- Medium vulnerabilities: Patch within 30 days
- Low vulnerabilities: Patch in next release

---

## API Key Security

### Storage Guidelines

| Environment | Recommendation                           |
| ----------- | ---------------------------------------- |
| Development | `.env` file (gitignored)                 |
| CI/CD       | Secret management (GitHub Secrets, etc.) |
| Production  | Secret manager (AWS, GCP, Azure, Vault)  |

### Rotation Policy

- Rotate keys every 90 days
- Rotate immediately if compromised
- Use separate keys for dev/staging/production

### Access Control

- Limit key access to necessary personnel
- Use separate keys per application/environment
- Monitor key usage for anomalies

---

## Security Checklist

### Before Deployment

- [ ] API keys stored securely (not in code/config)
- [ ] Rate limiting enabled
- [ ] Input validation implemented
- [ ] Logging does not include sensitive data
- [ ] TLS configured for external connections
- [ ] File permissions set correctly
- [ ] Dependencies scanned for vulnerabilities
- [ ] Security tests passing

### Ongoing Security

- [ ] Regular dependency updates
- [ ] Regular security scans
- [ ] Monitor for security advisories
- [ ] Review access controls
- [ ] Audit logging enabled
- [ ] Incident response plan ready

### For Contributors

- [ ] No secrets in code
- [ ] Input validation for new features
- [ ] Security tests for new functionality
- [ ] Documentation updated for security features
- [ ] Security review for significant changes

---

## Security Features

### Built-in Security

| Feature           | Description                            |
| ----------------- | -------------------------------------- |
| Rate Limiting     | Prevent abuse with configurable limits |
| Input Validation  | JSON schema validation for all inputs  |
| Path Sanitization | Prevent directory traversal attacks    |
| Command Allowlist | Restrict executable commands           |
| Error Handling    | Secure error messages without leaks    |
| Logging           | Sensitive data redaction               |

### Configuration

```yaml
security:
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10

  api_key_validation: true

  allowed_commands:
    - "go"
    - "make"
    - "git"
```

---

## Contact

For security concerns, please contact:

- **Security Email**: security@devopscorner.id
- **GitHub Security**: [Security Advisories](https://github.com/telemetryflow/telemetryflow-go-mcp/security/advisories)

---

Thank you for helping keep TelemetryFlow GO MCP and its users secure!
