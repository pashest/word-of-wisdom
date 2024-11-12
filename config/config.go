package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Quotes    []string  `json:"quotes"`
	TcpServer TcpServer `json:"tcp_server"`
	Equihash  Equihash  `json:"equihash"`
}

type TcpServer struct {
	ServerHost                   string `json:"server_host" envconfig:"SERVER_HOST"`
	ServerPort                   int    `json:"server_port" envconfig:"SERVER_PORT"`
	ParallelConnectionsThreshold int64  `json:"parallel_connections_threshold"`
}

type Equihash struct {
	Difficulties []Difficulty `json:"difficulties"`
}

type Difficulty struct {
	N int `json:"n"`
	K int `json:"k"`
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

	err = envconfig.Process("", cfg)
	return cfg, err
}
