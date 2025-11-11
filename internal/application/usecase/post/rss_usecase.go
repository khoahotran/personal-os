package post

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/feeds"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type RSSUseCase struct {
	postRepo post.Repository
	logger   logger.Logger
}

func NewRSSUseCase(pRepo post.Repository, log logger.Logger) *RSSUseCase {
	return &RSSUseCase{
		postRepo: pRepo,
		logger:   log,
	}
}

func (uc *RSSUseCase) Execute(ctx context.Context) (*feeds.Feed, error) {
	uc.logger.Info("Generating RSS feed...")

	now := time.Now()
	feed := &feeds.Feed{
		Title:       "My Personal OS - Blog",
		Link:        &feeds.Link{Href: "http://localhost:8080/api"},
		Description: "Reflections and projects.",
		Author:      &feeds.Author{Name: "The Owner"},
		Created:     now,
	}

	limit := 20
	offset := 0
	posts, err := uc.postRepo.ListPublic(ctx, limit, offset)
	if err != nil {
		uc.logger.Error("Failed to list public posts for RSS", err)
		return nil, err
	}

	var feedItems []*feeds.Item
	for _, p := range posts {

		postURL := fmt.Sprintf("http://localhost:3000/blog/%s", p.Slug)

		item := &feeds.Item{
			Title:       p.Title,
			Link:        &feeds.Link{Href: postURL},
			Description: p.ContentMarkdown,
			Created:     p.CreatedAt,
		}
		if p.PublishedAt != nil {
			item.Created = *p.PublishedAt
		}

		feedItems = append(feedItems, item)
	}

	feed.Items = feedItems
	uc.logger.Info("RSS feed generated successfully", zap.Int("item_count", len(feed.Items)))
	return feed, nil
}
