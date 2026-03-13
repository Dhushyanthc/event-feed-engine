package main

import (
	"log"
	"net/http"

	"github.com/Dhushyanthc/event-feed-engine/internal/config"
	"github.com/Dhushyanthc/event-feed-engine/internal/database"
	"github.com/Dhushyanthc/event-feed-engine/internal/handlers"
	"github.com/Dhushyanthc/event-feed-engine/internal/logger"
	"github.com/Dhushyanthc/event-feed-engine/internal/middleware"
	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
	"go.uber.org/zap"
)

func main(){
	
	//Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	


	zapLogger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	//Initailize the database connection
	db,err := database.NewPostgres(cfg.DatabaseUrl)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	zapLogger.Info("Database connected")

	//Initialize the User repository 
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	followRepo := repository.NewFollowRepository(db)
	feedRepo := repository.NewFeedRepository(db)


	//Initialize User handler
	userHandler := handlers.NewUserHandler(userRepo)
	loginHandler := handlers.NewLoginHandler(userRepo)
	postHandler := handlers.NewPostHandler(postRepo)
	followHandler := handlers.NewFollowHandler(followRepo)
	feedHandler := handlers.NewFeedHandler(feedRepo, postRepo)

	//start the http server
	mux:= http.NewServeMux()

	//Register route
	mux.HandleFunc("/users", middleware.LoggingMiddleware(zapLogger,userHandler.CreateUser))

	mux.HandleFunc("/login", middleware.LoggingMiddleware(zapLogger,loginHandler.Login))

	mux.HandleFunc("/posts", middleware.LoggingMiddleware(zapLogger,middleware.AuthMiddleware(postHandler.CreatePost)))

	mux.HandleFunc("/feed", middleware.LoggingMiddleware(zapLogger,middleware.AuthMiddleware(feedHandler.GetFeed)))

	mux.HandleFunc("/follow", middleware.LoggingMiddleware(zapLogger,middleware.AuthMiddleware(followHandler.FollowUser)))

	mux.HandleFunc("/unfollow",middleware.LoggingMiddleware(zapLogger, middleware.AuthMiddleware(followHandler.UnfollowUser)))

	mux.HandleFunc("/followers",middleware.LoggingMiddleware(zapLogger, middleware.AuthMiddleware(followHandler.GetFollowers)))
	
	mux.HandleFunc("/following", middleware.LoggingMiddleware(zapLogger, middleware.AuthMiddleware(followHandler.GetFollowing)))

	//Start the server
	zapLogger.Info("Starting server", zap.String("port", cfg.Port))
	err = http.ListenAndServe(":"+cfg.Port, mux)
	if err != nil {
		zapLogger.Fatal("Failed to start server", zap.Error(err))
	}

}