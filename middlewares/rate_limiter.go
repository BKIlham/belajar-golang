package middlewares

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/redis/go-redis/v9"
)

func NewRateLimiter(rdb *redis.Client) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,

		Storage: &redisStorage{rdb: rdb},

		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Too many requests. Please slow down and try again later.",
			})
		},
	})
}

// Struct pembantu untuk menjembatani Fiber Limiter dengan Go-Redis v9
type redisStorage struct {
	rdb *redis.Client
}

func (s *redisStorage) Get(key string) ([]byte, error) {
	val, err := s.rdb.Get(context.Background(), "limiter:"+key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (s *redisStorage) Set(key string, val []byte, exp time.Duration) error {
	return s.rdb.Set(context.Background(), "limiter:"+key, val, exp).Err()
}

func (s *redisStorage) Delete(key string) error {
	return s.rdb.Del(context.Background(), "limiter:"+key).Err()
}

func (s *redisStorage) Reset() error {
	return s.rdb.FlushDB(context.Background()).Err()
}

// 🛠️ TAMBAHKAN METHOD CLOSE INI AGAR MEMENUHI KONTRAK INTERFACE fiber.Storage
func (s *redisStorage) Close() error {
	// Karena koneksi Redis utama kita diatur dan ditutup di main.go/config,
	// di sini kita cukup return nil saja.
	return nil
}