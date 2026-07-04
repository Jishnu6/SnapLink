package main

import (
	"log"
	"os"
	"time"

	"url-shortner/api"
	"url-shortner/internal/database"
	workers "url-shortner/internal/models/Workers"
	"url-shortner/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	// 2. Initialize Databases
	database.ConnectMongo()
	database.ConnectRedis()
	database.CreateTTLIndex()

	workers.StartAnalyticsWorker(1 * time.Minute)

	// 3. Setup Fiber App
	app := fiber.New(fiber.Config{
		AppName:     "Go URL Shortener v1.0",
		ProxyHeader: "X-Forwarded-For",
	})

	// Add standard middleware for logging requests
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		// Allow the origin of your Vite development server
		AllowOrigins: "http://localhost:5173",
		// Explicitly allow headers your React application sends
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		// Allow the HTTP verbs your API uses, plus OPTIONS for preflight checking
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
		// Allows browser to pass auth tokens/cookies if needed later
		AllowCredentials: true,
	}))

	app.Use(middleware.IPBlacklist(database.RedisClient))

	redirectLimiter, strictLimiter := middleware.NewRateLimiter()

	// =====================================================================
	// PUBLIC AUTH ROUTES
	// =====================================================================
	// We apply the strict rate limiter here to prevent malicious bots from
	// spamming users with OTP emails or brute-forcing OTP verifications.
	authGroup := app.Group("/api/auth", strictLimiter)
	authGroup.Post("/request", api.RequestOTP)
	authGroup.Post("/verify-otp", api.VerifyOTP)
	authGroup.Post("/signup-request", api.SignupRequest)
	authGroup.Post("/login-request", api.RequestLoginOTP)

	// =====================================================================
	// PROTECTED ROUTES (Requires valid JWT)
	// =====================================================================
	// By adding middleware.RequireAuth here, EVERY route under /api/v1
	// is automatically secured. We also keep the strict rate limiter.
	apiGroup := app.Group("/api/v1", strictLimiter, middleware.RequireAuth)
	apiGroup.Post("/shorten", api.ShortenURL)
	apiGroup.Get("/stats/:id", api.GetURLStats)

	// Secure Logout (Requires Auth context to know which token to blacklist in Redis)
	apiGroup.Post("/logout", api.Logout)

	// =====================================================================
	// PUBLIC REDIRECT ROUTE
	// =====================================================================
	// Tier 1 (Loose Limit): High throughput redirect line
	app.Get("/:id", redirectLimiter, api.Redirect)

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
