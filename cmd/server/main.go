package main

import (
	
	"fmt"
	"log"
	"net/http"

	"github.com/Dhushyanthc/event-feed-engine/internal/config"
	"github.com/Dhushyanthc/event-feed-engine/internal/database"
	"github.com/Dhushyanthc/event-feed-engine/internal/handlers"
	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
)

func main(){
	
	//Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Configuration Loaded and App envirnment:", cfg.AppEnv)


	//Initailize the database connection
	db,err := database.NewPostgres(cfg.DatabaseUrl)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Database is connected successfully")

	//Initialize the User repository 
	userRepo := repository.NewUserRepository(db)

	//Initialize User handler
	userHandler := handlers.NewUserHandler(userRepo)
	loginHandler := handlers.NewLoginHandler(userRepo)

	//start the http server
	mux:= http.NewServeMux()

	//Register route
	mux.HandleFunc("/users", userHandler.CreateUser)
	mux.HandleFunc("/login", loginHandler.Login)

	//Start the server
	log.Printf("Starting server on port %s\n", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, mux)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}


	
	
}