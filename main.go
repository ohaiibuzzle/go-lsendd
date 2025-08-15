package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	api_server "github.com/ohaiibuzzle/go-lsendd/api_server"
	broadcast "github.com/ohaiibuzzle/go-lsendd/broadcast"
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

func main() {
	var workdir string
	flag.StringVar(&workdir, "wd", ".", "Working directory")
	flag.Parse()

	// Set the working directory
	if err := os.Chdir(workdir); err != nil {
		log.Fatalf("Failed to change working directory: %v", err)
	}

	go api_server.StartAPIServer("0.0.0.0", 53317)

	for {
		// Generate a random 64-character string
		fingerprint := generateRandomString(64)
		broadcast.SetUniqueFingerprint(fingerprint)

		// Call the sendAnnouncement function
		success, err := broadcast.SendAnnouncement("go-lsendd")
		if err != nil {
			fmt.Printf("Error sending announcement: %v\n", err)
		}
		if success {
			fmt.Println("Announcement sent successfully!")
		}

		time.Sleep(10 * time.Second) // Wait for 10 seconds before sending the next announcement
	}
}
