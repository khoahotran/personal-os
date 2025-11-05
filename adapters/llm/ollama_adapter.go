package llm

import (
	"context"
	"fmt"

	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

type ollamaLLMAdapter struct {
	client *openai.Client
	log    logger.Logger
}

// NewOllamaLLMAdapter là constructor
func NewOllamaLLMAdapter(cfg config.Config, log logger.Logger) (service.LLMService, error) {
	if cfg.Ollama.Host == "" {
		return nil, fmt.Errorf("ollama Host is not configured")
	}

	config := openai.DefaultConfig("dummy-key")
	config.BaseURL = cfg.Ollama.Host

	client := openai.NewClientWithConfig(config)

	log.Info("Ollama Chat (LLM) Adapter initialized")
	return &ollamaLLMAdapter{client: client, log: log}, nil
}

func (a *ollamaLLMAdapter) GenerateChatResponse(ctx context.Context, prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: "phi3:mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Stream: false,
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("ollama chat completion request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("ollama returned no chat choices")
	}

	return resp.Choices[0].Message.Content, nil
}
