package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Quotes []string `json:"quotes"`
}

// GetConfig - get config from config file
func GetConfig() (*Config, error) {
	cfg := &Config{}
	f, err := os.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, err
}
