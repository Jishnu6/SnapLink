package middleware

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// IPBlacklist checks incoming IPs against a Redis blocklist
func IPBlacklist(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract the user's IP.
		// Note: If you are behind a proxy (Cloudflare/Nginx), make sure Fiber
		// is configured to trust proxy headers, or manually check X-Forwarded-For.
		clientIP := c.IP()

		// 2. Check if the IP exists in the Redis Set "blacklist:ips"
		isBlacklisted, err := redisClient.SIsMember(context.Background(), "blacklist:ips", clientIP).Result()

		if err != nil {
			// Fail-open strategy: If Redis temporarily fails, allow the request
			// rather than taking down your entire service.
			log.Printf("Redis blacklist check failed for IP %s: %v", clientIP, err)
			return c.Next()
		}

		// 3. Block the request if the IP is blacklisted
		if isBlacklisted {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Access denied. Your IP has been restricted.",
			})
		}

		// 4. IP is clean, proceed to the next middleware/handler
		return c.Next()
	}
}
