package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/adapters/media_storage"
	"github.com/khoahotran/personal-os/adapters/persistence"
	mediaUC "github.com/khoahotran/personal-os/internal/application/usecase/media"
	postUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
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
	mediaRepo := persistence.NewPostgresMediaRepo(dbPool)

	// Worker Use Case
	processPostEventUC := postUC.NewProcessPostEventUseCase(postRepo, uploader)
	processMediaEventUC := mediaUC.NewProcessMediaUseCase(mediaRepo, uploader)

	// Kafka Consumer
	postConsumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Kafka.Brokers,
		Topic:    event.TopicPostEvents,
		GroupID:  "post-processor-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer postConsumer.Close()

	mediaConsumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Kafka.Brokers,
		Topic:    event.TopicMediaEvents,
		GroupID:  "media-processor-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer mediaConsumer.Close()

	// Context and run

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Printf(" Worker listening on topic '%s'...", event.TopicPostEvents)

		for {

			select {
			case <-ctx.Done():
				log.Println("Stopping Post consumer...")
				return
			default:

				msg, err := postConsumer.FetchMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						break
					}
					log.Printf("ERROR: Failed to read message from %s: %v", event.TopicPostEvents, err)
					continue
				}

				log.Printf("Received POST event [Key: %s]", string(msg.Key))
				var payload event.PostEventPayload
				if err := json.Unmarshal(msg.Value, &payload); err != nil {
					log.Printf("ERROR: Failed to unmarshal post event: %v. Skipping.", err)
					commitMessage(postConsumer, msg)
					continue
				}

				log.Printf("Processing event: [%s] for PostID: %s", payload.EventType, payload.PostID)
				if err := processPostEventUC.Execute(ctx, payload); err != nil {
					log.Printf("ERROR: Failed to process post event for %s: %v", payload.PostID, err)
					continue
				}
				commitMessage(postConsumer, msg)
			}
		}
	}()

	go func() {
		defer wg.Done()
		log.Printf("Worker listening on topic '%s'...", event.TopicMediaEvents)
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping Media consumer...")
				return
			default:
				msg, err := mediaConsumer.FetchMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						break
					}
					log.Printf("ERROR: Failed to read message from %s: %v", event.TopicMediaEvents, err)
					continue
				}

				log.Printf("📬 Received MEDIA event [Key: %s]", string(msg.Key))
				var payload event.MediaEventPayload
				if err := json.Unmarshal(msg.Value, &payload); err != nil {
					log.Printf("ERROR: Failed to unmarshal media event: %v. Skipping.", err)
					commitMessage(mediaConsumer, msg)
					continue
				}

				log.Printf("⚙️ Processing event: [%s] for MediaID: %s", payload.EventType, payload.MediaID)
				if err := processMediaEventUC.Execute(ctx, payload); err != nil {
					log.Printf("ERROR: Failed to process media event for %s: %v", payload.MediaID, err)
					continue
				}
				commitMessage(mediaConsumer, msg)
			}
		}
	}()
	wg.Wait()
}

func commitMessage(consumer *kafka.Reader, msg kafka.Message) {
	if err := consumer.CommitMessages(context.Background(), msg); err != nil {
		log.Printf("ERROR: Failed to commit message: %v", err)
	}
}
