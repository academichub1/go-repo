package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.StandardClaims
}

var (
	jwtKey        = []byte("your_secret_key")
	maxWorkers    = 25000 // Number of worker goroutines
	maxQueue      = 26000 // Size of request queue
	refreshTokens = map[string]string{}
)

func main() {
	// Create a buffered channel to queue incoming requests with capacity maxQueue
	queue := make(chan *http.Request, maxQueue)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

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

		time.Sleep(3 * time.Second) // Simulate some processing time

		return c.String(http.StatusOK, "Request received and queued")
	})

	e.POST("/login", LoginHandler)
	e.POST("/refresh", RefreshTokenHandler)

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

// LoginHandler handles user login and issues JWT tokens
func LoginHandler(c echo.Context) error {
	var creds Credentials
	if err := c.Bind(&creds); err != nil {
		return c.String(http.StatusBadRequest, "Invalid credentials")
	}

	// Example: Simulated database check
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	storedPassword := hashedPassword // Simulated stored hashed password

	if err := bcrypt.CompareHashAndPassword(storedPassword, []byte(creds.Password)); err != nil {
		return c.String(http.StatusUnauthorized, "Invalid credentials")
	}

	const letters = "abcdefghijklmnopqrstuvwxyz"
	ran := make([]byte, 5)
	for i := range ran {
		ran[i] = letters[rand.Intn(len(letters))]
	}

	result := string(ran)

	// Generate JWT
	expirationTime := time.Now().Add(50000 * time.Second)
	claims := &Claims{
		Email: creds.Username,
		Name:  "test " + result,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate token")
	}

	// Generate refresh token
	refreshExpirationTime := time.Now().Add(24 * time.Hour) // Example: Refresh token lasts 24 hours
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["email"] = creds.Username
	refreshClaims["name"] = "test " + result
	refreshClaims["exp"] = refreshExpirationTime.Unix()

	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate refresh token")
	}

	// Store refresh token (in a real application, you would store this securely)
	refreshTokens[refreshTokenString] = creds.Username

	// Return tokens to the client
	response := map[string]string{
		"access_token":    tokenString,
		"refresh_token":   refreshTokenString,
		"expires_at":      expirationTime.Format(time.RFC3339),
		"refresh_expires": refreshExpirationTime.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, response)
}

// RefreshTokenHandler handles refresh token requests
func RefreshTokenHandler(c echo.Context) error {
	refreshToken := c.FormValue("refresh_token")
	if refreshToken == "" {
		return c.String(http.StatusBadRequest, "Refresh token missing")
	}

	// Verify the refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.String(http.StatusUnauthorized, "Invalid refresh token")
	}

	// Extract claims from the refresh token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid refresh token claims")
	}

	// Check if the refresh token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.String(http.StatusUnauthorized, "Refresh token expired")
	}

	// Generate a new access token
	expirationTime := time.Now().Add(50000 * time.Second)
	newClaims := &Claims{
		Email: claims["email"].(string),
		Name:  claims["name"].(string),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := newToken.SignedString(jwtKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate token")
	}

	// Generate a new refresh token (optional: refresh the refresh token)
	refreshExpirationTime := time.Now().Add(24 * time.Hour) // Example: Refresh token lasts 24 hours
	newRefreshToken := jwt.New(jwt.SigningMethodHS256)
	newRefreshClaims := newRefreshToken.Claims.(jwt.MapClaims)
	newRefreshClaims["email"] = claims["email"]
	newRefreshClaims["name"] = claims["name"]
	newRefreshClaims["exp"] = refreshExpirationTime.Unix()

	newRefreshTokenString, err := newRefreshToken.SignedString(jwtKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate refresh token")
	}

	// Update refresh token in the map (optional: if refreshing the refresh token)
	refreshTokens[newRefreshTokenString] = claims["email"].(string)

	// Return the new tokens to the client
	response := map[string]string{
		"access_token":    newTokenString,
		"refresh_token":   newRefreshTokenString,
		"expires_at":      expirationTime.Format(time.RFC3339),
		"refresh_expires": refreshExpirationTime.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, response)
}
