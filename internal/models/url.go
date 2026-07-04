package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShortenRequest is what the user sends to your API
type ShortenRequest struct {
	URL            string `json:"url"`
	CustomAlias    string `json:"custom_alias,omitempty"`
	ExpiryDuration int64  `json:"expiry_duration,omitempty"` // e.g., 24h
}

// ShortenResponse is what your API sends back to the user
type ShortenResponse struct {
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	ID          string    `json:"id"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type URL struct {
	ID          string             `json:"id" bson:"_id,omitempty"` // The short code (e.g., "abc12")
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	OriginalURL string             `json:"original_url" bson:"original_url"` // The long link
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	ExpiresAt   time.Time          `json:"expires_at" bson:"expires_at"` // Optional: for cleanup logic
	Clicked     uint64             `json:"clicked" bson:"clicked"`       // Analytics
}
