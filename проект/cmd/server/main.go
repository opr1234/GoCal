package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/yourusername/calculator/internal/auth"
	"github.com/yourusername/calculator/internal/storage"
	grpcClient "github.com/yourusername/calculator/internal/transport/grpc/client"
	httpTransport "github.com/yourusername/calculator/internal/transport/http"
)

func main() {
	store, err := storage.New("calc.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	if err := store.Migrate(); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	grpcConn, err := grpcClient.NewConnection(context.Background(), "localhost:50051")
	if err != nil {
		log.Fatalf("gRPC connection failed: %v", err)
	}
	defer grpcConn.Close()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}

	handler := httpTransport.NewHandler(
		store,
		grpcClient.NewCalculatorClient(grpcConn),
		secret,
	)

	router := httpTransport.NewRouter(handler, auth.Middleware(secret))

	log.Println("Starting HTTP server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
