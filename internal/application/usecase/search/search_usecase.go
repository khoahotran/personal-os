package search

import (
	"context"

	"github.com/google/uuid"

	"github.com/khoahotran/personal-os/internal/domain/search"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type SearchUseCase struct {
	searchRepo search.Repository
	logger     logger.Logger
}

func NewSearchUseCase(sr search.Repository, log logger.Logger) *SearchUseCase {
	return &SearchUseCase{
		searchRepo: sr,
		logger:     log,
	}
}

type SearchInput struct {
	Query    string
	OwnerID  uuid.UUID
	IsPublic bool
	Limit    int
}

type SearchOutput struct {
	Results []search.SearchResult
}

func (uc *SearchUseCase) Execute(ctx context.Context, input SearchInput) (*SearchOutput, error) {
	if input.Query == "" {
		return &SearchOutput{Results: []search.SearchResult{}}, nil
	}
	if input.Limit <= 0 {
		input.Limit = 10
	}

	var results []search.SearchResult
	var err error

	if input.IsPublic {

		uc.logger.Info("Executing public search", zap.String("query", input.Query))
		results, err = uc.searchRepo.SearchPublic(ctx, input.Query, input.Limit)
	} else {

		uc.logger.Info("Executing private search", zap.String("query", input.Query), zap.String("owner_id", input.OwnerID.String()))
		results, err = uc.searchRepo.SearchPrivate(ctx, input.Query, input.OwnerID, input.Limit)
	}

	if err != nil {
		uc.logger.Error("Search execution failed", err)
		return nil, apperror.NewInternal("search failed", err)
	}

	return &SearchOutput{Results: results}, nil
}
