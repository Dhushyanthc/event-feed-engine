package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Dhushyanthc/event-feed-engine/internal/config"
	"github.com/Dhushyanthc/event-feed-engine/internal/database"
	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
	// "github.com/Dhushyanthc/event-feed-engine/internal/database"
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

	// example user 
	user := models.User{
		Name: "Dhanush",
		Email: "abc@gmail.com",
		PasswordHash: "hashed_password",
	}

	// Create user in the database
	err = userRepo.CreateUser(context.Background(), &user)
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}
	fmt.Println("User created successfully", user.Id)


	
}