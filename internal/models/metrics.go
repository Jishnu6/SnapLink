package models

import "time"

// To store raw details of an individual click event
type ClickLog struct {
	URLID      string    `bson:"url_id"`
	ClickedAt  time.Time `bson:"clicked_at"`
	Country    string    `bson:"country"`
	City       string    `bson:"city"`
	Referrer   string    `bson:"referrer"`
	IsUnique   bool      `bson:"is_unique" json:"is_unique"`
	DeviceType string    `bson:"device_type"` // e.g., "Mobile", "Desktop", "Tablet"
	OS         string    `bson:"os"`          // e.g., "iOS", "Android", "Windows", "Mac OS X"
	Browser    string    `bson:"browser"`
}

type IPAPIResponse struct {
	Status      string  `json:"status"`
	Message     string  `json:"message,omitempty"` // Included if status is "fail"
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"` // The IP address
}

// To display aggregated statistics back to the user
type URLStats struct {
	TotalClicks int            `json:"total_clicks"`
	Countries   map[string]int `json:"countries"`
	Referrers   map[string]int `json:"referrers"`
}
