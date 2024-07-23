package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
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
		time.Sleep(3000 * time.Millisecond)
		return c.String(http.StatusOK, "Request received and queued")
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

	// Example: Normally, you would do additional processing here

	fmt.Println("Processed request:", r.URL.Path)
}
