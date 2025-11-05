package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type ChatUseCase struct {
	embedder service.EmbeddingService
	llm      service.LLMService
	postRepo post.Repository
	logger   logger.Logger
}

func NewChatUseCase(
	em service.EmbeddingService,
	llm service.LLMService,
	pr post.Repository,
	log logger.Logger,
) *ChatUseCase {
	return &ChatUseCase{
		embedder: em,
		llm:      llm,
		postRepo: pr,
		logger:   log,
	}
}

type ChatInput struct {
	Query   string
	OwnerID uuid.UUID
	Limit   int
}

type ChatOutput struct {
	Response string       `json:"response"`
	Sources  []*post.Post `json:"sources"`
}

func (uc *ChatUseCase) Execute(ctx context.Context, input ChatInput) (*ChatOutput, error) {
	l := uc.logger.With(zap.String("query", input.Query))
	l.Info("ChatUseCase received query")

	l.Info("Generating embedding for query...")
	queryVector, err := uc.embedder.GenerateEmbeddings(ctx, input.Query)
	if err != nil {
		l.Error("Failed to generate query embedding", err)
		return nil, apperror.NewInternal("failed to process query embedding", err)
	}
	l.Info("Query embedding generated")

	if input.Limit <= 0 {
		input.Limit = 3
	}

	sources, err := uc.postRepo.SearchByEmbedding(ctx, queryVector, input.OwnerID, input.Limit)
	if err != nil {
		l.Error("Failed to search by embedding", err)
		return nil, apperror.NewInternal("failed to retrieve relevant documents", err)
	}
	l.Info("Found relevant sources", zap.Int("count", len(sources)))

	prompt := uc.buildPrompt(input.Query, sources)
	l.Info("Prompt built for LLM")

	l.Info("Generating response from LLM...")
	response, err := uc.llm.GenerateChatResponse(ctx, prompt)
	if err != nil {
		return nil, apperror.NewInternal("failed to generate LLM response", err)
	}
	l.Info("LLM response generated")

	return &ChatOutput{
		Response: response,
		Sources:  sources,
	}, nil
}

func (uc *ChatUseCase) buildPrompt(query string, sources []*post.Post) string {
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Based on the following contexts:\n\n")
	for i, s := range sources {
		contextBuilder.WriteString(fmt.Sprintf("--- Context %d (Title: %s) ---\n", i+1, s.Title))
		contextBuilder.WriteString(s.ContentMarkdown)
		contextBuilder.WriteString("\n\n")
	}

	var promptBuilder strings.Builder
	promptBuilder.WriteString(contextBuilder.String())
	promptBuilder.WriteString("--- Question ---\n")
	promptBuilder.WriteString(query)
	promptBuilder.WriteString("\n\n--- Answer ---\n")
	promptBuilder.WriteString("Please answer the question above based only on the provided contexts:")

	return promptBuilder.String()
}
