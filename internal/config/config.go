package config

import (
	"encoding/json"
	"os"

	"github.com/ohaiibuzzle/go-lsendd/internal/utils"
)

type Config struct {
	WorkingDirectory string `json:"workdir"`
	AnnouncementName string `json:"announcement_name"`
	Fingerprint      string `json:"fingerprint"`
}

func NewConfig() *Config {
	config := &Config{}
	config.setDefaults()
	return config
}

func (c *Config) setDefaults() {
	c.WorkingDirectory = "."
	c.AnnouncementName = "go-lsendd"
	c.Fingerprint = utils.GenerateRandomString(64)
}

func (c *Config) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(c)
}

func (c *Config) SaveToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(c)
}
