package service

import (
	"context"
	"github.com/pgvector/pgvector-go"
)

type EmbeddingService interface {
	GenerateEmbeddings(ctx context.Context, text string) (pgvector.Vector, error)
}
