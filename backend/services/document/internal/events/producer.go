package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// Producer publishes events to Kafka
type Producer struct {
	writer *kafka.Writer
	topic  string
	logger *slog.Logger
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, topic string, logger *slog.Logger) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
		topic:  topic,
		logger: logger,
	}, nil
}

// PublishDocumentUploaded publishes a document uploaded event
func (p *Producer) PublishDocumentUploaded(ctx context.Context, event DocumentUploadedEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal event", "error", err)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.DocumentID),
		Value: eventJSON,
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		p.logger.Error("Failed to publish message", "error", err, "document_id", event.DocumentID)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Info("Document uploaded event published",
		"document_id", event.DocumentID,
		"organization_id", event.OrganizationID,
	)

	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
