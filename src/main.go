package main

import "log"

func main() {
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
