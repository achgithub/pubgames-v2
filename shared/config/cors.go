package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CORSConfig represents the CORS configuration
type CORSConfig struct {
	Environment string    `json:"environment"`
	PubID       string    `json:"pub_id"`
	PubName     string    `json:"pub_name"`
	CORS        CORSRules `json:"cors"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   string    `json:"updated_by"`
}

// CORSRules contains the actual CORS rules
type CORSRules struct {
	Mode            string   `json:"mode"` // "pattern" or "explicit"
	Patterns        []string `json:"patterns"`
	ExplicitOrigins []string `json:"explicit_origins"`
}

// LoadCORSConfig loads the CORS configuration from the shared config file
// Falls back to safe defaults if file is missing or invalid
func LoadCORSConfig() (*CORSConfig, error) {
	// Try to find config file in shared directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not determine home directory: %v", err)
		return getDefaultConfig(), nil
	}

	configPath := filepath.Join(homeDir, "pubgames-v2", "shared", "config", "cors-config.json")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Warning: CORS config file not found at %s, using defaults", configPath)
		return getDefaultConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: Could not read CORS config: %v, using defaults", err)
		return getDefaultConfig(), nil
	}

	// Parse JSON
	var config CORSConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Warning: Could not parse CORS config: %v, using defaults", err)
		return getDefaultConfig(), nil
	}

	log.Printf("✅ Loaded CORS config: mode=%s, environment=%s", config.CORS.Mode, config.Environment)
	return &config, nil
}

// getDefaultConfig returns safe default configuration for development
func getDefaultConfig() *CORSConfig {
	config := &CORSConfig{
		Environment: "development",
		PubID:       "dev-default",
		PubName:     "Default Development",
		UpdatedAt:   time.Now(),
		UpdatedBy:   "system-default",
	}
	config.CORS.Mode = "pattern"
	config.CORS.Patterns = []string{"http://localhost:*"}
	config.CORS.ExplicitOrigins = []string{}

	log.Println("⚠️  Using default CORS config (localhost only)")
	return config
}

// IsOriginAllowed checks if an origin is allowed based on configuration
func (c *CORSConfig) IsOriginAllowed(origin string) bool {
	if c.CORS.Mode == "explicit" {
		// Explicit mode - must match exactly
		for _, allowed := range c.CORS.ExplicitOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	}

	// Pattern mode - match against patterns
	for _, pattern := range c.CORS.Patterns {
		if matchPattern(origin, pattern) {
			return true
		}
	}
	return false
}

// matchPattern checks if origin matches a pattern with * wildcard
// Supports patterns like:
//   - "http://localhost:*" (any port on localhost)
//   - "http://192.168.1.*:*" (any host/port in subnet)
//   - "https://example.com" (exact match)
func matchPattern(origin, pattern string) bool {
	// Exact wildcard
	if pattern == "*" {
		return true
	}

	// No wildcards - exact match
	if !strings.Contains(pattern, "*") {
		return origin == pattern
	}

	// Split into parts for wildcard matching
	// Simple implementation: split by "*" and check if all non-wildcard parts match in order
	parts := strings.Split(pattern, "*")

	// Origin must start with first part
	if !strings.HasPrefix(origin, parts[0]) {
		return false
	}

	// Check remaining parts appear in order
	currentPos := len(parts[0])
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if part == "" {
			continue // consecutive wildcards or trailing wildcard
		}

		// Find part in remaining origin string
		idx := strings.Index(origin[currentPos:], part)
		if idx == -1 {
			return false
		}
		currentPos += idx + len(part)
	}

	// If pattern ends with *, we're done
	// If not, origin must end with the last part
	lastPart := parts[len(parts)-1]
	if lastPart != "" && !strings.HasSuffix(origin, lastPart) {
		return false
	}

	return true
}

// GetAllowedOrigins returns a list of all allowed origins for logging/debugging
// For pattern mode, returns the patterns themselves
// For explicit mode, returns the explicit list
func (c *CORSConfig) GetAllowedOrigins() []string {
	if c.CORS.Mode == "explicit" {
		return c.CORS.ExplicitOrigins
	}
	return c.CORS.Patterns
}

// SaveCORSConfig saves the configuration back to the file
// This can be used by an admin UI to update configuration
func SaveCORSConfig(config *CORSConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, "pubgames-v2", "shared", "config", "cors-config.json")

	// Update timestamp
	config.UpdatedAt = time.Now()

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	log.Printf("✅ Saved CORS config to %s", configPath)
	return nil
}
