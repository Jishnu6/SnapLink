package utils

import (
	"net"
	"encoding/json"
	"net/http"
	"log"
	"github.com/oschwald/geoip2-golang"
	"fmt"	
	"time"
	"url-shortner/internal/models"

)

var geoDB *geoip2.Reader

// InitGeoDB loads the .mmdb file into memory on startup
func InitGeoDB() {
	var err error
	geoDB, err = geoip2.Open("internal/database/GeoLite2-Country.mmdb")
	if err != nil {
		log.Printf("⚠️ Could not load GeoIP database: %v", err)
	}
}

// CloseGeoDB cleans up the file reader when the server stops
func CloseGeoDB() {
	if geoDB != nil {
		geoDB.Close()
	}
}

// GetCountryByIP looks up the country code from a raw IP address string
func GetCountryByIP(ipStr string) string {
	if geoDB == nil {
		return "Unknown"
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "Unknown"
	}

	record, err := geoDB.Country(ip)
	if err != nil || record.Country.IsoCode == "" {
		return "Unknown"
	}

	return record.Country.IsoCode // Returns strings like "IN", "US", "DE"
}


func FetchIPDetails(ipAddress string) (*models.IPAPIResponse, error) {
	// Handle localhost gracefully to avoid wasting API calls during local development
	if ipAddress == "127.0.0.1" || ipAddress == "::1" {
		return nil, fmt.Errorf("cannot geolocate localhost IP")
	}

	// Construct the target URL
	url := fmt.Sprintf("http://ip-api.com/json/%s", ipAddress)

	// Initialize the HTTP client with a timeout so it doesn't block your background worker forever
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	// Execute the HTTP GET request
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("network error while fetching IP details: %w", err)
	}
	defer resp.Body.Close()

	// Decode the incoming JSON directly into the struct
	var result models.IPAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// ip-api returns HTTP 200 OK even if the lookup fails (e.g., private IP space), 
	// but the JSON body will have "status": "fail". We must check for this.
	if result.Status == "fail" {
		return nil, fmt.Errorf("ip-api lookup failed: %s", result.Message)
	}

	return &result, nil
}