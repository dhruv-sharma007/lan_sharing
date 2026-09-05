package util

import (
	"encoding/json"
	"math/rand/v2"
	"os"
	"time"
)

// Config represents the application configuration.
type Config struct {
	NodeID uint64 `json:"node_id"`
}

// LoadConfig loads the configuration from a file or creates one with a new NodeID if it doesn't exist.
func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Generate a new ID if config doesn't exist
			config.NodeID = generateNodeID()
			return config, saveConfig(path, config)
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// Just in case it's 0 (invalid/empty JSON)
	if config.NodeID == 0 {
		config.NodeID = generateNodeID()
		if err := saveConfig(path, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func saveConfig(path string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func generateNodeID() uint64 {
	// (time in ms << 14) | random(10000)
	return (uint64(time.Now().UnixMilli()) << 14) | uint64(rand.IntN(10000))
}
