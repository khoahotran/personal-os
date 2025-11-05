package service

import (
	"context"
)

type LLMService interface {
	GenerateChatResponse(ctx context.Context, prompt string) (string, error)
}
