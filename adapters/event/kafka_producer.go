package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const (
	TopicPostEvents  = "post.events"
	TopicMediaEvents = "media.events"
	TopicViewEvents  = "view.events"
)

type PostEventType string

const (
	PostEventTypeCreated   PostEventType = "post.created"
	PostEventTypeUpdated   PostEventType = "post.updated"
	PostEventTypeDeleted   PostEventType = "post.deleted"
	PostEventTypePublished PostEventType = "post.published"
)

type PostEventPayload struct {
	EventType PostEventType `json:"event_type"`
	PostID    uuid.UUID     `json:"post_id"`
	OwnerID   uuid.UUID     `json:"owner_id"`
}

type MediaEventType string

const (
	MediaEventTypeUploaded MediaEventType = "media.uploaded"
	MediaEventTypeDeleted  MediaEventType = "media.deleted"
)

type MediaEventPayload struct {
	EventType        MediaEventType `json:"event_type"`
	MediaID          uuid.UUID      `json:"media_id"`
	OwnerID          uuid.UUID      `json:"owner_id"`
	Provider         string         `json:"provider"`
	OriginalURL      string         `json:"original_url"`
	OriginalPublicID string         `json:"original_public_id"`
}

type KafkaProducerClient struct {
	PostEventsWriter  *kafka.Writer
	MediaEventsWriter *kafka.Writer
	ViewEventsWriter  *kafka.Writer
	logger            logger.Logger
}

func NewKafkaProducerClient(cfg config.Config, log logger.Logger) (*KafkaProducerClient, error) {
	brokers := cfg.Kafka.Brokers
	if len(brokers) == 0 {
		return nil, fmt.Errorf("config Kafka brokers not found")
	}

	createWriter := func(topic string) *kafka.Writer {
		return &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
			Async:    true,
		}
	}

	client := &KafkaProducerClient{
		PostEventsWriter:  createWriter(TopicPostEvents),
		MediaEventsWriter: createWriter(TopicMediaEvents),
		ViewEventsWriter:  createWriter(TopicViewEvents),
		logger:            log,
	}

	log.Info("Initialize Kafka Producers successfully.")
	return client, nil
}

func (c *KafkaProducerClient) PublishPostEvent(ctx context.Context, payload PostEventPayload) error {
	msgBody, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error("Kafka Marshal (Post) failed", err, zap.String("post_id", payload.PostID.String()))
		return err
	}

	err = c.PostEventsWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payload.PostID.String()),
		Value: msgBody,
	})

	if err != nil {
		c.logger.Error("Kafka Write (Post) failed", err, zap.String("post_id", payload.PostID.String()))
	} else {
		c.logger.Info("Kafka event sent", zap.String("topic", TopicPostEvents), zap.String("event_type", string(payload.EventType)))
	}
	return err
}

func (c *KafkaProducerClient) PublishMediaEvent(ctx context.Context, payload MediaEventPayload) error {
	msgBody, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error("Kafka Marshal (Media) failed", err, zap.String("media_id", payload.MediaID.String()))
		return err
	}

	err = c.MediaEventsWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payload.MediaID.String()),
		Value: msgBody,
	})

	if err != nil {
		c.logger.Error("Kafka Write (Media) failed", err, zap.String("media_id", payload.MediaID.String()))
	} else {
		c.logger.Info("Kafka event sent", zap.String("topic", TopicMediaEvents), zap.String("event_type", string(payload.EventType)))
	}
	return err
}

func (c *KafkaProducerClient) Close() {
	if c.PostEventsWriter != nil {
		c.PostEventsWriter.Close()
	}
	if c.MediaEventsWriter != nil {
		c.MediaEventsWriter.Close()
	}
	if c.ViewEventsWriter != nil {
		c.ViewEventsWriter.Close()
	}
	c.logger.Info("Closed Kafka Producers")
}
