// config/redis.go
package config

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func InitRedis() {
    RDB = redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_ADDR"),     // "localhost:6379"
        Password: os.Getenv("REDIS_PASSWORD"), // kosong kalau tidak ada password
        DB:       0,
    })

    // Test koneksi
    ctx := context.Background()
    _, err := RDB.Ping(ctx).Result()
    if err != nil {
        log.Fatalf("Gagal koneksi ke Redis: %v", err)
    }

    log.Println("Redis connected")
}