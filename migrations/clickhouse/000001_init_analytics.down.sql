-- ============================================================================
-- TelemetryFlow GO MCP - ClickHouse Analytics Schema Migration (Rollback)
-- Version: 000001
-- Description: Drops analytics tables for TelemetryFlow GO MCP
-- ============================================================================

-- Drop materialized views first
DROP VIEW IF EXISTS telemetryflow_mcp.mv_tool_usage_hourly;
DROP VIEW IF EXISTS telemetryflow_mcp.mv_token_usage_hourly;

-- Drop tables
DROP TABLE IF EXISTS telemetryflow_mcp.latency_percentiles_hourly;
DROP TABLE IF EXISTS telemetryflow_mcp.error_analytics;
DROP TABLE IF EXISTS telemetryflow_mcp.tool_usage_hourly;
DROP TABLE IF EXISTS telemetryflow_mcp.token_usage_hourly;
DROP TABLE IF EXISTS telemetryflow_mcp.session_analytics;
DROP TABLE IF EXISTS telemetryflow_mcp.api_request_analytics;
DROP TABLE IF EXISTS telemetryflow_mcp.tool_call_analytics;
DROP TABLE IF EXISTS telemetryflow_mcp.schema_migrations;

-- Optionally drop the database
-- DROP DATABASE IF EXISTS telemetryflow_mcp;
