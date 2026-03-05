package main 

import (
	"context"
	"fmt"
	"log"

	"github.com/Dhushyanthc/event-feed-engine/internal/database"
	"github.com/joho/godotenv"
)

func main(){
	ctx := context.Background()

	err := godotenv.Load()

	dbPool, err := database.NewPostgresPool(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer dbPool.Close()

	fmt.Println("Successfully connected to the database")

	var result int
	err = dbPool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil{
		log.Fatalf("query failed: %v", err)
	}

	fmt.Println("Database query succesful", result)
}