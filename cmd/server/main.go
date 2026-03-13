package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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


	server := &http.Server{
		Addr: ":"+cfg.Port,
		Handler: mux,
	}

	go func(){
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed{
		zapLogger.Fatal("failed to start the server", zap.Error(err))
	}
}()

quit := make(chan os.Signal, 1)

signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

<-quit

zapLogger.Info("shutdown signal recieved")

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil{
	zapLogger.Fatal("server forced to shutdown", zap.Error(err))
}

zapLogger.Info("server exited properly")
}