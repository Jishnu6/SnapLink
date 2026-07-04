package api

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"url-shortner/internal/database" // Update to match your module path
	"url-shortner/internal/models"
	"url-shortner/internal/utils"
)

// DTOs matching your React UI payloads
type OTPRequest struct {
	Email string `json:"email"`
}

type OTPVerify struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// 1. Request OTP Handler
func RequestOTP(c *fiber.Ctx) error {
	req := new(OTPRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	// Generate a secure 6-digit OTP
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	redisKey := fmt.Sprintf("otp:auth:%s", req.Email)

	// Save to Redis with a 5-minute Time-To-Live (TTL)
	ctx := context.Background()
	err := database.RedisClient.Set(ctx, redisKey, otp, 5*time.Minute).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not process request"})
	}

	// Execute the SMTP network call in a background goroutine!
	// This ensures your React UI gets an instant response while the email sends.
	go func(email, code string) {
		if err := utils.SendOTPEmail(email, code); err != nil {
			fmt.Println("Failed to send OTP Email:", err)
		}
	}(req.Email, otp)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If the email exists, an OTP was sent.", // Vague message for security
	})
}

// 2. Verify OTP Handler
func VerifyOTP(c *fiber.Ctx) error {
	req := new(OTPVerify)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	ctx := context.Background()
	redisKey := fmt.Sprintf("otp:login:%s", req.Email)

	// Verify against Redis
	storedOTP, err := database.RedisClient.Get(ctx, redisKey).Result()
	fmt.Printf("the otp is : %s", storedOTP)
	if err != nil || storedOTP != req.OTP {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired OTP"})
	}

	// OTP is verified - Delete it from Redis immediately so it cannot be reused
	database.RedisClient.Del(ctx, redisKey)

	// MongoDB: Upsert the User (Log them in, or create their account if it's their first time)
	collection := database.MongoClient.Database("url_shortener").Collection("users")
	filter := bson.M{"email": req.Email}
	update := bson.M{"$setOnInsert": bson.M{
		"email":      req.Email,
		"created_at": time.Now(),
	}}
	opts := options.Update().SetUpsert(true)

	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Fetch the generated/existing User ID for the token payload
	var user models.User
	collection.FindOne(ctx, filter).Decode(&user)

	// Generate the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token expires in 3 days
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	signedToken, _ := token.SignedString([]byte(jwtSecret))

	// Return the token to the React frontend to store in localStorage
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Authentication successful",
		"token":   signedToken,
	})
}

func RequestLoginOTP(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userCollection := database.MongoClient.Database("url_shortener").Collection("users")

	var existingUser models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&existingUser)

	// --- CORRECTION HERE ---
	// If we are logging in, we expect the user to exist.
	if err == mongo.ErrNoDocuments {
		// Stop the process: User is not registered
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Account not found. Please sign up first.",
		})
	} else if err != nil {
		// Real database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	// If err == nil, the user exists. Proceed to OTP generation.
	otp, err := utils.GenerateOTP()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate OTP",
		})
	}

	// Cache the OTP in Redis for 5 minutes
	redisKey := "otp:login:" + req.Email
	cacheErr := database.RedisClient.Set(context.Background(), redisKey, otp, 5*time.Minute).Err()
	if cacheErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store OTP",
		})
	}

	go utils.SendOTPEmail(req.Email, otp)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login code sent successfully",
	})
}

func SignupRequest(c *fiber.Ctx) error {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 1. CHECK IF USER ALREADY EXISTS
	// Get your users collection reference
	userCollection := database.MongoClient.Database("url_shortener").Collection("users")

	var existingUser models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&existingUser)

	if err == nil {
		// err == nil means a document WAS successfully found. The user exists!
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "An account with this email already exists. Please log in.",
		})
	} else if err != mongo.ErrNoDocuments {
		// If it's an error OTHER than "no documents found", the database crashed or timed out
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error while verifying email",
		})
	}

	// 2. USER DOES NOT EXIST -> PROCEED WITH OTP GENERATION
	// (err == mongo.ErrNoDocuments implies it is safe to proceed)

	otp, err := utils.GenerateOTP() // Your 6-digit random number generator

	// 3. Cache the OTP in Redis for 5 minutes
	redisKey := "otp:login:" + req.Email
	cacheErr := database.RedisClient.Set(context.Background(), redisKey, otp, 5*time.Minute).Err()
	if cacheErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create secure session",
		})
	}

	// 4. Send the OTP via Email in a background Goroutine so the API doesn't block
	go utils.SendOTPEmail(req.Email, otp)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Verification code sent successfully",
	})
}

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

func Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	ctx := context.Background()

	// Add the token to the Redis blacklist.
	// We set the TTL to 72 hours to match the token's maximum lifespan.
	// After 72 hours, the token expires naturally anyway, so Redis can delete the blacklist entry.
	err := database.RedisClient.Set(ctx, "blacklist:"+tokenString, "revoked", 72*time.Hour).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to log out completely"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}
