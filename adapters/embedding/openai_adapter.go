package embedding

import (
	"context"
	"fmt"

	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/pkg/logger"
	"github.com/pgvector/pgvector-go"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type ollamaAdapter struct {
	client *openai.Client
	log    logger.Logger
}

func NewOllamaAdapter(cfg config.Config, log logger.Logger) (service.EmbeddingService, error) {
	if cfg.Ollama.Host == "" {
		return nil, fmt.Errorf("ollama Host is not configured")
	}

	config := openai.DefaultConfig("dummy-key")
	config.BaseURL = cfg.Ollama.Host

	client := openai.NewClientWithConfig(config)

	log.Info("Ollama Embedding Adapter initialized", zap.String("host", cfg.Ollama.Host))
	return &ollamaAdapter{client: client, log: log}, nil
}

func (a *ollamaAdapter) GenerateEmbeddings(ctx context.Context, text string) (pgvector.Vector, error) {
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: "nomic-embed-text",
	}

	resp, err := a.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("ollama embedding request failed: %w", err)
	}
	if len(resp.Data) == 0 {
		return pgvector.Vector{}, fmt.Errorf("ollama returned no embeddings")
	}
	return pgvector.NewVector(resp.Data[0].Embedding), nil
}
