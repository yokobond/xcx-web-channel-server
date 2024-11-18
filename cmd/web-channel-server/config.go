package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	CertFile       string   `json:"certFile"`
	KeyFile        string   `json:"keyFile"`
	WSSPort        int      `json:"wssPort"`
	WSPort         int      `json:"wsPort"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

func loadConfig(filename string) (*Config, error) {
	config := &Config{
		WSPort:         8080,
		WSSPort:        8443,
		AllowedOrigins: []string{}, // Allow any origin by default
	}

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file not found; return default config
			return config, nil
		}
		return nil, err // Return error for other issues
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
