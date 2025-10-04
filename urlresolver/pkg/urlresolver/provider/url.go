package provider

import (
	"encoding/json"
	"fmt"
	"os"
)

type UrlProvider interface {
	GetUrlsMap(filePath string) (map[string]string, error)
}

func NewUrlProvider() UrlProvider {
	return &urlProvider{}
}

type urlProvider struct {
}

type payload struct {
	Paths map[string]string `json:"paths"`
}

func (p *urlProvider) GetUrlsMap(filePath string) (map[string]string, error) {
	data, err := p.loadFromFile(filePath)
	if err != nil {
		return nil, err
	}

	return data.Paths, nil
}

func (p *urlProvider) loadFromFile(filename string) (payload, error) {
	file, err := os.Open(filename)
	if err != nil {
		return payload{}, fmt.Errorf("failed to open config file %q: %w", filename, err)
	}
	defer file.Close()

	var data payload
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&data); err != nil {
		return payload{}, fmt.Errorf("failed to parse JSON in %q: %w", filename, err)
	}

	if data.Paths == nil {
		data.Paths = make(map[string]string)
	}

	return data, nil
}
