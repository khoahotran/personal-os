package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/adapters/media_storage"
	"github.com/khoahotran/personal-os/adapters/persistence"
	workerUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
	"github.com/khoahotran/personal-os/internal/config"
)

func main() {
	fmt.Println("Starting Personal OS Worker...")

	// Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: cannot load config: %v", err)
	}

	// Database
	dbPool, err := persistence.NewPostgresPool(cfg)
	if err != nil {
		log.Fatalf("FATAL: cannot connect Postgres: %v", err)
	}
	defer dbPool.Close()

	// Cloudinary Uploader
	uploader, err := media_storage.NewCloudinaryAdapter(cfg)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize uploader: %v", err)
	}

	// Repositories
	postRepo := persistence.NewPostgresPostRepo(dbPool)

	// Worker Use Case
	processPostEventUC := workerUC.NewProcessPostEventUseCase(postRepo, uploader)

	// Kafka Consumer
	postConsumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Kafka.Brokers,
		Topic:    event.TopicPostEvents,
		GroupID:  "post-processor-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer postConsumer.Close()

	log.Printf("Worker listening on topic '%s'...", event.TopicPostEvents)

	ctx := context.Background()
	for {
		msg, err := postConsumer.ReadMessage(ctx)
		if err != nil {
			log.Printf("ERROR: Failed to read message from Kafka: %v", err)
			continue
		}

		log.Printf("Received message from [Topic: %s], [Key: %s]", msg.Topic, string(msg.Key))

		var payload event.PostEventPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			log.Printf("ERROR: Failed to unmarshal event: %v. Skipping.", err)
			commitMessage(postConsumer, msg)
			continue
		}

		log.Printf("Processing event: [%s] for PostID: %s", payload.EventType, payload.PostID)

		err = processPostEventUC.Execute(ctx, payload)
		if err != nil {
			log.Printf("ERROR: Failed to process event for PostID %s: %v", payload.PostID, err)
			continue
		}

		commitMessage(postConsumer, msg)
	}
}

func commitMessage(consumer *kafka.Reader, msg kafka.Message) {
	if err := consumer.CommitMessages(context.Background(), msg); err != nil {
		log.Printf("ERROR: Failed to commit message: %v", err)
	}
}
