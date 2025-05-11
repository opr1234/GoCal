package main

import (
	"log"

	"github.com/yourusername/calculator/internal/storage"
)

func main() {
	store, err := storage.New("calc.db")
	if err != nil {
		log.Fatalf("Storage initialization failed: %v", err)
	}

	if err := store.Migrate(); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	log.Println("✅ Database migrations completed successfully")

	if err := store.VerifySchema(); err != nil {
		log.Fatalf("Schema verification failed: %v", err)
	}
	log.Println("✅ Database schema is valid")
}
