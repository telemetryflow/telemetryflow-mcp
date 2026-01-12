# TelemetryFlow GO MCP Documentation Knowledge Base

## Overview

This steering file provides comprehensive references to the TelemetryFlow GO MCP Server documentation located in the `docs/` directory. Use these references to understand the complete system architecture, development processes, and operational procedures.

## Core Documentation References

### System Architecture

#### #[[file:docs/ARCHITECTURE.md]]

Complete system architecture documentation including:

- Domain-Driven Design implementation with LEGO Builder methodology
- CQRS patterns and clean architecture layers
- Component diagrams and system boundaries
- Database schemas and entity relationships
- Integration patterns with Claude API and MCP protocol

#### #[[file:docs/DFD.md]]

Data Flow Diagrams showing:

- MCP protocol message flows
- Claude API integration patterns
- Tool execution workflows
- Session and conversation management flows
- Error handling and recovery processes

#### #[[file:docs/ERD.md]]

Entity Relationship Diagrams covering:

- Domain model relationships
- Database schema design
- Aggregate boundaries and relationships
- Value object compositions
- Event sourcing patterns

### Development Guidelines

#### #[[file:docs/DEVELOPMENT.md]]

Development standards and practices including:

- Go coding standards and conventions
- Domain-Driven Design implementation patterns
- CQRS command and query patterns
- Testing strategies and coverage requirements
- Code review and quality assurance processes

#### #[[file:docs/GIT-WORKFLOW.md]]

Git workflow and branching strategy:

- Git Flow-inspired branching model
- Commit message conventions (Conventional Commits)
- Pull request processes and code review
- Release management and versioning
- Hotfix procedures for critical issues

### Configuration and Deployment

#### #[[file:docs/CONFIGURATION.md]]

Complete configuration reference:

- Environment variable configuration
- YAML configuration file structure
- Claude API integration settings
- Database and persistence configuration
- Observability and monitoring setup
- Security and authentication configuration

#### #[[file:docs/INSTALLATION.md]]

Installation and deployment guides:

- Local development setup
- Docker containerization
- Kubernetes deployment manifests
- Production environment configuration
- Monitoring and observability setup

### Operations and Commands

#### #[[file:docs/COMMANDS.md]]

CLI commands and MCP protocol reference:

- Server management commands
- MCP protocol method implementations
- Tool management and execution
- Session and conversation operations
- Administrative and maintenance commands

#### #[[file:docs/TROUBLESHOOTING.md]]

Troubleshooting and operational guidance:

- Common issues and solutions
- Error code references
- Performance optimization tips
- Debugging procedures
- Log analysis and monitoring

### Project Overview

#### #[[file:docs/README.md]]

Project overview and quick start:

- Project description and goals
- Quick start guide
- Feature overview
- Architecture summary
- Contributing guidelines

## Implementation Guidelines

### When to Reference Documentation

1. **Architecture Decisions**: Always reference `docs/ARCHITECTURE.md` when making architectural decisions or understanding system boundaries.

2. **Development Standards**: Follow patterns and conventions outlined in `docs/DEVELOPMENT.md` for all code implementations.

3. **Configuration Changes**: Consult `docs/CONFIGURATION.md` for all configuration-related implementations and environment setup.

4. **Protocol Implementation**: Reference `docs/COMMANDS.md` for MCP protocol compliance and command implementations.

5. **Data Flow Understanding**: Use `docs/DFD.md` to understand how data flows through the system before implementing new features.

6. **Database Design**: Reference `docs/ERD.md` for understanding entity relationships and database schema design.

7. **Deployment Procedures**: Follow `docs/INSTALLATION.md` for all deployment and infrastructure setup.

8. **Issue Resolution**: Use `docs/TROUBLESHOOTING.md` as the first reference for debugging and issue resolution.

### Documentation Integration Patterns

#### Architecture Alignment

```go
// Always align implementations with documented architecture
// Reference: docs/ARCHITECTURE.md - Domain Layer Implementation
type Session struct {
    id           vo.SessionID
    state        SessionState
    tools        map[string]*entities.Tool
    resources    map[string]*entities.Resource
    prompts      map[string]*entities.Prompt
    // Implementation follows documented aggregate patterns
}
```

#### Configuration Consistency

```yaml
# Follow documented configuration patterns
# Reference: docs/CONFIGURATION.md - Server Configuration
server:
  host: "localhost"
  port: 8080
  protocol_version: "2024-11-05"
  # Configuration structure matches documentation
```

#### Command Implementation

```go
// Implement commands according to documented specifications
// Reference: docs/COMMANDS.md - MCP Protocol Commands
func (h *MCPHandler) HandleInitialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
    // Implementation follows documented MCP protocol patterns
    return h.initializeSession(ctx, req)
}
```

## Development Workflow Integration

### Pre-Development Checklist

1. Review relevant sections in `docs/ARCHITECTURE.md`
2. Check `docs/DEVELOPMENT.md` for coding standards
3. Understand data flows from `docs/DFD.md`
4. Review entity relationships in `docs/ERD.md`
5. Check configuration requirements in `docs/CONFIGURATION.md`

### Implementation Phase

1. Follow patterns documented in `docs/DEVELOPMENT.md`
2. Implement commands according to `docs/COMMANDS.md`
3. Ensure configuration compatibility with `docs/CONFIGURATION.md`
4. Maintain architectural consistency with `docs/ARCHITECTURE.md`

### Testing and Validation

1. Test against specifications in `docs/COMMANDS.md`
2. Validate configuration with `docs/CONFIGURATION.md`
3. Verify architectural compliance with `docs/ARCHITECTURE.md`
4. Use `docs/TROUBLESHOOTING.md` for debugging

### Deployment and Operations

1. Follow deployment procedures in `docs/INSTALLATION.md`
2. Configure monitoring per `docs/CONFIGURATION.md`
3. Use `docs/TROUBLESHOOTING.md` for operational issues
4. Follow maintenance procedures in `docs/COMMANDS.md`

## Quality Assurance

### Documentation Compliance

- All implementations MUST align with documented architecture patterns
- Configuration changes MUST be compatible with documented schemas
- Command implementations MUST follow documented MCP protocol specifications
- Error handling MUST follow documented troubleshooting procedures

### Consistency Checks

- Verify architectural decisions against `docs/ARCHITECTURE.md`
- Validate data flow implementations against `docs/DFD.md`
- Check entity relationships against `docs/ERD.md`
- Ensure configuration consistency with `docs/CONFIGURATION.md`

### Review Process

- Reference appropriate documentation during code reviews
- Validate implementations against documented patterns
- Ensure new features align with documented architecture
- Update documentation when making architectural changes

## Maintenance and Updates

### Documentation Synchronization

- Keep implementations synchronized with documentation updates
- Update code when documentation patterns change
- Maintain consistency between code and documented architecture
- Review documentation regularly for updates and changes

### Knowledge Transfer

- Use documentation as primary source for onboarding
- Reference specific documentation sections in code comments
- Maintain traceability between implementation and documentation
- Ensure team understanding of documented patterns and procedures

## Best Practices

### Documentation-Driven Development

1. **Read First**: Always read relevant documentation before implementing
2. **Reference Explicitly**: Include documentation references in code comments
3. **Validate Continuously**: Regularly check implementation against documentation
4. **Update Together**: Update both code and documentation when making changes

### Implementation Patterns

```go
// Example: Documentation-referenced implementation
// Reference: docs/ARCHITECTURE.md - Session Aggregate Pattern
// Reference: docs/DFD.md - Session Lifecycle Flow
func (s *Session) Initialize(clientInfo ClientInfo) error {
    // Implementation follows documented patterns
    if err := s.validateClientInfo(clientInfo); err != nil {
        return fmt.Errorf("client validation failed: %w", err)
    }

    s.state = SessionStateActive
    s.lastActivity = time.Now()

    // Publish event as documented in DFD
    s.publishEvent(events.NewSessionInitializedEvent(s.id))

    return nil
}
```

### Configuration Management

```yaml
# Reference: docs/CONFIGURATION.md - Complete Configuration Schema
# Ensure all configuration follows documented structure and validation rules
mcp:
  protocol_version: "2024-11-05" # As specified in docs/COMMANDS.md
  capabilities: # As defined in docs/ARCHITECTURE.md
    tools: true
    resources: true
    prompts: true
```

This knowledge base ensures that all development work is grounded in the comprehensive documentation available in the `docs/` directory, maintaining consistency and quality across the entire TelemetryFlow GO MCP Server implementation.
