package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(databaseURL string) (*pgxpool.Pool, error){

	if databaseURL == ""{
		return nil, errors.New("database URL is missing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) 
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err !=nil {
		return nil, err
	}

	return pool, nil
}
