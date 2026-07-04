package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func ConnectMongo() {
	// 1. Get the URI from your .env file
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI not found in environment variables")
	}

	// 2. Set a timeout for the connection attempt
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 3. Connect to the cloud cluster
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	// 4. Ping the database to verify the connection is live
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Could not ping MongoDB Atlas: %v", err)
	}

	log.Println("Connected to MongoDB Atlas successfully!")
	MongoClient = client
}

// GetCollection is a helper to quickly access your URL collection
func GetCollection(collectionName string) *mongo.Collection {
	dbName := os.Getenv("DB_NAME")
	return MongoClient.Database(dbName).Collection(collectionName)
}

// Add this to your existing mongodb.go
func CreateIndexes() {
	collection := MongoClient.Database("url_shortener").Collection("urls")

	// Create a unique index on the "id" field (which is our _id)
	// Note: In MongoDB, _id is unique by default, but if you use a custom
	// field for the short code, you'd create the index like this:
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "original_url", Value: 1}},
		Options: options.Index().SetUnique(false), // You might want to index original_url for faster lookups
	}

	_, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Printf("Could not create index: %v", err)
	}
}

func CreateTTLIndex() {
	collection := GetCollection("urls")

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "expires_at", Value: 1}},
		// SetExpireAfterSeconds(0) means "delete when current_time > expires_at"
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	_, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Printf("Could not create TTL index: %v", err)
	}
}

type URL struct {
	ID          string    `json:"id" bson:"_id,omitempty"`          // The short code (e.g., "abc12")
	OriginalURL string    `json:"original_url" bson:"original_url"` // The long link
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	ExpiresAt   time.Time `json:"expires_at" bson:"expires_at"` // Optional: for cleanup logic
	Clicked     uint64    `json:"clicked" bson:"clicked"`       // Analytics
}
