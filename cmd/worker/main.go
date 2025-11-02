package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/adapters/media_storage"
	"github.com/khoahotran/personal-os/adapters/persistence"
	mediaUC "github.com/khoahotran/personal-os/internal/application/usecase/media"
	postUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/pkg/logger"
)

func main() {
	fmt.Println("Starting Personal OS Worker...")

	// Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: cannot load config: %v", err)
	}

	// Logger
	appLogger := logger.NewZapLogger("development")
	appLogger.Info("Worker Logger initialized")

	// Database
	dbPool, err := persistence.NewPostgresPool(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("FATAL: cannot connect Postgres", err)
	}
	defer dbPool.Close()

	// Cloudinary Uploader
	uploader, err := media_storage.NewCloudinaryAdapter(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("FATAL: Failed to initialize uploader", err)
	}

	// Repositories
	postRepo := persistence.NewPostgresPostRepo(dbPool, appLogger)
	mediaRepo := persistence.NewPostgresMediaRepo(dbPool, appLogger)

	// Worker Use Case
	processPostEventUC := postUC.NewProcessPostEventUseCase(postRepo, uploader, appLogger)
	processMediaEventUC := mediaUC.NewProcessMediaUseCase(mediaRepo, uploader, appLogger)

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
		appLogger.Info("Worker listening on topic", zap.String("topic", event.TopicPostEvents))

		for {

			select {
			case <-ctx.Done():
				appLogger.Info("Stopping Post consumer...")
				return
			default:

				msg, err := postConsumer.FetchMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					appLogger.Error("Failed to read message", err, zap.String("topic", event.TopicPostEvents))
					continue
				}

				appLogger.Info("Received message", zap.String("topic", msg.Topic), zap.String("key", string(msg.Key)))
				var payload event.PostEventPayload
				if err := json.Unmarshal(msg.Value, &payload); err != nil {
					appLogger.Error("Failed to unmarshal post event", err, zap.ByteString("value", msg.Value))
					commitMessage(postConsumer, msg, appLogger)
					continue
				}

				l := appLogger.With(zap.String("post_id", payload.PostID.String()), zap.String("event_type", string(payload.EventType)))
				l.Info("Processing event")
				if err := processPostEventUC.Execute(ctx, payload); err != nil {
					l.Error("Failed to process post event", err)
					continue
				}
				commitMessage(postConsumer, msg, appLogger)
			}
		}
	}()

	go func() {
		defer wg.Done()
		appLogger.Info("Worker listening on topic", zap.String("topic", event.TopicMediaEvents))
		for {
			select {
			case <-ctx.Done():
				appLogger.Info("Stopping Media consumer...")
				return
			default:
				msg, err := mediaConsumer.FetchMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					appLogger.Error("Failed to read message", err, zap.String("topic", event.TopicMediaEvents))
					continue
				}

				appLogger.Info("Received message", zap.String("topic", msg.Topic), zap.String("key", string(msg.Key)))
				var payload event.MediaEventPayload
				if err := json.Unmarshal(msg.Value, &payload); err != nil {
					appLogger.Error("Failed to unmarshal media event", err, zap.ByteString("value", msg.Value))
					commitMessage(mediaConsumer, msg, appLogger)
					continue
				}

				l := appLogger.With(zap.String("media_id", payload.MediaID.String()), zap.String("event_type", string(payload.EventType)))
				l.Info("Processing event")
				if err := processMediaEventUC.Execute(ctx, payload); err != nil {
					l.Error("Failed to process media event", err)
					continue
				}
				commitMessage(mediaConsumer, msg, appLogger)
			}
		}
	}()

	go func() {
		defer wg.Done()
		appLogger.Info("Worker listening on topic", zap.String("topic", event.TopicMediaEvents))
		for {
			select {
			case <-ctx.Done():
				appLogger.Info("Stopping Media consumer...")
				return
			default:
				msg, err := mediaConsumer.FetchMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					appLogger.Error("Failed to read message", err, zap.String("topic", event.TopicMediaEvents))
					continue
				}

				appLogger.Info("Received message", zap.String("topic", msg.Topic), zap.String("key", string(msg.Key)))
				var payload event.MediaEventPayload
				if err := json.Unmarshal(msg.Value, &payload); err != nil {
					appLogger.Error("Failed to unmarshal media event", err, zap.ByteString("value", msg.Value))
					commitMessage(mediaConsumer, msg, appLogger)
					continue
				}

				l := appLogger.With(zap.String("media_id", payload.MediaID.String()), zap.String("event_type", string(payload.EventType)))
				l.Info("Processing event")
				if err := processMediaEventUC.Execute(ctx, payload); err != nil {
					l.Error("Failed to process media event", err)
					continue
				}
				commitMessage(mediaConsumer, msg, appLogger)
			}
		}
	}()

	// Ctrl+C
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	appLogger.Info("Worker started. Press Ctrl+C to exit.")
	<-sigterm

	// Shuting down
	appLogger.Info("Shutting down workers...")
	cancel()
	wg.Wait()
	appLogger.Info("All workers stopped.")
}

func commitMessage(consumer *kafka.Reader, msg kafka.Message, log logger.Logger) {
	if err := consumer.CommitMessages(context.Background(), msg); err != nil {
		log.Error("Failed to commit message", err, zap.String("topic", msg.Topic), zap.Int64("offset", msg.Offset))
	}
}
