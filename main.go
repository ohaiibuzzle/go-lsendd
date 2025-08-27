package main

import (
	"flag"
	"log"
	"os"
	"time"

	api_server "github.com/ohaiibuzzle/go-lsendd/internal/api_server"
	broadcast "github.com/ohaiibuzzle/go-lsendd/internal/broadcast"
	"github.com/ohaiibuzzle/go-lsendd/internal/config"
)

func main() {
	var config_file string
	flag.StringVar(&config_file, "config", "/tmp/lsendd-config.json", "Path to configuration file")
	flag.Parse()

	var cfg *config.Config

	// Check if config file exists
	if _, err := os.Stat(config_file); os.IsNotExist(err) {
		cfg = config.NewConfig()
	} else {
		cfg = config.NewConfig()
		if err := cfg.LoadFromFile(config_file); err != nil {
			cfg = new(config.Config)
		}
	}

	// Save config file once
	if err := cfg.SaveToFile(config_file); err != nil {
		log.Printf("Error saving config file: %v", err)
	}

	go api_server.StartAPIServer("0.0.0.0", 53317)
	for {
		broadcast.SetUniqueFingerprint(cfg.Fingerprint)

		go func() {
			_, err := broadcast.SendAnnouncement(cfg.AnnouncementName)
			if err != nil {
				log.Printf("Error sending announcement: %v", err)
			}
		}()

		time.Sleep(10 * time.Second) // Wait for 10 seconds before sending the next announcement
	}
}
