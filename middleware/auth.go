package middleware

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"url-shortner/internal/database" // Update to match your module path
)

func RequireAuth(c *fiber.Ctx) error {
	// 1. Extract the Token from the Authorization Header
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid token"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// 2. Security Check: Is this token in the Redis Blacklist? (Logged out)
	ctx := context.Background()
	isBlacklisted, _ := database.RedisClient.Get(ctx, "blacklist:"+tokenString).Result()
	if isBlacklisted != "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Session revoked. Please log in again."})
	}

	// 3. Parse and Mathematically Validate the JWT Signature
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect (HMAC)
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	// 4. Extract Claims (Payload) and pass the User ID downstream
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token payload"})
	}

	// Save the user_id into Fiber's local context so your route handlers can use it
	c.Locals("user_id", claims["user_id"])

	// Token is valid! Proceed to the requested route
	return c.Next()
}