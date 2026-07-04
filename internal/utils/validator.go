package utils

import (
	"net/http"
	"net/url"
	"time"
)

// IsValidURL checks if a string is a valid absolute URL
func IsValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func IsLiveLink(url string) bool {
	// 1. Setup a client with a strict timeout (don't wait forever)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 2. Send a HEAD request
	resp, err := client.Head(url)
	if err != nil {
		return false // Server is down or URL is unreachable
	}
	defer resp.Body.Close()

	// 3. Check if the status code is in the 2xx (Success) or 3xx (Redirect) range
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}


