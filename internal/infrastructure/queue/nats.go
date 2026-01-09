// Package queue provides NATS-based job queue for the MCP server.
//
// TelemetryFlow MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Common errors
var (
	ErrQueueDisabled     = errors.New("queue is disabled")
	ErrInvalidTask       = errors.New("invalid task")
	ErrTaskNotFound      = errors.New("task not found")
	ErrSerializeFailed   = errors.New("failed to serialize payload")
	ErrDeserializeFailed = errors.New("failed to deserialize payload")
	ErrStreamNotFound    = errors.New("stream not found")
	ErrConsumerNotFound  = errors.New("consumer not found")
)

// TaskState represents the state of a task.
type TaskState string

const (
	TaskStatePending   TaskState = "pending"
	TaskStateActive    TaskState = "active"
	TaskStateCompleted TaskState = "completed"
	TaskStateFailed    TaskState = "failed"
	TaskStateRetry     TaskState = "retry"
)

// TaskPriority represents task priority levels.
type TaskPriority int

const (
	PriorityLow      TaskPriority = 1
	PriorityDefault  TaskPriority = 3
	PriorityHigh     TaskPriority = 6
	PriorityCritical TaskPriority = 9
)

// Stream names
const (
	StreamTasks     = "TASKS"
	StreamEvents    = "EVENTS"
	StreamTelemetry = "TELEMETRY"
)

// Subject prefixes
const (
	SubjectTaskPrefix      = "tasks"
	SubjectEventPrefix     = "events"
	SubjectTelemetryPrefix = "telemetry"
)

// TaskHandler is a function that handles a task.
type TaskHandler func(ctx context.Context, task *Task) error

// Task represents a job in the queue.
type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Subject   string                 `json:"subject"`
	Priority  TaskPriority           `json:"priority"`
	MaxRetry  int                    `json:"max_retry"`
	Retries   int                    `json:"retries"`
	Timeout   time.Duration          `json:"timeout"`
	Deadline  time.Time              `json:"deadline,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	Success   bool                   `json:"success"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Retries   int                    `json:"retries"`
	Timestamp time.Time              `json:"timestamp"`
}

// NATSConfig configures the NATS queue.
type NATSConfig struct {
	// URL is the NATS server URL
	URL string `mapstructure:"url" yaml:"url" json:"url"`
	// URLs is a list of NATS server URLs for clustering
	URLs []string `mapstructure:"urls" yaml:"urls" json:"urls"`
	// Name is the client connection name
	Name string `mapstructure:"name" yaml:"name" json:"name"`
	// Token is the authentication token
	Token string `mapstructure:"token" yaml:"token" json:"token"`
	// Username for authentication
	Username string `mapstructure:"username" yaml:"username" json:"username"`
	// Password for authentication
	Password string `mapstructure:"password" yaml:"password" json:"password"`
	// Enabled enables or disables the queue
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	// MaxReconnects is the maximum number of reconnection attempts
	MaxReconnects int `mapstructure:"max_reconnects" yaml:"max_reconnects" json:"max_reconnects"`
	// ReconnectWait is the wait time between reconnection attempts
	ReconnectWait time.Duration `mapstructure:"reconnect_wait" yaml:"reconnect_wait" json:"reconnect_wait"`
	// Timeout is the connection timeout
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	// StreamRetention is the retention policy for streams
	StreamRetention string `mapstructure:"stream_retention" yaml:"stream_retention" json:"stream_retention"`
	// StreamMaxAge is the maximum age of messages in streams
	StreamMaxAge time.Duration `mapstructure:"stream_max_age" yaml:"stream_max_age" json:"stream_max_age"`
	// StreamMaxMsgs is the maximum number of messages in streams
	StreamMaxMsgs int64 `mapstructure:"stream_max_msgs" yaml:"stream_max_msgs" json:"stream_max_msgs"`
	// StreamMaxBytes is the maximum size of streams in bytes
	StreamMaxBytes int64 `mapstructure:"stream_max_bytes" yaml:"stream_max_bytes" json:"stream_max_bytes"`
	// AckWait is the time to wait for acknowledgment
	AckWait time.Duration `mapstructure:"ack_wait" yaml:"ack_wait" json:"ack_wait"`
	// MaxDeliver is the maximum number of delivery attempts
	MaxDeliver int `mapstructure:"max_deliver" yaml:"max_deliver" json:"max_deliver"`
}

// DefaultNATSConfig returns default configuration.
func DefaultNATSConfig() *NATSConfig {
	return &NATSConfig{
		URL:             "nats://localhost:4222",
		Name:            "tfo-mcp",
		Enabled:         true,
		MaxReconnects:   60,
		ReconnectWait:   2 * time.Second,
		Timeout:         5 * time.Second,
		StreamRetention: "limits",
		StreamMaxAge:    24 * time.Hour,
		StreamMaxMsgs:   100000,
		StreamMaxBytes:  1024 * 1024 * 1024, // 1GB
		AckWait:         30 * time.Second,
		MaxDeliver:      3,
	}
}

// NATSQueue provides NATS JetStream-based job queue functionality.
type NATSQueue struct {
	conn        *nats.Conn
	js          jetstream.JetStream
	config      *NATSConfig
	handlers    map[string]TaskHandler
	consumers   map[string]jetstream.Consumer
	streams     map[string]jetstream.Stream
	enabled     bool
	running     bool
	mu          sync.RWMutex
	initialized bool
	cancelFuncs []context.CancelFunc
}

// NewNATSQueue creates a new NATS-based queue.
func NewNATSQueue(cfg *NATSConfig) (*NATSQueue, error) {
	if cfg == nil {
		cfg = DefaultNATSConfig()
	}

	if !cfg.Enabled {
		return &NATSQueue{
			enabled:   false,
			handlers:  make(map[string]TaskHandler),
			consumers: make(map[string]jetstream.Consumer),
			streams:   make(map[string]jetstream.Stream),
		}, nil
	}

	return &NATSQueue{
		config:    cfg,
		handlers:  make(map[string]TaskHandler),
		consumers: make(map[string]jetstream.Consumer),
		streams:   make(map[string]jetstream.Stream),
		enabled:   true,
	}, nil
}

// Initialize initializes the NATS connection and JetStream.
func (q *NATSQueue) Initialize(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return nil
	}

	if q.initialized {
		return nil
	}

	// Build connection options
	opts := []nats.Option{
		nats.Name(q.config.Name),
		nats.Timeout(q.config.Timeout),
		nats.MaxReconnects(q.config.MaxReconnects),
		nats.ReconnectWait(q.config.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("NATS disconnected: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %s\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Println("NATS connection closed")
		}),
	}

	// Add authentication if configured
	if q.config.Token != "" {
		opts = append(opts, nats.Token(q.config.Token))
	} else if q.config.Username != "" && q.config.Password != "" {
		opts = append(opts, nats.UserInfo(q.config.Username, q.config.Password))
	}

	// Connect to NATS
	var err error
	if len(q.config.URLs) > 0 {
		q.conn, err = nats.Connect(q.config.URLs[0], opts...)
	} else {
		q.conn, err = nats.Connect(q.config.URL, opts...)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	q.js, err = jetstream.New(q.conn)
	if err != nil {
		q.conn.Close()
		return fmt.Errorf("failed to create JetStream context: %w", err)
	}

	// Create default streams
	if err := q.createDefaultStreams(ctx); err != nil {
		q.conn.Close()
		return fmt.Errorf("failed to create streams: %w", err)
	}

	q.initialized = true
	return nil
}

// createDefaultStreams creates the default JetStream streams.
func (q *NATSQueue) createDefaultStreams(ctx context.Context) error {
	streams := []struct {
		name     string
		subjects []string
	}{
		{StreamTasks, []string{SubjectTaskPrefix + ".>"}},
		{StreamEvents, []string{SubjectEventPrefix + ".>"}},
		{StreamTelemetry, []string{SubjectTelemetryPrefix + ".>"}},
	}

	for _, s := range streams {
		cfg := jetstream.StreamConfig{
			Name:        s.name,
			Description: fmt.Sprintf("TFO-MCP %s stream", s.name),
			Subjects:    s.subjects,
			Retention:   jetstream.LimitsPolicy,
			MaxAge:      q.config.StreamMaxAge,
			MaxMsgs:     q.config.StreamMaxMsgs,
			MaxBytes:    q.config.StreamMaxBytes,
			Storage:     jetstream.FileStorage,
			Replicas:    1,
			Discard:     jetstream.DiscardOld,
		}

		stream, err := q.js.CreateOrUpdateStream(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to create stream %s: %w", s.name, err)
		}
		q.streams[s.name] = stream
	}

	return nil
}

// Close closes the NATS connection.
func (q *NATSQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled || q.conn == nil {
		return nil
	}

	// Cancel all consumer contexts
	for _, cancel := range q.cancelFuncs {
		cancel()
	}

	// Drain and close connection
	if err := q.conn.Drain(); err != nil {
		q.conn.Close()
	}

	q.initialized = false
	q.running = false
	return nil
}

// RegisterHandler registers a task handler for a specific task type.
func (q *NATSQueue) RegisterHandler(taskType string, handler TaskHandler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[taskType] = handler
}

// StartConsumer starts a consumer for processing tasks.
func (q *NATSQueue) StartConsumer(ctx context.Context, streamName, consumerName, filterSubject string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.isReadyLocked() {
		return ErrQueueDisabled
	}

	stream, ok := q.streams[streamName]
	if !ok {
		return ErrStreamNotFound
	}

	// Create consumer configuration
	cfg := jetstream.ConsumerConfig{
		Name:          consumerName,
		Durable:       consumerName,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       q.config.AckWait,
		MaxDeliver:    q.config.MaxDeliver,
		DeliverPolicy: jetstream.DeliverAllPolicy,
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	q.consumers[consumerName] = consumer

	// Start consuming in a goroutine
	consumerCtx, cancel := context.WithCancel(ctx)
	q.cancelFuncs = append(q.cancelFuncs, cancel)

	go q.consumeMessages(consumerCtx, consumer)

	q.running = true
	return nil
}

// consumeMessages processes messages from a consumer.
func (q *NATSQueue) consumeMessages(ctx context.Context, consumer jetstream.Consumer) {
	iter, err := consumer.Messages()
	if err != nil {
		fmt.Printf("Failed to get message iterator: %v\n", err)
		return
	}
	defer iter.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := iter.Next()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				fmt.Printf("Error getting message: %v\n", err)
				continue
			}

			q.processMessage(ctx, msg)
		}
	}
}

// processMessage processes a single message.
func (q *NATSQueue) processMessage(ctx context.Context, msg jetstream.Msg) {
	var task Task
	if err := json.Unmarshal(msg.Data(), &task); err != nil {
		fmt.Printf("Failed to unmarshal task: %v\n", err)
		_ = msg.Term() // Terminal failure, don't retry
		return
	}

	q.mu.RLock()
	handler, ok := q.handlers[task.Type]
	q.mu.RUnlock()

	if !ok {
		fmt.Printf("No handler for task type: %s\n", task.Type)
		_ = msg.Term()
		return
	}

	// Execute handler with timeout
	execCtx := ctx
	if task.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, task.Timeout)
		defer cancel()
	}

	startTime := time.Now()
	err := handler(execCtx, &task)
	duration := time.Since(startTime)

	if err != nil {
		metadata, _ := msg.Metadata()
		if metadata != nil && metadata.NumDelivered >= uint64(q.config.MaxDeliver) { //nolint:gosec // MaxDeliver is always positive
			fmt.Printf("Task %s failed after max retries: %v\n", task.ID, err)
			_ = msg.Term()
		} else {
			fmt.Printf("Task %s failed, will retry: %v\n", task.ID, err)
			_ = msg.Nak()
		}
		return
	}

	fmt.Printf("Task %s completed in %v\n", task.ID, duration)
	_ = msg.Ack()
}

// Publish publishes a task to the queue.
func (q *NATSQueue) Publish(ctx context.Context, task *Task) (string, error) {
	if !q.isReady() {
		return "", ErrQueueDisabled
	}

	if task == nil || task.Type == "" {
		return "", ErrInvalidTask
	}

	// Generate ID if not set
	if task.ID == "" {
		task.ID = generateTaskID()
	}

	// Set defaults
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.Subject == "" {
		task.Subject = fmt.Sprintf("%s.%s", SubjectTaskPrefix, task.Type)
	}

	// Serialize task
	data, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSerializeFailed, err)
	}

	// Publish to JetStream
	ack, err := q.js.Publish(ctx, task.Subject, data)
	if err != nil {
		return "", fmt.Errorf("failed to publish task: %w", err)
	}

	return fmt.Sprintf("%s:%d", ack.Stream, ack.Sequence), nil
}

// PublishDelayed publishes a task to be processed after a delay.
func (q *NATSQueue) PublishDelayed(ctx context.Context, task *Task, delay time.Duration) (string, error) {
	if !q.isReady() {
		return "", ErrQueueDisabled
	}

	// For delayed publishing, we use a scheduled approach
	// NATS JetStream doesn't have native delayed delivery, so we use a workaround
	go func() {
		select {
		case <-time.After(delay):
			_, _ = q.Publish(context.Background(), task)
		case <-ctx.Done():
			return
		}
	}()

	if task.ID == "" {
		task.ID = generateTaskID()
	}

	return task.ID, nil
}

// PublishEvent publishes an event to the events stream.
func (q *NATSQueue) PublishEvent(ctx context.Context, eventType string, payload map[string]interface{}) error {
	if !q.isReady() {
		return ErrQueueDisabled
	}

	event := map[string]interface{}{
		"type":      eventType,
		"payload":   payload,
		"timestamp": time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSerializeFailed, err)
	}

	subject := fmt.Sprintf("%s.%s", SubjectEventPrefix, eventType)
	_, err = q.js.Publish(ctx, subject, data)
	return err
}

// PublishTelemetry publishes telemetry data to the telemetry stream.
func (q *NATSQueue) PublishTelemetry(ctx context.Context, telemetryType string, data map[string]interface{}) error {
	if !q.isReady() {
		return ErrQueueDisabled
	}

	telemetry := map[string]interface{}{
		"type":      telemetryType,
		"data":      data,
		"timestamp": time.Now(),
	}

	payload, err := json.Marshal(telemetry)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSerializeFailed, err)
	}

	subject := fmt.Sprintf("%s.%s", SubjectTelemetryPrefix, telemetryType)
	_, err = q.js.Publish(ctx, subject, payload)
	return err
}

// GetStreamInfo returns information about a stream.
func (q *NATSQueue) GetStreamInfo(ctx context.Context, streamName string) (*jetstream.StreamInfo, error) {
	if !q.isReady() {
		return nil, ErrQueueDisabled
	}

	stream, ok := q.streams[streamName]
	if !ok {
		return nil, ErrStreamNotFound
	}

	return stream.Info(ctx)
}

// GetConsumerInfo returns information about a consumer.
func (q *NATSQueue) GetConsumerInfo(ctx context.Context, consumerName string) (*jetstream.ConsumerInfo, error) {
	if !q.isReady() {
		return nil, ErrQueueDisabled
	}

	consumer, ok := q.consumers[consumerName]
	if !ok {
		return nil, ErrConsumerNotFound
	}

	return consumer.Info(ctx)
}

// PurgeStream purges all messages from a stream.
func (q *NATSQueue) PurgeStream(ctx context.Context, streamName string) error {
	if !q.isReady() {
		return ErrQueueDisabled
	}

	stream, ok := q.streams[streamName]
	if !ok {
		return ErrStreamNotFound
	}

	return stream.Purge(ctx)
}

// DeleteStream deletes a stream.
func (q *NATSQueue) DeleteStream(ctx context.Context, streamName string) error {
	if !q.isReady() {
		return ErrQueueDisabled
	}

	if err := q.js.DeleteStream(ctx, streamName); err != nil {
		return err
	}

	delete(q.streams, streamName)
	return nil
}

// IsEnabled returns whether the queue is enabled.
func (q *NATSQueue) IsEnabled() bool {
	return q.enabled
}

// IsRunning returns whether the queue worker is running.
func (q *NATSQueue) IsRunning() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.running
}

// IsReady returns whether the queue is ready for use.
func (q *NATSQueue) IsReady() bool {
	return q.isReady()
}

func (q *NATSQueue) isReady() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.isReadyLocked()
}

func (q *NATSQueue) isReadyLocked() bool {
	return q.enabled && q.initialized && q.conn != nil && q.conn.IsConnected()
}

// Stats returns queue statistics.
func (q *NATSQueue) Stats(ctx context.Context) (map[string]interface{}, error) {
	if !q.isReady() {
		return nil, ErrQueueDisabled
	}

	stats := make(map[string]interface{})

	// Gather stream info
	streamStats := make(map[string]interface{})
	for name, stream := range q.streams {
		info, err := stream.Info(ctx)
		if err != nil {
			continue
		}
		streamStats[name] = map[string]interface{}{
			"messages":  info.State.Msgs,
			"bytes":     info.State.Bytes,
			"consumers": info.State.Consumers,
			"first_seq": info.State.FirstSeq,
			"last_seq":  info.State.LastSeq,
		}
	}
	stats["streams"] = streamStats

	// Gather consumer info
	consumerStats := make(map[string]interface{})
	for name, consumer := range q.consumers {
		info, err := consumer.Info(ctx)
		if err != nil {
			continue
		}
		consumerStats[name] = map[string]interface{}{
			"pending":     info.NumPending,
			"waiting":     info.NumWaiting,
			"ack_pending": info.NumAckPending,
			"redelivered": info.NumRedelivered,
			"delivered":   info.Delivered.Consumer,
		}
	}
	stats["consumers"] = consumerStats

	// Connection info
	stats["connection"] = map[string]interface{}{
		"connected": q.conn.IsConnected(),
		"server":    q.conn.ConnectedUrl(),
		"cluster":   q.conn.ConnectedClusterName(),
	}

	return stats, nil
}

// Conn returns the underlying NATS connection for advanced operations.
func (q *NATSQueue) Conn() *nats.Conn {
	return q.conn
}

// JetStream returns the JetStream context for advanced operations.
func (q *NATSQueue) JetStream() jetstream.JetStream {
	return q.js
}

// generateTaskID generates a unique task ID.
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
