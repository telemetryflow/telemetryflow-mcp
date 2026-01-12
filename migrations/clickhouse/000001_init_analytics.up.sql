-- ============================================================================
-- TelemetryFlow GO MCP - ClickHouse Analytics Schema Migration
-- Version: 000001
-- Description: Creates analytics tables for TelemetryFlow GO MCP
-- ============================================================================

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS telemetryflow_mcp;

-- ============================================================================
-- Tool Call Analytics Table
-- Stores detailed tool execution metrics
-- ============================================================================
CREATE TABLE IF NOT EXISTS telemetryflow_mcp.tool_call_analytics (
    timestamp DateTime64(3) CODEC(Delta, ZSTD(1)),
    session_id UUID,
    conversation_id UUID,
    tool_name LowCardinality(String),
    duration_ms UInt64,
    is_error UInt8,
    error_message String DEFAULT '',
    input_size UInt32,
    output_size UInt32,
    metadata String DEFAULT '{}'
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, tool_name, session_id)
TTL timestamp + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- ============================================================================
-- API Request Analytics Table
-- Stores Claude API request metrics
-- ============================================================================
CREATE TABLE IF NOT EXISTS telemetryflow_mcp.api_request_analytics (
    timestamp DateTime64(3) CODEC(Delta, ZSTD(1)),
    request_id UUID,
    session_id UUID,
    conversation_id UUID,
    model LowCardinality(String),
    input_tokens UInt32,
    output_tokens UInt32,
    total_tokens UInt32,
    duration_ms UInt64,
    status_code UInt16,
    is_error UInt8,
    is_streaming UInt8 DEFAULT 0,
    stop_reason LowCardinality(String) DEFAULT '',
    metadata String DEFAULT '{}'
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, model, session_id)
TTL timestamp + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Session Analytics Table
-- Stores session lifecycle events
-- ============================================================================
CREATE TABLE IF NOT EXISTS telemetryflow_mcp.session_analytics (
    timestamp DateTime64(3) CODEC(Delta, ZSTD(1)),
    session_id UUID,
    event_type LowCardinality(String),
    client_name String,
    client_version String,
    protocol_version String DEFAULT '2024-11-05',
    duration_ms UInt64 DEFAULT 0,
    message_count UInt32 DEFAULT 0,
    tool_call_count UInt32 DEFAULT 0,
    total_tokens UInt64 DEFAULT 0,
    error_count UInt32 DEFAULT 0,
    metadata String DEFAULT '{}'
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, event_type, session_id)
TTL timestamp + INTERVAL 180 DAY
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Token Usage Hourly Aggregates (Materialized View Target)
-- ============================================================================
CREATE TABLE IF NOT EXISTS telemetryflow_mcp.token_usage_hourly (
    hour DateTime CODEC(Delta, ZSTD(1)),
    model LowCardinality(String),
    input_tokens UInt64,
    output_tokens UInt64,
    total_tokens UInt64,
    request_count UInt64
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (hour, model)
SETTINGS index_granularity = 8192;

-- Materialized View for Token Usage Aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS telemetryflow_mcp.mv_token_usage_hourly
TO telemetryflow_mcp.token_usage_hourly
AS SELECT
    toStartOfHour(timestamp) AS hour,
    model,
    sum(input_tokens) AS input_tokens,
    sum(output_tokens) AS output_tokens,
    sum(total_tokens) AS total_tokens,
    count() AS request_count
FROM telemetryflow_mcp.api_request_analytics
GROUP BY hour, model;

-- ============================================================================
-- Tool Usage Hourly Aggregates (Materialized View Target)
-- ============================================================================
CREATE TABLE IF NOT EXISTS telemetryflow_mcp.tool_usage_hourly (
    hour DateTime CODEC(Delta, ZSTD(1)),
    tool_name LowCardinality(String),
    call_count UInt64,
    error_count UInt64,
    total_duration_ms UInt64,
    min_duration_ms UInt64,
    max_duration_ms UInt64
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (hour, tool_name)
SETTINGS index_granularity = 8192;

-- Materialized View for Tool Usage Aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS telemetryflow_mcp.mv_tool_usage_hourly
TO telemetryflow_mcp.tool_usage_hourly
AS SELECT
    toStartOfHour(timestamp) AS hour,
    tool_name,
    count() AS call_count,
    countIf(is_error = 1) AS error_count,
    sum(duration_ms) AS total_duration_ms,
    min(duration_ms) AS min_duration_ms,
    max(duration_ms) AS max_duration_ms
FROM telemetryflow_mcp.tool_call_analytics
GROUP BY hour, tool_name;

-- ============================================================================
-- Error Analytics Table
-- Stores error events for debugging
-- ============================================================================
CREATE TABLE IF NOT EXISTS telemetryflow_mcp.error_analytics (
    timestamp DateTime64(3) CODEC(Delta, ZSTD(1)),
    session_id UUID,
    conversation_id UUID,
    error_type LowCardinality(String),
    error_code String,
    error_message String,
    stack_trace String DEFAULT '',
    context String DEFAULT '{}',
    metadata String DEFAULT '{}'
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, error_type, session_id)
TTL timestamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Request Latency Percentiles Table (for dashboards)
-- ============================================================================
CREATE TABLE IF NOT EXISTS telemetryflow_mcp.latency_percentiles_hourly (
    hour DateTime CODEC(Delta, ZSTD(1)),
    model LowCardinality(String),
    p50_ms Float64,
    p90_ms Float64,
    p95_ms Float64,
    p99_ms Float64,
    avg_ms Float64,
    request_count UInt64
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (hour, model)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Schema Migrations Tracking
-- ============================================================================
CREATE TABLE IF NOT EXISTS telemetryflow_mcp.schema_migrations (
    version String,
    applied_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY version;

INSERT INTO telemetryflow_mcp.schema_migrations (version) VALUES ('000001_init_analytics');
