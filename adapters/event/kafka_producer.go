package event

import (
	"fmt"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/segmentio/kafka-go"
)

const (
	TopicPostEvents  = "post.events"
	TopicMediaEvents = "media.events"
	TopicViewEvents  = "view.events"	
)

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

	// writer 'post.events'
	postWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    TopicPostEvents,
		Balancer: &kafka.LeastBytes{},
	}

	// writer 'media.events'
	mediaWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    TopicMediaEvents,
		Balancer: &kafka.LeastBytes{},
	}

	// writer 'view.events'
	viewWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    TopicViewEvents,
		Balancer: &kafka.LeastBytes{},
	}

	fmt.Println("Initialize Kafka Producers successfully.")

	return &KafkaProducerClient{
		PostEventsWriter:  postWriter,
		MediaEventsWriter: mediaWriter,
		ViewEventsWriter:  viewWriter,
	}, nil
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