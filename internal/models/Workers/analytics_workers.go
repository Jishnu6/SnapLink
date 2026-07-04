package workers

import (
	"context"
	"log"
	"strings"
	"time"
	"url-shortner/internal/database"
	"url-shortner/internal/models"

	"github.com/ua-parser/uap-go/uaparser"
	"go.mongodb.org/mongo-driver/mongo"
)

var parser = uaparser.NewFromSaved()

func StartAnalyticsWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			syncMetricsToDB()
		}
	}()
}

func syncMetricsToDB() {
	ctx := context.TODO()
	redis := database.RedisClient
	collection := database.GetCollection("click_logs")

	// 1. Fetch all items currently in the Redis queue
	// RPopLUses atomic operations to safely move data out
	var logsToInsert []mongo.WriteModel

	for {
		data, err := redis.RPop(ctx, "analytics:queue").Result()
		if err != nil {
			break // Queue is empty, exit loop
		}

		parts := strings.Split(data, "::")
		if len(parts) < 6 {
			continue
		}

		urlID := parts[0]

		country := parts[1]
		city := parts[2]
		referrer := parts[3]
		fingerprint := parts[4]
		userAgent := parts[5]

		client := parser.Parse(userAgent)

		osName := client.Os.Family             // e.g., "Windows", "iOS"
		browserName := client.UserAgent.Family // e.g., "Chrome", "Firefox"

		// Compute standardized device form factor
		// deviceType := "Desktop"
		// if client.Device.Family == "Spider" || client.Device.Family == "Bot" {
		// 	deviceType = "Bot"
		// } else if osName == "iOS" || strings.Contains(strings.ToLower(osName), "android") {
		// 	deviceType = "Mobile"
		// }

		hyperLogLogKey := "uv:" + urlID
		isNewUnique, err := redis.PFAdd(ctx, hyperLogLogKey, fingerprint).Result()

		isUnique := false
		if err == nil && isNewUnique == 1 {
			isUnique = true
			log.Printf("✨ New Unique Visitor recognized for link: %s", urlID)
		}

		logEntry := models.ClickLog{
			URLID:      urlID,
			ClickedAt:  time.Now(),
			Country:    country,
			City:       city,
			Referrer:   referrer,
			IsUnique:   isUnique, // Safely assigned boolean
			DeviceType: client.Os.Family,
			OS:         osName,
			Browser:    browserName,
		}

		// Prepare a Bulk Insert Model
		insertModel := mongo.NewInsertOneModel().SetDocument(logEntry)
		logsToInsert = append(logsToInsert, insertModel)
	}

	// 2. If we found records, execute a Bulk Write to MongoDB
	if len(logsToInsert) > 0 {
		_, err := collection.BulkWrite(ctx, logsToInsert)
		if err != nil {
			log.Printf("Error running bulk write for analytics: %v", err)
		} else {
			log.Printf("Successfully flushed %d click records to MongoDB", len(logsToInsert))
		}
	}
}
