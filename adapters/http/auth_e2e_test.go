package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoahotran/personal-os/adapters/persistence"
	authUC "github.com/khoahotran/personal-os/internal/application/usecase/auth"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/internal/domain/user"
	"github.com/khoahotran/personal-os/pkg/auth"
	"github.com/khoahotran/personal-os/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthE2ETestSuite struct {
	suite.Suite
	Router   *gin.Engine
	testUser user.User
	testPass string
}

func (s *AuthE2ETestSuite) SetupSuite() {

	cfg, err := config.LoadConfig("../..")
	if err != nil {
		s.T().Fatalf("Failed to load config for E2E test: %v", err)
	}

	dbPool, err := pgxpool.New(context.Background(), cfg.DB.DSN)
	if err != nil {
		s.T().Fatalf("E2E test failed to connect postgres: %v", err)
	}

	appLogger := logger.NewZapLogger("development")

	s.testPass = "e2e_test_password_123"
	hash, _ := auth.HashPassword(s.testPass)
	s.testUser = user.User{
		ID:           uuid.New(),
		Email:        "e2e_test@example.com",
		PasswordHash: hash,
	}
	query := `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3) ON CONFLICT (email) DO UPDATE SET password_hash = $3`
	_, err = dbPool.Exec(context.Background(), query, s.testUser.ID, s.testUser.Email, s.testUser.PasswordHash)
	if err != nil {
		s.T().Fatalf("E2E test failed to seed user: %v", err)
	}

	userRepo := persistence.NewPostgresUserRepo(dbPool, appLogger)
	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenLifespan)
	loginUseCase := authUC.NewLoginUseCase(userRepo, jwtSvc, appLogger)
	authHandler := NewAuthHandler(loginUseCase, appLogger)
	authMiddleware := AuthMiddleware(jwtSvc, appLogger)
	errorMiddleware := ErrorMiddleware(appLogger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(errorMiddleware)

	api := router.Group("/api")
	{
		admin := api.Group("/admin")
		{
			admin.POST("/auth/login", authHandler.Login)
			adminPrivate := admin.Group("/")
			adminPrivate.Use(authMiddleware)
			{
				adminPrivate.GET("/health-auth", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"status": "OK"})
				})
			}
		}
	}

	s.Router = router
}

func (s *AuthE2ETestSuite) TearDownSuite() {}

func TestAuthE2E(t *testing.T) {

	if os.Getenv("E2E_TESTS") == "" {
		t.Skip("Skipping E2E tests. Set E2E_TESTS=1 to run.")
	}
	suite.Run(t, new(AuthE2ETestSuite))
}

func (s *AuthE2ETestSuite) Test_Login_Flow() {

	bodyBad, _ := json.Marshal(gin.H{"email": s.testUser.Email, "password": "wrongpassword"})
	reqBad := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", bytes.NewBuffer(bodyBad))
	reqBad.Header.Set("Content-Type", "application/json")

	rrBad := httptest.NewRecorder()
	s.Router.ServeHTTP(rrBad, reqBad)

	assert.Equal(s.T(), http.StatusUnauthorized, rrBad.Code)

	bodyGood, _ := json.Marshal(gin.H{"email": s.testUser.Email, "password": s.testPass})
	reqGood := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", bytes.NewBuffer(bodyGood))
	reqGood.Header.Set("Content-Type", "application/json")

	rrGood := httptest.NewRecorder()
	s.Router.ServeHTTP(rrGood, reqGood)

	assert.Equal(s.T(), http.StatusOK, rrGood.Code)

	var loginResponse map[string]string
	json.Unmarshal(rrGood.Body.Bytes(), &loginResponse)
	accessToken := loginResponse["access_token"]
	assert.NotEmpty(s.T(), accessToken)

	reqAuth := httptest.NewRequest(http.MethodGet, "/api/admin/health-auth", nil)
	reqAuth.Header.Set("Authorization", "Bearer "+accessToken)

	rrAuth := httptest.NewRecorder()
	s.Router.ServeHTTP(rrAuth, reqAuth)

	assert.Equal(s.T(), http.StatusOK, rrAuth.Code)

	reqNoAuth := httptest.NewRequest(http.MethodGet, "/api/admin/health-auth", nil)
	rrNoAuth := httptest.NewRecorder()
	s.Router.ServeHTTP(rrNoAuth, reqNoAuth)

	assert.Equal(s.T(), http.StatusUnauthorized, rrNoAuth.Code)
}
