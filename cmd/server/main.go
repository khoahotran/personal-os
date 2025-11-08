package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/khoahotran/personal-os/adapters/embedding"
	"github.com/khoahotran/personal-os/adapters/event"
	httpAdapter "github.com/khoahotran/personal-os/adapters/http"
	"github.com/khoahotran/personal-os/adapters/llm"
	"github.com/khoahotran/personal-os/adapters/media_storage"
	"github.com/khoahotran/personal-os/adapters/persistence"
	authUC "github.com/khoahotran/personal-os/internal/application/usecase/auth"
	chatUC "github.com/khoahotran/personal-os/internal/application/usecase/chat"
	hobbyUC "github.com/khoahotran/personal-os/internal/application/usecase/hobby"
	mediaUC "github.com/khoahotran/personal-os/internal/application/usecase/media"
	postUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
	profileUC "github.com/khoahotran/personal-os/internal/application/usecase/profile"
	projectUC "github.com/khoahotran/personal-os/internal/application/usecase/project"
	searchUC "github.com/khoahotran/personal-os/internal/application/usecase/search"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/pkg/auth"
	"github.com/khoahotran/personal-os/pkg/logger"
	"github.com/khoahotran/personal-os/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	fmt.Println("Start Personal OS API Server...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: cannot load config: %v", err)
	}

	// Logger
	appLogger := logger.NewZapLogger("development")
	appLogger.Info("Logger initialized")

	// Initialize dependencies
	dbPool, err := persistence.NewPostgresPool(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("FATAL: cannot connect Postgres", err)
	}
	defer dbPool.Close()

	redisClient, err := persistence.NewRedisClient(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("FATAL: cannot connect Redis", err)
	}
	defer redisClient.Close()

	kafkaClient, err := event.NewKafkaProducerClient(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("FATAL: cannot init Kafka", err)
	}
	defer kafkaClient.Close()

	tracerProvider, err := tracing.NewTracerProvider(cfg, appLogger, "personal-os-api")
	if err != nil {
		appLogger.Fatal("FATAL: Failed to initialize Tracer", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			appLogger.Error("Failed to shutdown Tracer", err)
		}
	}()

	// Repositories
	userRepo := persistence.NewPostgresUserRepo(dbPool, appLogger)
	profileRepo := persistence.NewPostgresProfileRepo(dbPool, appLogger)
	postRepo := persistence.NewPostgresPostRepo(dbPool, appLogger)
	tagRepo := persistence.NewPostgresTagRepo(dbPool, appLogger)
	projectRepo := persistence.NewPostgresProjectRepo(dbPool, appLogger)
	mediaRepo := persistence.NewPostgresMediaRepo(dbPool, appLogger)
	hobbyRepo := persistence.NewPostgresHobbyRepo(dbPool, appLogger)
	searchRepo := persistence.NewPostgresSearchRepo(dbPool, appLogger)

	// Services
	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenLifespan)
	uploader, err := media_storage.NewCloudinaryAdapter(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("FATAL: Failed to initialize uploader", err)
	}
	embedder, err := embedding.NewOllamaAdapter(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("FATAL: Failed to initialize Ollama adapter", err)
	}
	llmService, err := llm.NewOllamaLLMAdapter(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("FATAL: Failed to initialize Ollama LLM adapter", err)
	}

	// Use Cases
	loginUseCase := authUC.NewLoginUseCase(userRepo, jwtSvc, appLogger)
	profileUseCase := profileUC.NewProfileUseCase(profileRepo, appLogger)

	createPostUseCase := postUC.NewCreatePostUseCase(postRepo, tagRepo, kafkaClient, uploader, appLogger)
	listPostsUseCase := postUC.NewListPostsUseCase(postRepo, tagRepo, appLogger)
	listPublicPostsUseCase := postUC.NewListPublicPostsUseCase(postRepo, tagRepo, appLogger)
	updatePostUseCase := postUC.NewUpdatePostUseCase(postRepo, tagRepo, kafkaClient, appLogger)
	deletePostUseCase := postUC.NewDeletePostUseCase(postRepo, tagRepo, kafkaClient, appLogger)
	getPostUseCase := postUC.NewGetPostUseCase(postRepo, tagRepo, appLogger)
	getPublicPostUseCase := postUC.NewGetPublicPostUseCase(postRepo, tagRepo, appLogger)

	createProjectUseCase := projectUC.NewCreateProjectUseCase(projectRepo, tagRepo, appLogger)
	listProjectsUseCase := projectUC.NewListProjectsUseCase(projectRepo, appLogger)
	listPublicProjectsUseCase := projectUC.NewListPublicProjectsUseCase(projectRepo, appLogger)
	getProjectUseCase := projectUC.NewGetProjectUseCase(projectRepo, tagRepo, appLogger)
	getPublicProjectUseCase := projectUC.NewGetPublicProjectUseCase(projectRepo, tagRepo, appLogger)
	updateProjectUseCase := projectUC.NewUpdateProjectUseCase(projectRepo, tagRepo, appLogger)
	deleteProjectUseCase := projectUC.NewDeleteProjectUseCase(projectRepo, tagRepo, appLogger)

	uploadMediaUseCase := mediaUC.NewUploadMediaUseCase(mediaRepo, uploader, kafkaClient, appLogger)
	listPublicMediaUseCase := mediaUC.NewListPublicMediaUseCase(mediaRepo, appLogger)
	updateMediaUseCase := mediaUC.NewUpdateMediaUseCase(mediaRepo, appLogger)
	deleteMediaUseCase := mediaUC.NewDeleteMediaUseCase(mediaRepo, uploader, appLogger)

	hobbyUseCase := hobbyUC.NewHobbyUseCase(hobbyRepo, appLogger)
	chatUseCase := chatUC.NewChatUseCase(
		embedder,
		llmService,
		postRepo,
		appLogger,
	)
	searchUseCase := searchUC.NewSearchUseCase(searchRepo, appLogger)

	// HTTP Handlers
	authHandler := httpAdapter.NewAuthHandler(loginUseCase, appLogger)
	profileHandler := httpAdapter.NewProfileHandler(profileUseCase, appLogger)
	postHandler := httpAdapter.NewPostHandler(
		createPostUseCase,
		listPostsUseCase,
		listPublicPostsUseCase,
		updatePostUseCase,
		deletePostUseCase,
		getPostUseCase,
		getPublicPostUseCase,
		appLogger,
	)
	hobbyHandler := httpAdapter.NewHobbyHandler(hobbyUseCase, appLogger)

	projectHandler := httpAdapter.NewProjectHandler(
		createProjectUseCase, listProjectsUseCase, listPublicProjectsUseCase,
		getProjectUseCase, getPublicProjectUseCase, updateProjectUseCase, deleteProjectUseCase, appLogger,
	)

	mediaHandler := httpAdapter.NewMediaHandler(
		uploadMediaUseCase,
		listPublicMediaUseCase,
		updateMediaUseCase,
		deleteMediaUseCase,
		appLogger,
	)

	chatHandler := httpAdapter.NewChatHandler(
		chatUseCase,
		appLogger,
	)

	searchHandler := httpAdapter.NewSearchHandler(searchUseCase, appLogger)

	// Middleware
	authMiddleware := httpAdapter.AuthMiddleware(jwtSvc, appLogger)

	// Setup Gin router
	router := gin.Default()
	router.Use(httpAdapter.ErrorMiddleware(appLogger))
	router.Use(otelgin.Middleware("personal-os-api"))
	api := router.Group("/api")
	{

		admin := api.Group("/admin")
		{

			adminAuth := admin.Group("/auth")
			adminAuth.POST("/login", authHandler.Login)

			adminPrivate := admin.Group("/")
			adminPrivate.Use(authMiddleware)
			{

				adminPrivate.GET("/health-auth", func(c *gin.Context) {

					userID, ok := httpAdapter.GetOwnerIDFromGinContext(c)
					if !ok {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get user id from context"})
						return
					}
					c.JSON(http.StatusOK, gin.H{
						"status":   "OK",
						"message":  "Authentication middleware is working!",
						"owner_id": userID,
					})
				})
				adminPrivate.POST("/chat", chatHandler.Chat)

				adminPrivate.GET("/profile", profileHandler.GetProfile)
				adminPrivate.PUT("/profile", profileHandler.UpdateProfile)

				posts := adminPrivate.Group("/posts")
				{
					posts.POST("", postHandler.CreatePost)
					posts.GET("", postHandler.ListPosts)
					posts.PUT("/:id", postHandler.UpdatePost)
					posts.DELETE("/:id", postHandler.DeletePost)
					posts.GET("/:id", postHandler.GetPost)
				}

				projects := adminPrivate.Group("/projects")
				{
					projects.POST("", projectHandler.CreateProject)
					projects.GET("", projectHandler.ListProjects)
					projects.GET("/:id", projectHandler.GetProject)
					projects.PUT("/:id", projectHandler.UpdateProject)
					projects.DELETE("/:id", projectHandler.DeleteProject)
				}

				media := adminPrivate.Group("/media")
				{
					media.POST("/upload", mediaHandler.UploadMedia)
					media.PUT("/:id", mediaHandler.UpdateMedia)
					media.DELETE("/:id", mediaHandler.DeleteMedia)
				}

				hobbies := adminPrivate.Group("/hobbies")
				{
					hobbies.POST("", hobbyHandler.CreateHobbyItem)
					hobbies.GET("", hobbyHandler.ListHobbyItems)   // ?category=...
					hobbies.GET("/:id", hobbyHandler.GetHobbyItem) // ?category=...
					hobbies.PUT("/:id", hobbyHandler.UpdateHobbyItem)
					hobbies.DELETE("/:id", hobbyHandler.DeleteHobbyItem)
				}

				adminPrivate.GET("/search", searchHandler.SearchPrivate)
			}
		}

		public := api.Group("/")
		{
			public.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "UP"}) })

			public.GET("/posts", postHandler.ListPublicPosts)
			public.GET("/posts/:slug", postHandler.GetPublicPost)

			public.GET("/projects", projectHandler.ListPublicProjects)
			public.GET("/projects/:slug", projectHandler.GetPublicProject)

			public.GET("/media", mediaHandler.ListPublicMedia)

			public.GET("/hobbies", hobbyHandler.ListPublicHobbyItems) // ?category=...

			public.GET("/search", searchHandler.SearchPublic)

			public.GET("/metrics", gin.WrapH(promhttp.Handler()))
		}
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	appLogger.Info(fmt.Sprintf("Server running on port %s", cfg.App.Port))
	if err := router.Run(":" + cfg.App.Port); err != nil {
		appLogger.Fatal("Cannot run server", err)
	}
}
