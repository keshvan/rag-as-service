package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// MessageWithEvent wraps a Kafka message and the decoded event
type MessageWithEvent struct {
	Message kafka.Message
	Event   *DocumentUploadedEvent
}

// Consumer reads messages from a Kafka topic
type Consumer struct {
	reader *kafka.Reader
	logger *slog.Logger
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic, groupID string, logger *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		StartOffset:    kafka.LastOffset,
		CommitInterval: 0, // Manual commit only via CommitMessages
	})

	return &Consumer{
		reader: reader,
		logger: logger,
	}
}

// ReadMessage reads a single message from the topic and unmarshals it into an event
func (c *Consumer) ReadMessage(ctx context.Context) (*MessageWithEvent, error) {
	message, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	var event DocumentUploadedEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		c.logger.Error("Failed to unmarshal message", "error", err, "value", string(message.Value))
		return nil, err
	}

	return &MessageWithEvent{
		Message: message,
		Event:   &event,
	}, nil
}

// CommitMessages commits the offset for processed messages
func (c *Consumer) CommitMessages(ctx context.Context, msg kafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
