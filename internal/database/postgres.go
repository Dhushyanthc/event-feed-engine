package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)


func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, os.ErrClosed
	}

	return pool, nil
}