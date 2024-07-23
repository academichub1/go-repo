package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	jwtKey        = []byte("your_secret_key")
	maxWorkers    = 26000 // Number of worker goroutines
	maxQueue      = 28000 // Size of request queue
	refreshTokens = map[string]string{}
	redisClient   *redis.Client
	ctx           = context.Background()
)

func main() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
		DB:   0,                // Use default DB
	})

	// Create a buffered channel to queue incoming requests with capacity maxQueue
	queue := make(chan *http.Request, maxQueue)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS Middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // You can restrict this to specific origins if needed
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// Route handlers
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Health OK")
	})

	e.GET("/", func(c echo.Context) error {
		select {
		case queue <- c.Request():
			fmt.Println("Request queued successfully")
		default:
			return c.String(http.StatusServiceUnavailable, "Queue full. Please try again later.")
		}

		time.Sleep(3000 * time.Millisecond)

		return c.String(http.StatusOK, "Request received and queued")
	})

	e.GET("/v2", func(c echo.Context) error {

		// Generate random text or number
		key := fmt.Sprintf("example_value_%d", rand.Intn(1000))
		// SET operation
		err := redisClient.Set(ctx, key, "example_value", 0).Err()
		if err != nil {
			fmt.Println("Failed to set key in Redis:", err)
			return c.String(http.StatusInternalServerError, "Failed to set key in Redis")
		}

		// GET operation
		value, err := redisClient.Get(ctx, key).Result()
		if err != nil {
			fmt.Println("Failed to get key from Redis:", err)
			return c.String(http.StatusInternalServerError, "Failed to get key from Redis")
		}

		redisClient.Del(ctx, key).Result()

		time.Sleep(3000 * time.Millisecond)
		fmt.Println("Processed request:", c.Request().URL.Path, "with Redis value:", value)

		return c.String(http.StatusOK, "Request received and queued with Redis value: "+value)
	})

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(queue, &wg)
	}

	// Start HTTP server
	fmt.Println("Server is listening on port 8080")
	e.Logger.Fatal(e.Start(":8080"))

	// Wait for all workers to finish
	wg.Wait()
}

func worker(queue chan *http.Request, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case req := <-queue:
			// Process the request
			processRequest(req)
		}
	}
}

func processRequest(r *http.Request) {
	// Simulate processing time for demonstration
	time.Sleep(3000 * time.Millisecond)

	// Example Redis operations
	key := "example_key"

	// SET operation
	err := redisClient.Set(ctx, key, "example_value", 0).Err()
	if err != nil {
		fmt.Println("Failed to set key in Redis:", err)
		return
	}

	// GET operation
	value, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		fmt.Println("Failed to get key from Redis:", err)
		return
	}

	fmt.Println("Processed request:", r.URL.Path, "with Redis value:", value)
}
