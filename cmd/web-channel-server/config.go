package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	WSSPort  int    `json:"wssPort"`
	WSPort   int    `json:"wsPort"`
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{
		WSPort:  8080, // default values
		WSSPort: 8443,
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
