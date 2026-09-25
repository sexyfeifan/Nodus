package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// PingResult latency probe result
type PingResult struct {
	Latency   int64  // Latency in milliseconds, -1 means unreachable
	Reachable bool   // Whether the host is reachable
	Error     string // Error message if any
}

// GeoLocation IP geolocation information
type GeoLocation struct {
	Country     string  `json:"country"`     // Country name
	CountryCode string  `json:"countryCode"` // Country code (e.g., CN, US)
	Region      string  `json:"region"`      // Province/State/Region
	City        string  `json:"city"`        // City name
	ISP         string  `json:"isp"`         // Internet Service Provider
	Latitude    float64 `json:"lat"`         // Latitude coordinate
	Longitude   float64 `json:"lon"`         // Longitude coordinate
}

// PingHost tests host latency via TCP connection
func PingHost(addr string, timeout time.Duration) PingResult {
	if addr == "" {
		return PingResult{
			Latency:   -1,
			Reachable: false,
			Error:     "empty address",
		}
	}

	// If address doesn't contain port, default to 7000
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "7000")
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return PingResult{
			Latency:   -1,
			Reachable: false,
			Error:     err.Error(),
		}
	}

	conn.Close()
	return PingResult{
		Latency:   elapsed,
		Reachable: true,
		Error:     "",
	}
}

// GetGeoLocation retrieves IP geolocation information (using free public API)
func GetGeoLocation(addr string) (*GeoLocation, error) {
	// Resolve IP address (resolve domain name if needed)
	ip, err := resolveAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("resolve address failed: %w", err)
	}

	// Return directly for private IP addresses
	if isPrivateIP(ip) {
		return &GeoLocation{
			Country:     "Private Network",
			CountryCode: "LAN",
			Region:      "Local",
			City:        "Local",
			ISP:         "Private",
		}, nil
	}

	// Query using ip-api.com (free, no key required, 45 requests/min)
	return getGeoLocationFromAPI(ip)
}

// resolveAddr resolves address and extracts IP
func resolveAddr(addr string) (string, error) {
	host := addr
	// Strip port if present (handles bare IPv6 literals without ports too)
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")

	// Return directly if already an IP address
	if net.ParseIP(host) != nil {
		return host, nil
	}

	// DNS resolution
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP found for %s", host)
	}

	// Prefer IPv4 address
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	// Return first IP if no IPv4 found
	return ips[0].String(), nil
}

// isPrivateIP checks if an IP is private, loopback, link-local or CGNAT
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	if parsedIP.IsLoopback() || parsedIP.IsPrivate() || parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() || parsedIP.IsUnspecified() {
		return true
	}

	// CGNAT 100.64.0.0/10
	if _, cgnat, _ := net.ParseCIDR("100.64.0.0/10"); cgnat.Contains(parsedIP) {
		return true
	}

	return false
}

// getGeoLocationFromAPI retrieves geolocation information from ipwho.is (HTTPS, no key)
func getGeoLocationFromAPI(ip string) (*GeoLocation, error) {
	url := fmt.Sprintf("https://ipwho.is/%s", ip)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Success    bool   `json:"success"`
		Country    string `json:"country"`
		CountryCode string `json:"country_code"`
		Region     string `json:"region"`
		City       string `json:"city"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		Connection struct {
			ISP string `json:"isp"`
		} `json:"connection"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("parse API response failed: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API query failed for IP: %s", ip)
	}

	return &GeoLocation{
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		Region:      apiResp.Region,
		City:        apiResp.City,
		ISP:         apiResp.Connection.ISP,
		Latitude:    apiResp.Latitude,
		Longitude:   apiResp.Longitude,
	}, nil
}
