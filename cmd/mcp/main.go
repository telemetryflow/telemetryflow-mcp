// Package main is the entry point for the TelemetryFlow GO MCP server
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/telemetryflow/telemetryflow-mcp/internal/application/handlers"
	"github.com/telemetryflow/telemetryflow-mcp/internal/infrastructure/claude"
	"github.com/telemetryflow/telemetryflow-mcp/internal/infrastructure/config"
	"github.com/telemetryflow/telemetryflow-mcp/internal/infrastructure/persistence"
	"github.com/telemetryflow/telemetryflow-mcp/internal/presentation/server"
	"github.com/telemetryflow/telemetryflow-mcp/internal/presentation/tools"
)

var (
	// Version information (set at build time)
	version   = "1.1.2"
	commit    = "unknown"
	buildDate = "unknown"

	// CLI flags
	configFile string
	debug      bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "tfo-mcp",
		Short:   "TelemetryFlow GO MCP Server",
		Long:    `TelemetryFlow GO MCP Server - Model Context Protocol server with Claude AI integration`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate),
		RunE:    runServer,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")

	// Add subcommands
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(validateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override debug if flag is set
	if debug {
		cfg.Server.Debug = true
		cfg.Logging.Level = "debug"
	}

	// Setup logger
	logger := setupLogger(cfg)
	logger.Info().
		Str("version", version).
		Str("transport", cfg.Server.Transport).
		Msg("Starting TelemetryFlow GO MCP Server")

	// Create Claude client
	claudeClient, err := claude.NewClient(&cfg.Claude, logger)
	if err != nil {
		return fmt.Errorf("failed to create Claude client: %w", err)
	}

	// Create repositories
	sessionRepo := persistence.NewInMemorySessionRepository()
	conversationRepo := persistence.NewInMemoryConversationRepository()
	toolRepo := persistence.NewInMemoryToolRepository()

	// Create event publisher (simple implementation)
	eventPublisher := &simpleEventPublisher{logger: logger}

	// Create handlers
	sessionHandler := handlers.NewSessionHandler(sessionRepo, eventPublisher)
	toolHandler := handlers.NewToolHandler(sessionRepo, toolRepo, eventPublisher)
	conversationHandler := handlers.NewConversationHandler(sessionRepo, conversationRepo, claudeClient, eventPublisher)

	// Create and register built-in tools
	toolRegistry := tools.NewToolRegistry(claudeClient)
	for _, tool := range toolRegistry.GetTools() {
		ctx := context.Background()
		if err := toolRepo.Register(ctx, tool); err != nil {
			logger.Warn().Err(err).Str("tool", tool.Name().String()).Msg("Failed to register tool")
		}
		// Register handler
		toolHandler.RegisterToolHandler(tool.Name().String(), tool.Handler())
	}

	// Create server
	srv := server.NewServer(cfg, logger, sessionHandler, toolHandler, conversationHandler)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info().Msg("Shutdown signal received")
		cancel()
		srv.Stop()
	}()

	// Run server
	if err := srv.Run(ctx); err != nil {
		if err == server.ErrServerClosed || err == context.Canceled {
			logger.Info().Msg("Server stopped")
			return nil
		}
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func setupLogger(cfg *config.Config) zerolog.Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Create logger
	var logger zerolog.Logger

	if cfg.Logging.Format == "text" || cfg.Server.Debug {
		// Pretty print for development
		output := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: cfg.Logging.TimeFormat,
		}
		logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		// JSON for production
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	// Add service info
	logger = logger.With().
		Str("service", cfg.Telemetry.ServiceName).
		Str("version", version).
		Logger()

	return logger
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("TelemetryFlow GO MCP Server\n")
			fmt.Printf("Version:    %s\n", version)
			fmt.Printf("Commit:     %s\n", commit)
			fmt.Printf("Build Date: %s\n", buildDate)
		},
	}
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return fmt.Errorf("configuration is invalid: %w", err)
			}
			fmt.Printf("Configuration is valid!\n")
			fmt.Printf("Server:    %s:%d\n", cfg.Server.Host, cfg.Server.Port)
			fmt.Printf("Transport: %s\n", cfg.Server.Transport)
			fmt.Printf("Model:     %s\n", cfg.Claude.DefaultModel)
			return nil
		},
	}
}

// simpleEventPublisher is a simple event publisher implementation
type simpleEventPublisher struct {
	logger zerolog.Logger
}

func (p *simpleEventPublisher) Publish(ctx context.Context, event interface{}) error {
	p.logger.Debug().Interface("event", event).Msg("Event published")
	return nil
}
