import requests
import random
import time

# Configurations
TARGET_URL = "http://localhost:3000/ui-app"  # Change to your local short URL link
TOTAL_CLICKS = 35                             # Total number of fake clicks to generate

# A curated pool of real public IPs mapped to their respective countries
IP_POOL = [
    "103.150.118.25",
   "198.51.100.12",
   "8.8.4.4",
   "203.0.113.84",
   "194.39.220.67",
    "122.106.111.45",
    "200.147.67.142"
]

# Popular referral sources
REFERRER_POOL = [
    "https://linkedin.com",
    "https://t.co", # Twitter
    "https://instagram.com",
    "https://news.ycombinator.com", # Hacker News
    "Direct",
    "Direct" # Double entry to give Direct traffic a higher probability weight
]

print(f"🚀 Starting analytics traffic simulation ({TOTAL_CLICKS} clicks)...")

for i in range(1, TOTAL_CLICKS + 1):
    fake_ip = random.choice(IP_POOL)
    fake_ref = random.choice(REFERRER_POOL)
    
    headers = {
        "X-Forwarded-For": fake_ip
    }
    
    # If it's not direct traffic, inject the HTTP Referer header
    if fake_ref != "Direct":
        headers["Referer"] = fake_ref

    try:
        # We set allow_redirects=False because we don't want python to waste time
        # downloading the final destination web page (like GitHub/Google). 
        # We only care about hitting our Go server.
        response = requests.get(TARGET_URL, headers=headers, allow_redirects=False)
        
        print(f"[Click {i:02d}] IP: {fake_ip:<15} | Source: {fake_ref:<30} | Response: {response.status_code}")
    except Exception as e:
        print(f"❌ Click {i} failed: {e}")
    
    # Tiny optional delay to prevent completely choking your system CPU
    time.sleep(0.05)

print("\n✅ Simulation Complete! Go check your stats endpoint now.")