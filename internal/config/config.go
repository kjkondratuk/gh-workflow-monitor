package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	GitHubToken string
	GitHubOwner string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return nil, fmt.Errorf("GITHUB_OWNER environment variable is required")
	}

	return &Config{
		GitHubToken: token,
		GitHubOwner: owner,
	}, nil
}
