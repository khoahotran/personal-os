package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/khoahotran/personal-os/adapters/event"
	httpAdapter "github.com/khoahotran/personal-os/adapters/http"
	"github.com/khoahotran/personal-os/adapters/media_storage"
	"github.com/khoahotran/personal-os/adapters/persistence"
	authUC "github.com/khoahotran/personal-os/internal/application/usecase/auth"
	mediaUC "github.com/khoahotran/personal-os/internal/application/usecase/media"
	postUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
	profileUC "github.com/khoahotran/personal-os/internal/application/usecase/profile"
	projectUC "github.com/khoahotran/personal-os/internal/application/usecase/project"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/pkg/auth"
)

func main() {
	fmt.Println("Start Personal OS API Server...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: cannot load config: %v", err)
	}

	// Initialize dependencies
	dbPool, err := persistence.NewPostgresPool(cfg)
	if err != nil {
		log.Fatalf("FATAL: cannot connect Postgres: %v", err)
	}
	defer dbPool.Close()

	redisClient, err := persistence.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("FATAL: cannot connect Redis: %v", err)
	}
	defer redisClient.Close()

	kafkaClient, err := event.NewKafkaProducerClient(cfg)
	if err != nil {
		log.Fatalf("FATAL: cannot init Kafka: %v", err)
	}
	defer kafkaClient.Close()

	// Repositories
	userRepo := persistence.NewPostgresUserRepo(dbPool)
	profileRepo := persistence.NewPostgresProfileRepo(dbPool)
	postRepo := persistence.NewPostgresPostRepo(dbPool)
	tagRepo := persistence.NewPostgresTagRepo(dbPool)
	projectRepo := persistence.NewPostgresProjectRepo(dbPool)
	mediaRepo := persistence.NewPostgresMediaRepo(dbPool)

	// Services
	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenLifespan)
	uploader, err := media_storage.NewCloudinaryAdapter(cfg)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize uploader: %v", err)
	}

	// Use Cases
	loginUseCase := authUC.NewLoginUseCase(userRepo, jwtSvc)
	profileUseCase := profileUC.NewProfileUseCase(profileRepo)

	createPostUseCase := postUC.NewCreatePostUseCase(postRepo, tagRepo, kafkaClient, uploader)
	listPostsUseCase := postUC.NewListPostsUseCase(postRepo, tagRepo)
	listPublicPostsUseCase := postUC.NewListPublicPostsUseCase(postRepo, tagRepo)
	updatePostUseCase := postUC.NewUpdatePostUseCase(postRepo, tagRepo, kafkaClient)
	deletePostUseCase := postUC.NewDeletePostUseCase(postRepo, tagRepo, kafkaClient)
	getPostUseCase := postUC.NewGetPostUseCase(postRepo, tagRepo)
	getPublicPostUseCase := postUC.NewGetPublicPostUseCase(postRepo, tagRepo)

	createProjectUseCase := projectUC.NewCreateProjectUseCase(projectRepo, tagRepo)
	listProjectsUseCase := projectUC.NewListProjectsUseCase(projectRepo)
	listPublicProjectsUseCase := projectUC.NewListPublicProjectsUseCase(projectRepo)
	getProjectUseCase := projectUC.NewGetProjectUseCase(projectRepo, tagRepo)
	getPublicProjectUseCase := projectUC.NewGetPublicProjectUseCase(projectRepo, tagRepo)
	updateProjectUseCase := projectUC.NewUpdateProjectUseCase(projectRepo, tagRepo)
	deleteProjectUseCase := projectUC.NewDeleteProjectUseCase(projectRepo, tagRepo)

	uploadMediaUseCase := mediaUC.NewUploadMediaUseCase(mediaRepo, uploader, kafkaClient)
	listPublicMediaUseCase := mediaUC.NewListPublicMediaUseCase(mediaRepo)
	updateMediaUseCase := mediaUC.NewUpdateMediaUseCase(mediaRepo)
	deleteMediaUseCase := mediaUC.NewDeleteMediaUseCase(mediaRepo, uploader)

	// HTTP Handlers
	authHandler := httpAdapter.NewAuthHandler(loginUseCase)
	profileHandler := httpAdapter.NewProfileHandler(profileUseCase)
	postHandler := httpAdapter.NewPostHandler(
		createPostUseCase,
		listPostsUseCase,
		listPublicPostsUseCase,
		updatePostUseCase,
		deletePostUseCase,
		getPostUseCase,
		getPublicPostUseCase,
	)

	projectHandler := httpAdapter.NewProjectHandler(
		createProjectUseCase, listProjectsUseCase, listPublicProjectsUseCase,
		getProjectUseCase, getPublicProjectUseCase, updateProjectUseCase, deleteProjectUseCase,
	)

	mediaHandler := httpAdapter.NewMediaHandler(
		uploadMediaUseCase,
		listPublicMediaUseCase,
		updateMediaUseCase,
		deleteMediaUseCase,
	)

	// Middleware
	authMiddleware := httpAdapter.AuthMiddleware(jwtSvc)

	// Setup Gin router
	router := gin.Default()

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
		}
	}

	log.Printf("Server running on port %s", cfg.App.Port)
	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Cannot run server: %v", err)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
