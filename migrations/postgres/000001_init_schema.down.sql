-- ============================================================================
-- TelemetryFlow GO MCP - PostgreSQL Initial Schema Migration (Rollback)
-- Version: 000001
-- Description: Drops the initial database schema for TelemetryFlow GO MCP
-- ============================================================================

-- Drop triggers
DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS update_prompts_updated_at ON prompts;
DROP TRIGGER IF EXISTS update_resources_updated_at ON resources;
DROP TRIGGER IF EXISTS update_tools_updated_at ON tools;
DROP TRIGGER IF EXISTS update_conversations_updated_at ON conversations;
DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS tool_executions;
DROP TABLE IF EXISTS resource_subscriptions;
DROP TABLE IF EXISTS prompts;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS tools;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS schema_migrations;

-- Drop extensions (optional - may affect other databases)
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
