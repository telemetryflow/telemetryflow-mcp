-- ============================================================================
-- TelemetryFlow GO MCP - ClickHouse Initialization Script
-- This script runs when the ClickHouse container starts for the first time
-- ============================================================================

-- Create database
CREATE DATABASE IF NOT EXISTS telemetryflow_mcp;

-- ============================================================================
-- Analytics Tables
-- ============================================================================

-- Tool Call Analytics
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

-- API Request Analytics
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

-- Session Analytics
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
-- Aggregation Tables
-- ============================================================================

-- Token Usage Hourly
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

-- Tool Usage Hourly
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

-- Error Analytics
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

-- Latency Percentiles
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
-- Materialized Views
-- ============================================================================

-- Token Usage Materialized View
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

-- Tool Usage Materialized View
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
-- Schema Migrations Tracking
-- ============================================================================

CREATE TABLE IF NOT EXISTS telemetryflow_mcp.schema_migrations (
    version String,
    applied_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY version;

INSERT INTO telemetryflow_mcp.schema_migrations (version) VALUES ('init');

-- ============================================================================
-- Seed Sample Analytics Data (for development/testing)
-- ============================================================================

-- Insert sample tool call analytics
INSERT INTO telemetryflow_mcp.tool_call_analytics (timestamp, session_id, conversation_id, tool_name, duration_ms, is_error, input_size, output_size)
SELECT
    now() - INTERVAL number MINUTE AS timestamp,
    generateUUIDv4() AS session_id,
    generateUUIDv4() AS conversation_id,
    arrayElement(['echo', 'read_file', 'write_file', 'list_directory', 'execute_command'], (number % 5) + 1) AS tool_name,
    50 + rand() % 500 AS duration_ms,
    if(rand() % 20 = 0, 1, 0) AS is_error,
    100 + rand() % 1000 AS input_size,
    200 + rand() % 5000 AS output_size
FROM numbers(100);

-- Insert sample API request analytics
INSERT INTO telemetryflow_mcp.api_request_analytics (timestamp, request_id, session_id, conversation_id, model, input_tokens, output_tokens, total_tokens, duration_ms, status_code, is_error)
SELECT
    now() - INTERVAL number MINUTE AS timestamp,
    generateUUIDv4() AS request_id,
    generateUUIDv4() AS session_id,
    generateUUIDv4() AS conversation_id,
    arrayElement(['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku', 'claude-3-5-sonnet'], (number % 4) + 1) AS model,
    100 + rand() % 2000 AS input_tokens,
    200 + rand() % 4000 AS output_tokens,
    300 + rand() % 6000 AS total_tokens,
    500 + rand() % 5000 AS duration_ms,
    if(rand() % 50 = 0, 500, 200) AS status_code,
    if(rand() % 50 = 0, 1, 0) AS is_error
FROM numbers(100);

-- Insert sample session analytics
INSERT INTO telemetryflow_mcp.session_analytics (timestamp, session_id, event_type, client_name, client_version, duration_ms, message_count, tool_call_count, total_tokens)
SELECT
    now() - INTERVAL number HOUR AS timestamp,
    generateUUIDv4() AS session_id,
    arrayElement(['session_created', 'session_active', 'session_closed'], (number % 3) + 1) AS event_type,
    arrayElement(['Claude Desktop', 'VS Code Extension', 'CLI Client'], (number % 3) + 1) AS client_name,
    '1.0.0' AS client_version,
    60000 + rand() % 3600000 AS duration_ms,
    5 + rand() % 50 AS message_count,
    2 + rand() % 20 AS tool_call_count,
    1000 + rand() % 100000 AS total_tokens
FROM numbers(50);
