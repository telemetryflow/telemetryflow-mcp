-- ============================================================================
-- TelemetryFlow GO MCP - PostgreSQL Initial Schema Migration
-- Version: 000001
-- Description: Creates the initial database schema for TelemetryFlow GO MCP
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Sessions Table
-- ============================================================================
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
    CONSTRAINT sessions_state_check CHECK (state IN ('created', 'initializing', 'ready', 'closed')),
    CONSTRAINT sessions_log_level_check CHECK (log_level IN ('debug', 'info', 'notice', 'warning', 'error', 'critical', 'alert', 'emergency'))
);

-- Session indexes
CREATE INDEX idx_sessions_state ON sessions(state);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);
CREATE INDEX idx_sessions_client_name ON sessions(client_name);

-- ============================================================================
-- Conversations Table
-- ============================================================================
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
    CONSTRAINT conversations_status_check CHECK (status IN ('active', 'paused', 'closed', 'archived')),
    CONSTRAINT conversations_temperature_check CHECK (temperature >= 0 AND temperature <= 2),
    CONSTRAINT conversations_top_p_check CHECK (top_p >= 0 AND top_p <= 1)
);

-- Conversation indexes
CREATE INDEX idx_conversations_session_id ON conversations(session_id);
CREATE INDEX idx_conversations_status ON conversations(status);
CREATE INDEX idx_conversations_model ON conversations(model);
CREATE INDEX idx_conversations_created_at ON conversations(created_at);

-- ============================================================================
-- Messages Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content JSONB NOT NULL DEFAULT '[]',
    token_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT messages_role_check CHECK (role IN ('user', 'assistant'))
);

-- Message indexes
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- ============================================================================
-- Tools Table
-- ============================================================================
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tool indexes
CREATE INDEX idx_tools_name ON tools(name);
CREATE INDEX idx_tools_category ON tools(category);
CREATE INDEX idx_tools_is_enabled ON tools(is_enabled);

-- ============================================================================
-- Resources Table
-- ============================================================================
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Resource indexes
CREATE INDEX idx_resources_uri ON resources(uri);
CREATE INDEX idx_resources_name ON resources(name);
CREATE INDEX idx_resources_mime_type ON resources(mime_type);

-- ============================================================================
-- Prompts Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS prompts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    arguments JSONB NOT NULL DEFAULT '[]',
    template TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Prompt indexes
CREATE INDEX idx_prompts_name ON prompts(name);

-- ============================================================================
-- Resource Subscriptions Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS resource_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    resource_uri VARCHAR(2048) NOT NULL,
    subscribed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, resource_uri)
);

-- Subscription indexes
CREATE INDEX idx_resource_subscriptions_session_id ON resource_subscriptions(session_id);
CREATE INDEX idx_resource_subscriptions_resource_uri ON resource_subscriptions(resource_uri);

-- ============================================================================
-- Tool Executions Table (for auditing)
-- ============================================================================
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
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tool execution indexes
CREATE INDEX idx_tool_executions_session_id ON tool_executions(session_id);
CREATE INDEX idx_tool_executions_tool_name ON tool_executions(tool_name);
CREATE INDEX idx_tool_executions_executed_at ON tool_executions(executed_at);
CREATE INDEX idx_tool_executions_is_error ON tool_executions(is_error);

-- ============================================================================
-- API Keys Table (for authentication)
-- ============================================================================
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- API key indexes
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);

-- ============================================================================
-- Updated At Trigger Function
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversations_updated_at
    BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tools_updated_at
    BEFORE UPDATE ON tools
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_resources_updated_at
    BEFORE UPDATE ON resources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_prompts_updated_at
    BEFORE UPDATE ON prompts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Schema Version Tracking
-- ============================================================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schema_migrations (version) VALUES ('000001_init_schema');
