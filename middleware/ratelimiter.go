package middleware

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage/redis/v3"
)

// NewRateLimiter setups distributed rate limiting tiers backed by Redis Cloud
func NewRateLimiter() (fiber.Handler, fiber.Handler) {
	// Pull your Redis Cloud URL from your environment variables
	// It should look like: redis://default:your_password@redis-13540.ec2.cloud.redislabs.com:13540
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Fallback to localhost for local testing safety
		redisURL = "redis://127.0.0.1:6379"
	}

	// Initialize shared Redis storage configuration using the Cloud URL
	store := redis.New(redis.Config{
		Host:     "redis-11277.c44.us-east-1-2.ec2.cloud.redislabs.com",
		Port:     11277,
		Username: "default",
		Password: "8<V:O0%es_U.b52^wS?%n",
	})

	// Tier 1: Loose Limiter for Redirect Endpoint (e.g., /:shortCode)
	RedirectLimiter := limiter.New(limiter.Config{
		Max:        100,             // 100 requests
		Expiration: 1 * time.Minute, // per minute per IP
		Storage:    store,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "limiter:redirect:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many redirect requests. Please slow down.",
			})
		},
	})

	// Tier 2: Strict Limiter for Resource-Intensive Paths (e.g., /api/v1/url)
	StrictLimiter := limiter.New(limiter.Config{
		Max:        10,              // 10 requests
		Expiration: 1 * time.Minute, // per minute per IP
		Storage:    store,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "limiter:strict:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "API limit exceeded. URL creation is restricted to 10 requests per minute.",
			})
		},
	})

	return RedirectLimiter, StrictLimiter
}
