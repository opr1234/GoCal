package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/opr1234/calculator/internal/calculator"
	pb "github.com/opr1234/calculator/proto"
	"google.golang.org/grpc"
)

func main() {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	calcService := calculator.NewService()
	pb.RegisterCalculatorServer(srv, calcService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Println("Starting gRPC server on :50051")
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	srv.GracefulStop()
	log.Println("Server stopped")
}

func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	log.Printf("gRPC method: %s, request: %+v", info.FullMethod, req)
	return handler(ctx, req)
}
