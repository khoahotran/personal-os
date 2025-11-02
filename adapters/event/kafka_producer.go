package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/segmentio/kafka-go"
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
}

func NewKafkaProducerClient(cfg config.Config) (*KafkaProducerClient, error) {
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
	}

	log.Println("Init Kafka Producers (Writers) successfully.")
	return client, nil
}

func (c *KafkaProducerClient) PublishPostEvent(ctx context.Context, payload PostEventPayload) error {
	msgBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Kafka Marshal failed: %v", err)
		return err
	}

	err = c.PostEventsWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payload.PostID.String()),
		Value: msgBody,
	})

	if err != nil {
		log.Printf("Kafka Write failed: %v", err)
	} else {
		log.Printf("Sent event to Kafka: %s for PostID: %s", payload.EventType, payload.PostID)
	}
	return err
}

func (c *KafkaProducerClient) PublishMediaEvent(ctx context.Context, payload MediaEventPayload) error {
	msgBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR Kafka Marshal (Media): %v", err)
		return err
	}

	err = c.MediaEventsWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payload.MediaID.String()),
		Value: msgBody,
	})

	if err != nil {
		log.Printf("ERROR Kafka Write (Media): %v", err)
	} else {
		log.Printf("Send event to Kafka: %s for MediaID: %s", payload.EventType, payload.MediaID)
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
	fmt.Println("Closed Kafka Producers")
}
