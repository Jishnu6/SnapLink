package api

import (
	"context"
	"log"
	"time"
	"url-shortner/internal/database"
	"url-shortner/internal/models"
	"url-shortner/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ShortenURL handles the POST request to create a short link
func ShortenURL(c *fiber.Ctx) error {
	body := new(models.ShortenRequest)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// 1. Validate the URL
	if !utils.IsValidURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL provided"})
	}

	if !utils.IsLiveLink(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Link is not live"})
	}

	userIDString, ok := c.Locals("user_id").(string)
	if !ok || userIDString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Missing user session"})
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	collection := database.GetCollection("urls")

	// 2. Generate short ID (using Base62 or custom alias)
	id := body.CustomAlias
	if id == "" {
		id = utils.GenerateShortID(6)
	} else {
		// Check if the requested custom alias is already inside the database
		var existing models.URL
		err := collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&existing)

		// If err is nil, it means MongoDB successfully FOUND a document with this ID
		if err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":     "Alias conflict",
				"message":   "The custom alias '" + id + "' is already taken. Please choose another one.",
				"field":     "customAlias",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		}
	}

	var expiryTime time.Time = time.Now().AddDate(0, 0, int(body.ExpiryDuration)) //time.Now().AddDate(0, 2, 0);

	urlEntry := models.URL{
		ID:          id,
		UserID:      userID,
		OriginalURL: body.URL,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiryTime,
	}

	// 3. Save to MongoDB
	_, insertErr := collection.InsertOne(context.TODO(), urlEntry)
	if insertErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not save URL"})
	}

	// 4. Cache in Redis (set expiration same as DB expiry)
	database.RedisClient.Set(context.TODO(), id, body.URL, time.Until(expiryTime))

	return c.Status(fiber.StatusCreated).JSON(models.ShortenResponse{
		OriginalURL: body.URL,
		ShortURL:    c.BaseURL() + "/" + id,
		ID:          id,
		ExpiresAt:   urlEntry.ExpiresAt,
	})
}

// Redirect handles the GET request to forward to the original URL
func Redirect(c *fiber.Ctx) error {
	id := c.Params("id")

	userIP := c.Get("X-Forwarded-For")
	if userIP == "" {
		userIP = c.IP()
	}

	userAgent := c.Get("User-Agent", "Unknown-Agent")
	referrer := c.Get("Referer", "Direct")

	// 1. Try fetching from Redis first (Cache Hit)
	val, err := database.RedisClient.Get(context.TODO(), id).Result()
	if err == nil {
		go collectAnalytics(id, userIP, userAgent, referrer)
		log.Println("REDIZ HIT: Fetching from Cache")
		return c.Redirect(val, fiber.StatusTemporaryRedirect)
	}

	// 2. If Redis fails, check MongoDB (Cache Miss)
	collection := database.GetCollection("urls")
	var result models.URL
	err = collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Inside Redirect function after finding 'result'
	if !result.ExpiresAt.IsZero() && time.Now().After(result.ExpiresAt) {
		return c.Status(410).JSON(fiber.Map{
			"error": "This link has expired",
		})
	}

	// 3. Update Redis cache for next time
	database.RedisClient.Set(context.TODO(), id, result.OriginalURL, time.Until(result.ExpiresAt))

	return c.Redirect(result.OriginalURL, fiber.StatusTemporaryRedirect)
}

// FetchIPDetails queries ip-api.com and returns the parsed JSON geolocation data

func collectAnalytics(id string, userIP string, userAgent string, referrer string) {
	ctx := context.TODO()

	// 1. Bypass local loopback for accurate tracking during script simulation
	if userIP == "127.0.0.1" || userIP == "::1" {
		userIP = "103.45.12.91"
	}

	locationDetails, locationError := utils.FetchIPDetails(userIP)
	var country string
	var city string

	if locationError != nil {
		country = "Unknown"
		city = "Unknown"
	} else {
		country = locationDetails.Country
		city = locationDetails.City
	}

	if referrer == "" {
		referrer = "Direct"
	}

	// 🔍 TRACKING PRINT: Let's see if the function actually executes
	log.Printf("📥 [collectAnalytics] Triggered for ID: %s | IP: %s | Country: %s | Ref: %s", id, userIP, country, referrer)

	// Verify Redis connection pointer before using it
	if database.RedisClient == nil {
		log.Println("❌ CRITICAL ERROR: database.RedisClient is a NIL pointer! Connection was never initialized.")
		return
	}

	fingerprint := utils.GenerateFingerprint(userIP, userAgent)

	// 2. Increment Hash Value
	err := database.RedisClient.HIncrBy(ctx, "analytics:clicks", id, 1).Err()
	if err != nil {
		log.Printf("❌ REDIS HINCRBY ERROR: %v", err)
	} else {
		log.Printf("✅ REDIS HINCRBY SUCCESS: Incremented total_clicks for %s", id)
	}

	payload := id + "::" + country + "::" + city + "::" + referrer + "::" + fingerprint + "::" + userAgent

	// 3. Push to List Queue
	err = database.RedisClient.LPush(ctx, "analytics:queue", payload).Err()
	if err != nil {
		log.Printf("❌ REDIS LPUSH ERROR: %v", err)
	} else {
		log.Printf("✅ REDIS LPUSH SUCCESS: Pushed data string to queue")
	}
}

func GetURLStats(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := context.TODO()

	// --- NEW: EXTRACT THE LOGGED-IN USER ID ---
	requesterIDString, ok := c.Locals("user_id").(string)
	if !ok || requesterIDString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Missing user session",
		})
	}
	// ------------------------------------------

	urlCollection := database.GetCollection("urls")

	// Update the anonymous struct to also fetch the user_id
	var urlDoc struct {
		UserID    primitive.ObjectID `bson:"user_id"` // The owner's ID
		ExpiresAt time.Time          `bson:"expires_at"`
	}

	err := urlCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&urlDoc)

	// If it's not found in the 'urls' collection, it's a completely fake/random ID
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "This short link ID does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error checking link validity"})
	}

	// --- NEW: THE SECURITY GATE ---
	// Does the person asking for stats actually own this link?
	if urlDoc.UserID.Hex() != requesterIDString {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access Denied: You do not own this link",
		})
	}
	// ------------------------------

	// --- STEP 2: EVALUATE IF THE LINK IS LIVE OR EXPIRED ---
	status := "Live"

	if !urlDoc.ExpiresAt.IsZero() && time.Now().After(urlDoc.ExpiresAt) {
		status = "Expired"
	}

	// 1. Get real-time total clicks from Redis (Always available, even if URL is expired)
	totalClicksStr, err := database.RedisClient.HGet(ctx, "analytics:clicks", id).Result()
	if err != nil && err != redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Redis error fetching stats"})
	}

	// If it doesn't exist in Redis yet, default to "0"
	if err == redis.Nil {
		totalClicksStr = "0"
	}

	// 2. Query MongoDB for granular location/referrer breakdowns
	collection := database.GetCollection("click_logs") // Make sure this matches your analytics collection name

	// Match all logs for this specific short ID
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "url_id", Value: id}}}}

	// Helper closure to dynamically handle aggregation fields without repeating code block layouts
	runAggregation := func(fieldName string) (map[string]int, error) {
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$" + fieldName},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}}

		cursor, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		var results []bson.M
		if err = cursor.All(ctx, &results); err != nil {
			return nil, err
		}

		resultMap := make(map[string]int)
		for _, res := range results {
			key, ok := res["_id"].(string)
			if !ok || key == "" {
				key = "Unknown" // Fallback label for unparsed strings
			}

			// Handle int32 conversions from MongoDB sums smoothly
			if count, countOk := res["count"].(int32); countOk {
				resultMap[key] = int(count)
			}
		}
		return resultMap, nil
	}

	// Run aggregations across all metric fields
	countriesMap, err := runAggregation("country")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate country stats"})
	}

	referrersMap, err := runAggregation("referrer")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate referrer stats"})
	}

	devicesMap, err := runAggregation("device_type")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate device stats"})
	}

	osMap, err := runAggregation("os")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate operating system stats"})
	}

	browsersMap, err := runAggregation("browser")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate browser stats"})
	}

	filter := bson.M{"url_id": id, "is_unique": true}

	// Efficiently count only documents marked as unique
	count, err := collection.CountDocuments(ctx, filter)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate number of unique visitor"})
	}

	// 3. Check if we have *any* record of this link existing at all
	// If clicks are 0 and MongoDB logs are empty, ONLY THEN return a 404
	if totalClicksStr == "0" && len(countriesMap) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No analytics data found for this short link",
		})
	}

	return c.JSON(fiber.Map{
		"id":             id,
		"expires_at":     urlDoc.ExpiresAt,
		"status":         status,
		"total_clicks":   totalClicksStr,
		"countries":      countriesMap,
		"referrers":      referrersMap,
		"devices":        devicesMap,
		"os":             osMap,
		"browsers":       browsersMap,
		"unique_visitor": count,
	})
}
