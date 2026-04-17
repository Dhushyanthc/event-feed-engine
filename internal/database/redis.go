package database

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(redisURL string) (*redis.Client, error){

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil,err
	}

	client := redis.NewClient(opts)

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil{
		return nil, err
	}

	return client, nil
}