package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ShortenUrls map[string]string
	ServeAddr   string
	HostName    string
}

func getConfig(args *Args) (*Config, error) {
	shortenUrls, err := loadShortenUrls(args.FilePath)
	if err != nil {
		return nil, err
	}

	return &Config{
		ShortenUrls: shortenUrls,
		ServeAddr:   ":8080",
		HostName:    "http://localhost",
	}, nil
}

func loadShortenUrls(filename string) (map[string]string, error) {
	type payload struct {
		Paths map[string]string `json:"paths"`
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Config file %q: %w", filename, err)
	}
	defer file.Close()

	var data payload
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON in %q: %w", filename, err)
	}

	if data.Paths == nil {
		data.Paths = make(map[string]string)
	}

	return data.Paths, nil
}
