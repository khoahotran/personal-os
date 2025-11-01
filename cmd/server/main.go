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
	postUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
	profileUC "github.com/khoahotran/personal-os/internal/application/usecase/profile"
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
			}
		}

		public := api.Group("/")
		{
			public.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "UP"}) })
			public.GET("/posts", postHandler.ListPublicPosts)
			public.GET("/posts/:slug", postHandler.GetPublicPost)

		}
	}

	log.Printf("Server running on port %s", cfg.App.Port)
	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Cannot run server: %v", err)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
