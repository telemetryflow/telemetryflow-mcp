-- ============================================================================
-- TelemetryFlow GO MCP - PostgreSQL Initialization Script
-- This script runs when the PostgreSQL container starts for the first time
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Core Tables
-- ============================================================================

-- Sessions Table
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    protocol_version VARCHAR(20) NOT NULL DEFAULT '2024-11-05',
    state VARCHAR(20) NOT NULL DEFAULT 'created',
    client_name VARCHAR(255),
    client_version VARCHAR(50),
    server_name VARCHAR(255) NOT NULL DEFAULT 'TelemetryFlow-MCP',
    server_version VARCHAR(50) NOT NULL DEFAULT '1.1.2',
    capabilities JSONB NOT NULL DEFAULT '{}',
    log_level VARCHAR(20) NOT NULL DEFAULT 'info',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT sessions_state_check CHECK (state IN ('created', 'initializing', 'ready', 'closed'))
);

-- Conversations Table
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    model VARCHAR(100) NOT NULL DEFAULT 'claude-sonnet-4-20250514',
    system_prompt TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    max_tokens INTEGER NOT NULL DEFAULT 4096,
    temperature DECIMAL(3,2) NOT NULL DEFAULT 1.0,
    top_p DECIMAL(3,2) NOT NULL DEFAULT 1.0,
    top_k INTEGER NOT NULL DEFAULT 0,
    stop_sequences JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT conversations_status_check CHECK (status IN ('active', 'paused', 'closed', 'archived'))
);

-- Messages Table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content JSONB NOT NULL DEFAULT '[]',
    token_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT messages_role_check CHECK (role IN ('user', 'assistant'))
);

-- Tools Table
CREATE TABLE IF NOT EXISTS tools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    input_schema JSONB NOT NULL DEFAULT '{}',
    category VARCHAR(100),
    tags JSONB NOT NULL DEFAULT '[]',
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    rate_limit JSONB,
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Resources Table
CREATE TABLE IF NOT EXISTS resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uri VARCHAR(2048) NOT NULL UNIQUE,
    uri_template VARCHAR(2048),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    mime_type VARCHAR(255),
    is_template BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Prompts Table
CREATE TABLE IF NOT EXISTS prompts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    arguments JSONB NOT NULL DEFAULT '[]',
    template TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Resource Subscriptions Table
CREATE TABLE IF NOT EXISTS resource_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    resource_uri VARCHAR(2048) NOT NULL,
    subscribed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(session_id, resource_uri)
);

-- Tool Executions Table
CREATE TABLE IF NOT EXISTS tool_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    tool_name VARCHAR(255) NOT NULL,
    input JSONB NOT NULL DEFAULT '{}',
    output JSONB,
    is_error BOOLEAN NOT NULL DEFAULT false,
    error_message TEXT,
    duration_ms INTEGER,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- API Keys Table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    scopes JSONB NOT NULL DEFAULT '["read", "write"]',
    rate_limit_per_minute INTEGER DEFAULT 60,
    rate_limit_per_hour INTEGER DEFAULT 1000,
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Schema Migrations Table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- Indexes
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_sessions_state ON sessions(state);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at ON sessions(deleted_at);

CREATE INDEX IF NOT EXISTS idx_conversations_session_id ON conversations(session_id);
CREATE INDEX IF NOT EXISTS idx_conversations_status ON conversations(status);
CREATE INDEX IF NOT EXISTS idx_conversations_deleted_at ON conversations(deleted_at);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_role ON messages(role);
CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages(deleted_at);

CREATE INDEX IF NOT EXISTS idx_tools_name ON tools(name);
CREATE INDEX IF NOT EXISTS idx_tools_category ON tools(category);
CREATE INDEX IF NOT EXISTS idx_tools_is_enabled ON tools(is_enabled);
CREATE INDEX IF NOT EXISTS idx_tools_deleted_at ON tools(deleted_at);

CREATE INDEX IF NOT EXISTS idx_resources_uri ON resources(uri);
CREATE INDEX IF NOT EXISTS idx_resources_deleted_at ON resources(deleted_at);

CREATE INDEX IF NOT EXISTS idx_prompts_name ON prompts(name);
CREATE INDEX IF NOT EXISTS idx_prompts_deleted_at ON prompts(deleted_at);

CREATE INDEX IF NOT EXISTS idx_resource_subscriptions_session_id ON resource_subscriptions(session_id);
CREATE INDEX IF NOT EXISTS idx_resource_subscriptions_deleted_at ON resource_subscriptions(deleted_at);

CREATE INDEX IF NOT EXISTS idx_tool_executions_session_id ON tool_executions(session_id);
CREATE INDEX IF NOT EXISTS idx_tool_executions_tool_name ON tool_executions(tool_name);
CREATE INDEX IF NOT EXISTS idx_tool_executions_executed_at ON tool_executions(executed_at);
CREATE INDEX IF NOT EXISTS idx_tool_executions_deleted_at ON tool_executions(deleted_at);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_api_keys_deleted_at ON api_keys(deleted_at);

-- ============================================================================
-- Updated At Trigger
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_sessions_updated_at') THEN
        CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_conversations_updated_at') THEN
        CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_tools_updated_at') THEN
        CREATE TRIGGER update_tools_updated_at BEFORE UPDATE ON tools FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_resources_updated_at') THEN
        CREATE TRIGGER update_resources_updated_at BEFORE UPDATE ON resources FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_prompts_updated_at') THEN
        CREATE TRIGGER update_prompts_updated_at BEFORE UPDATE ON prompts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_api_keys_updated_at') THEN
        CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- Record migration
INSERT INTO schema_migrations (version) VALUES ('init') ON CONFLICT (version) DO NOTHING;

-- ============================================================================
-- Seed Default Data
-- ============================================================================

-- Default Tools
INSERT INTO tools (id, name, description, input_schema, category, tags, is_enabled, timeout_seconds)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'echo', 'Echoes back the input message', '{"type":"object","properties":{"message":{"type":"string","description":"The message to echo back"}},"required":["message"]}', 'utility', '["testing","debug"]', true, 10),
    ('00000000-0000-0000-0000-000000000002', 'read_file', 'Reads the contents of a file', '{"type":"object","properties":{"path":{"type":"string","description":"The path to the file"}},"required":["path"]}', 'filesystem', '["file","read"]', true, 30),
    ('00000000-0000-0000-0000-000000000003', 'write_file', 'Writes content to a file', '{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"}},"required":["path","content"]}', 'filesystem', '["file","write"]', true, 30),
    ('00000000-0000-0000-0000-000000000004', 'list_directory', 'Lists directory contents', '{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}', 'filesystem', '["directory","list"]', true, 30),
    ('00000000-0000-0000-0000-000000000005', 'execute_command', 'Executes a shell command', '{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}', 'system', '["shell","command"]', true, 60),
    ('00000000-0000-0000-0000-000000000006', 'search_files', 'Searches for files by pattern', '{"type":"object","properties":{"pattern":{"type":"string"}},"required":["pattern"]}', 'filesystem', '["search","glob"]', true, 60),
    ('00000000-0000-0000-0000-000000000007', 'system_info', 'Returns system information', '{"type":"object","properties":{}}', 'system', '["system","info"]', true, 10),
    ('00000000-0000-0000-0000-000000000008', 'claude_conversation', 'Initiates conversation with Claude', '{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}', 'ai', '["claude","ai"]', true, 120)
ON CONFLICT (name) DO NOTHING;

-- Default Resources
INSERT INTO resources (id, uri, name, description, mime_type, is_template)
VALUES
    ('00000000-0000-0000-0000-000000000101', 'config://server', 'Server Configuration', 'Current server configuration', 'application/json', false),
    ('00000000-0000-0000-0000-000000000102', 'status://health', 'Health Status', 'Server health information', 'application/json', false)
ON CONFLICT (uri) DO NOTHING;

-- Default Prompts
INSERT INTO prompts (id, name, description, arguments, template)
VALUES
    ('00000000-0000-0000-0000-000000000201', 'code_review', 'Reviews code for quality and bugs', '[{"name":"code","required":true},{"name":"language","required":false}]', 'Please review the following code...'),
    ('00000000-0000-0000-0000-000000000202', 'explain_code', 'Explains code in plain language', '[{"name":"code","required":true}]', 'Please explain the following code...'),
    ('00000000-0000-0000-0000-000000000203', 'debug_help', 'Helps debug errors', '[{"name":"error","required":true}]', 'Help debug this error...')
ON CONFLICT (name) DO NOTHING;

RAISE NOTICE 'TelemetryFlow GO MCP PostgreSQL initialization complete!';
