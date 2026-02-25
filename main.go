package main

import (
	"api-arveshop-go/config"
	"api-arveshop-go/jobs"
	"api-arveshop-go/models"
	"api-arveshop-go/routes"
	"api-arveshop-go/utils"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	r := gin.Default()

	// CORS config
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Database
	config.ConnectDB()
	config.DB.AutoMigrate(
		&models.User{},
		&models.Whatsapp{},
		&models.Transaction{},
		&models.Service{},
		&models.Product{},
		&models.ProductPasca{},
		&models.PaymentMethod{},
		&models.Category{},
		&models.ProfilAplikasi{},		
	)

	// Redis untuk Asynq
	config.InitRedis()

	// Cloudinary
	if err := utils.InitCloudinary(); err != nil {
		log.Fatal("Failed to initialize Cloudinary: ", err)
	}

	// 🟢 JALANKAN WORKER DI GOROUTINE
	go startWorker()

	// Routes
	routes.SetupRoutes(r)

	// Jalankan server
	log.Println("🚀 Server running on 0.0.0.0:8080")
	r.Run(":8080")
}

func startWorker() {
	// 🔴 PERBAIKAN 1: Set default Redis address
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379" // default asynq
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // 🔴 PERBAIKAN 2: Tambahkan DB jika perlu
	}

	// 🔴 PERBAIKAN 3: Cek koneksi Redis dulu
	client := asynq.NewClient(redisOpt)
	defer client.Close()
	
	// if _, err := client.Ping(); err != nil {
	// 	log.Printf("⚠️ Redis not available: %v", err)
	// 	return
	// }
	log.Println("✅ Redis connected for worker")

	// Server config
	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		// 🔴 PERBAIKAN 4: Backoff yang benar sesuai dokumentasi
		RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
    backoff := []time.Duration{1, 3, 5, 10, 15} // dalam menit
    // Asynq: n = 1 untuk retry pertama, n = 2 untuk retry kedua, dst.
    // Jika n = 0 (bukan retry), kembalikan 0.
    if n == 0 {
        return 0 * time.Second
    }
    if n <= len(backoff) {
        return backoff[n-1] * time.Minute
    }
    return 15 * time.Minute // max 15 menit
},
	})

	// Processor
	processor := jobs.NewDigiflazzProcessor(
		config.DB,
		config.RDB,
		jobs.DigiflazzConfig{
			Username: os.Getenv("DIGIFLAZZ_USERNAME"),
			ProdKey:  os.Getenv("DIGIFLAZZ_PROD_KEY"),
			BaseURL:  "https://api.digiflazz.com",
		},
	)

	// Router
	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TaskDigiflazzTopup, processor.ProcessTask)

	// 🔴 PERBAIKAN 5: Tambahkan log
	log.Println("👷 Worker started, waiting for jobs...")

	// Jalankan server (blocking)
	if err := srv.Run(mux); err != nil {
		log.Printf("❌ Worker error: %v", err)
	}
}