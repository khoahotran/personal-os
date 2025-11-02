package persistence

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/user"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type PostRepoIntegrationTestSuite struct {
	suite.Suite
	dbPool      *pgxpool.Pool
	pgContainer *postgres.PostgresContainer
	testLogger  logger.Logger
	postRepo    post.Repository
	userRepo    user.Repository
	testOwner   *user.User
}

func (s *PostRepoIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(1*time.Minute),
		),
	)
	if err != nil {
		s.T().Fatalf("Failed to start postgres container: %s", err)
	}
	s.pgContainer = pgContainer

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		s.T().Fatalf("Failed to get connection string: %s", err)
	}

	m, err := migrate.New("file://../../migrations", dsn)
	if err != nil {
		s.T().Fatalf("Failed to create migrate instance: %s", err)
	}
	if err := m.Up(); err != nil {

		log.Printf("Lỗi chạy migration, kiểm tra đường dẫn. Lỗi: %s", err)

		s.T().Fatalf("Failed to run migrations: %s", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		s.T().Fatalf("Failed to create pgxpool: %s", err)
	}
	s.dbPool = pool

	s.postRepo = NewPostgresPostRepo(s.dbPool, s.testLogger)
	s.userRepo = NewPostgresUserRepo(s.dbPool, s.testLogger)

	s.testOwner = &user.User{
		ID:           uuid.New(),
		Email:        "testowner@example.com",
		PasswordHash: "hashedpassword",
	}
	query := `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`
	_, err = s.dbPool.Exec(ctx, query, s.testOwner.ID, s.testOwner.Email, s.testOwner.PasswordHash)
	if err != nil {
		s.T().Fatalf("Failed to seed owner: %s", err)
	}
}

func (s *PostRepoIntegrationTestSuite) TearDownSuite() {
	if s.dbPool != nil {
		s.dbPool.Close()
	}
	if s.pgContainer != nil {
		if err := s.pgContainer.Terminate(context.Background()); err != nil {
			s.T().Fatalf("Failed to terminate postgres container: %s", err)
		}
	}
}

func TestPostRepoIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	suite.Run(t, new(PostRepoIntegrationTestSuite))
}

func (s *PostRepoIntegrationTestSuite) Test_Save_And_FindByID() {
	ctx := context.Background()

	newPost := &post.Post{
		ID:              uuid.New(),
		OwnerID:         s.testOwner.ID,
		Slug:            "my-first-post",
		Title:           "My First Post",
		ContentMarkdown: "Hello world",
		Status:          post.StatusDraft,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	err := s.postRepo.Save(ctx, newPost)
	s.NoError(err)

	foundPost, err := s.postRepo.FindByID(ctx, newPost.ID, s.testOwner.ID)

	s.NoError(err)
	s.NotNil(foundPost)
	s.Equal(newPost.Title, foundPost.Title)
	s.Equal(newPost.Slug, foundPost.Slug)
}

func (s *PostRepoIntegrationTestSuite) Test_ListPublic() {
	ctx := context.Background()

	publishedAtTime := time.Now()
	publicPost := &post.Post{
		ID: uuid.New(), OwnerID: s.testOwner.ID, Slug: "public-post", Title: "Public",
		Status: post.StatusPublic, PublishedAt: &publishedAtTime,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	privatePost := &post.Post{
		ID: uuid.New(), OwnerID: s.testOwner.ID, Slug: "private-post", Title: "Private",
		Status:    post.StatusPrivate,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	s.NoError(s.postRepo.Save(ctx, publicPost))
	s.NoError(s.postRepo.Save(ctx, privatePost))

	publicPosts, err := s.postRepo.ListPublic(ctx, 10, 0)

	s.NoError(err)
	s.Len(publicPosts, 1)
	s.Equal(publicPost.ID, publicPosts[0].ID)
}
