// Package config contains configuration management for the TelemetryFlow GO MCP service
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the MCP server
type Config struct {
	// Server configuration
	Server ServerConfig `mapstructure:"server"`

	// Claude API configuration
	Claude ClaudeConfig `mapstructure:"claude"`

	// MCP configuration
	MCP MCPConfig `mapstructure:"mcp"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging"`

	// Telemetry configuration
	Telemetry TelemetryConfig `mapstructure:"telemetry"`

	// Security configuration
	Security SecurityConfig `mapstructure:"security"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`

	// Transport type: "stdio", "sse", "websocket"
	Transport string `mapstructure:"transport"`

	// Timeouts
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`

	// Debug mode
	Debug bool `mapstructure:"debug"`
}

// ClaudeConfig holds Claude API configuration
type ClaudeConfig struct {
	APIKey         string        `mapstructure:"api_key"`
	BaseURL        string        `mapstructure:"base_url"`
	DefaultModel   string        `mapstructure:"default_model"`
	MaxTokens      int           `mapstructure:"max_tokens"`
	Temperature    float64       `mapstructure:"temperature"`
	TopP           float64       `mapstructure:"top_p"`
	TopK           int           `mapstructure:"top_k"`
	Timeout        time.Duration `mapstructure:"timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
	EnableBatching bool          `mapstructure:"enable_batching"`
}

// MCPConfig holds MCP protocol configuration
type MCPConfig struct {
	ProtocolVersion string `mapstructure:"protocol_version"`

	// Capabilities
	EnableTools     bool `mapstructure:"enable_tools"`
	EnableResources bool `mapstructure:"enable_resources"`
	EnablePrompts   bool `mapstructure:"enable_prompts"`
	EnableLogging   bool `mapstructure:"enable_logging"`
	EnableSampling  bool `mapstructure:"enable_sampling"`

	// Limits
	MaxToolsPerSession     int `mapstructure:"max_tools_per_session"`
	MaxResourcesPerSession int `mapstructure:"max_resources_per_session"`
	MaxPromptsPerSession   int `mapstructure:"max_prompts_per_session"`
	MaxConversations       int `mapstructure:"max_conversations"`
	MaxMessagesPerConv     int `mapstructure:"max_messages_per_conv"`

	// Tool execution
	ToolTimeout time.Duration `mapstructure:"tool_timeout"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // "json" or "text"
	Output     string `mapstructure:"output"` // "stdout", "stderr", or file path
	AddSource  bool   `mapstructure:"add_source"`
	TimeFormat string `mapstructure:"time_format"`
}

// TelemetryConfig holds OpenTelemetry configuration
type TelemetryConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	ServiceName  string `mapstructure:"service_name"`
	Environment  string `mapstructure:"environment"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
	OTLPInsecure bool   `mapstructure:"otlp_insecure"`

	// Trace sampling
	TraceSampleRate float64 `mapstructure:"trace_sample_rate"`

	// Metrics
	MetricsEnabled  bool          `mapstructure:"metrics_enabled"`
	MetricsInterval time.Duration `mapstructure:"metrics_interval"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	// API Key validation
	RequireAPIKey  bool     `mapstructure:"require_api_key"`
	AllowedAPIKeys []string `mapstructure:"allowed_api_keys"`

	// Rate limiting
	RateLimitEnabled   bool `mapstructure:"rate_limit_enabled"`
	RateLimitPerMinute int  `mapstructure:"rate_limit_per_minute"`

	// CORS (for SSE transport)
	CORSEnabled        bool     `mapstructure:"cors_enabled"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Name:            "TelemetryFlow-MCP",
			Version:         "1.1.2",
			Host:            "localhost",
			Port:            8080,
			Transport:       "stdio",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			Debug:           false,
		},
		Claude: ClaudeConfig{
			BaseURL:        "https://api.anthropic.com",
			DefaultModel:   "claude-sonnet-4-20250514",
			MaxTokens:      4096,
			Temperature:    1.0,
			TopP:           1.0,
			TopK:           0,
			Timeout:        120 * time.Second,
			MaxRetries:     3,
			RetryDelay:     1 * time.Second,
			EnableBatching: false,
		},
		MCP: MCPConfig{
			ProtocolVersion:        "2024-11-05",
			EnableTools:            true,
			EnableResources:        true,
			EnablePrompts:          true,
			EnableLogging:          true,
			EnableSampling:         false,
			MaxToolsPerSession:     100,
			MaxResourcesPerSession: 100,
			MaxPromptsPerSession:   50,
			MaxConversations:       10,
			MaxMessagesPerConv:     1000,
			ToolTimeout:            30 * time.Second,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stderr",
			AddSource:  false,
			TimeFormat: time.RFC3339,
		},
		Telemetry: TelemetryConfig{
			Enabled:         true,
			ServiceName:     "telemetryflow-mcp",
			Environment:     "development",
			OTLPEndpoint:    "localhost:4317",
			OTLPInsecure:    true,
			TraceSampleRate: 1.0,
			MetricsEnabled:  true,
			MetricsInterval: 30 * time.Second,
		},
		Security: SecurityConfig{
			RequireAPIKey:      false,
			RateLimitEnabled:   true,
			RateLimitPerMinute: 100,
			CORSEnabled:        true,
			CORSAllowedOrigins: []string{"*"},
		},
	}
}

// Load loads configuration from files and environment
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	// Set config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config in standard locations
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/telemetryflow-mcp")
		v.AddConfigPath("$HOME/.telemetryflow-mcp")
	}

	// Environment variable settings
	v.SetEnvPrefix("TELEMETRYFLOW_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind specific environment variables
	bindEnvVars(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults and env vars
	}

	// Unmarshal config
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables for sensitive data
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.Claude.APIKey = apiKey
	}
	if apiKey := os.Getenv("TELEMETRYFLOW_MCP_CLAUDE_API_KEY"); apiKey != "" {
		config.Claude.APIKey = apiKey
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// bindEnvVars binds environment variables to config keys
func bindEnvVars(v *viper.Viper) {
	// Claude API (errors ignored as BindEnv only fails on empty key names)
	_ = v.BindEnv("claude.api_key", "ANTHROPIC_API_KEY", "TELEMETRYFLOW_MCP_CLAUDE_API_KEY")
	_ = v.BindEnv("claude.base_url", "TELEMETRYFLOW_MCP_CLAUDE_BASE_URL")
	_ = v.BindEnv("claude.default_model", "TELEMETRYFLOW_MCP_CLAUDE_DEFAULT_MODEL")

	// Server
	_ = v.BindEnv("server.host", "TELEMETRYFLOW_MCP_SERVER_HOST")
	_ = v.BindEnv("server.port", "TELEMETRYFLOW_MCP_SERVER_PORT")
	_ = v.BindEnv("server.transport", "TELEMETRYFLOW_MCP_SERVER_TRANSPORT")
	_ = v.BindEnv("server.debug", "TELEMETRYFLOW_MCP_DEBUG")

	// Logging
	_ = v.BindEnv("logging.level", "TELEMETRYFLOW_MCP_LOG_LEVEL")
	_ = v.BindEnv("logging.format", "TELEMETRYFLOW_MCP_LOG_FORMAT")

	// Telemetry
	_ = v.BindEnv("telemetry.enabled", "TELEMETRYFLOW_MCP_TELEMETRY_ENABLED")
	_ = v.BindEnv("telemetry.otlp_endpoint", "TELEMETRYFLOW_MCP_OTLP_ENDPOINT")
	_ = v.BindEnv("telemetry.service_name", "TELEMETRYFLOW_MCP_SERVICE_NAME")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Claude.APIKey == "" {
		return errors.New("claude.api_key is required (set ANTHROPIC_API_KEY environment variable)")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return errors.New("server.port must be between 1 and 65535")
	}

	validTransports := map[string]bool{"stdio": true, "sse": true, "websocket": true}
	if !validTransports[c.Server.Transport] {
		return errors.New("server.transport must be 'stdio', 'sse', or 'websocket'")
	}

	if c.Claude.MaxTokens < 1 {
		return errors.New("claude.max_tokens must be positive")
	}

	if c.Claude.Temperature < 0 || c.Claude.Temperature > 2 {
		return errors.New("claude.temperature must be between 0 and 2")
	}

	if c.Telemetry.TraceSampleRate < 0 || c.Telemetry.TraceSampleRate > 1 {
		return errors.New("telemetry.trace_sample_rate must be between 0 and 1")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Telemetry.Environment == "development" || c.Server.Debug
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Telemetry.Environment == "production"
}
