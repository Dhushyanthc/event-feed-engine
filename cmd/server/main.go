package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	projectdb "github.com/Dhushyanthc/event-feed-engine/db"
	"github.com/Dhushyanthc/event-feed-engine/internal/config"
	"github.com/Dhushyanthc/event-feed-engine/internal/database"
	"github.com/Dhushyanthc/event-feed-engine/internal/feed"
	"github.com/Dhushyanthc/event-feed-engine/internal/handlers"
	"github.com/Dhushyanthc/event-feed-engine/internal/logger"
	"github.com/Dhushyanthc/event-feed-engine/internal/middleware"
	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
	"go.uber.org/zap"
)

func main() {

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
	db, err := database.NewPostgres(cfg.DatabaseUrl)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	zapLogger.Info("Database connected")

	if err := projectdb.RunMigrations(context.Background(), db); err != nil {
		zapLogger.Fatal("failed to run migrations", zap.Error(err))
	}
	zapLogger.Info("Database migrations applied")

	//Initialize redis
	redisClient, err := database.NewRedisClient(cfg.RedisUrl)
	if err != nil {
		zapLogger.Fatal("failed to connect to redis", zap.Error(err))
	}
	zapLogger.Info("redis connected")

	//Initialize the User repository
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	followRepo := repository.NewFollowRepository(db)
	feedRepo := repository.NewFeedRepository(db, redisClient)
	eventRepo := repository.NewEventRepository(db)

	rateLimiter := middleware.RateLimitMiddleware(redisClient, 10, time.Minute)

	// Initialize feed fanout workers
	fanoutSvc := feed.NewFeedFanout(followRepo, feedRepo)
	dlq := feed.NewDeadLetterQueue(100)
	dlqWorker := feed.NewDLQWorker(dlq, zapLogger, fanoutSvc)
	dbWorker := feed.NewDBWorker(eventRepo, fanoutSvc, zapLogger)

	// start workers
	go dbWorker.Start(context.Background())
	go dlqWorker.Start(context.Background())

	//Initialize User handler
	userHandler := handlers.NewUserHandler(userRepo)
	loginHandler := handlers.NewLoginHandler(userRepo)
	postHandler := handlers.NewPostHandler(postRepo, eventRepo)
	followHandler := handlers.NewFollowHandler(followRepo)
	feedHandler := handlers.NewFeedHandler(feedRepo, postRepo)

	//start the http server
	mux := http.NewServeMux()

	//Register route
	mux.HandleFunc("/users", middleware.CORSMiddleware(
    middleware.LoggingMiddleware(zapLogger, userHandler.CreateUser),
))

mux.HandleFunc("/login", middleware.CORSMiddleware(
    rateLimiter(
        middleware.LoggingMiddleware(zapLogger,
            loginHandler.Login,
        ),
    ),
))

mux.HandleFunc("/posts", middleware.CORSMiddleware(
    rateLimiter(
        middleware.LoggingMiddleware(zapLogger,
            middleware.AuthMiddleware(postHandler.CreatePost),
        ),
    ),
))

mux.HandleFunc("/feed", middleware.CORSMiddleware(
    middleware.LoggingMiddleware(zapLogger, middleware.AuthMiddleware(feedHandler.GetFeed)),
))

mux.HandleFunc("/follow", middleware.CORSMiddleware(
    rateLimiter(
        middleware.LoggingMiddleware(zapLogger,
            middleware.AuthMiddleware(followHandler.FollowUser),
        ),
    ),
))

mux.HandleFunc("/unfollow", middleware.CORSMiddleware(
    middleware.LoggingMiddleware(zapLogger, middleware.AuthMiddleware(followHandler.UnfollowUser)),
))

mux.HandleFunc("/followers", middleware.CORSMiddleware(
    middleware.LoggingMiddleware(zapLogger, middleware.AuthMiddleware(followHandler.GetFollowers)),
))

mux.HandleFunc("/following", middleware.CORSMiddleware(
    middleware.LoggingMiddleware(zapLogger, middleware.AuthMiddleware(followHandler.GetFollowing)),
))

	//Start the server
	zapLogger.Info("Starting server", zap.String("port", cfg.Port))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("failed to start the server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	zapLogger.Info("shutdown signal recieved")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("server exited properly")
}
