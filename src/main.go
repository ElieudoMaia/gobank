package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	storage, err := NewPostgresStorage()
	if err != nil {
		log.Fatalf("Error creating storage: %v", err)
	}

	if err := storage.InitDB(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	log.Print("Database initialized")

	server := NewAPIServer(":8080", storage)
	server.Run()
}
