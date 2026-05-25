package config

import (
	"context"
	"os"
	"strconv"
	"github.com/redis/go-redis/v9"
)

func ConnectRedis() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("Gagal terkoneksi ke Redis Server: " + err.Error())
	}

	return rdb
}