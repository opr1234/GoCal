package grpc

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/calculator/internal/calculator"
	pb "github.com/yourusername/calculator/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedCalculatorServer
	evaluator *calculator.Evaluator
}

func NewServer() *Server {
	return &Server{
		evaluator: calculator.NewEvaluator(),
	}
}

func (s *Server) Evaluate(
	ctx context.Context,
	req *pb.ExpressionRequest,
) (*pb.ExpressionResponse, error) {

	if req.Expression == "" {
		return nil, status.Error(codes.InvalidArgument, "empty expression")
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	result, err := s.evaluator.Evaluate(ctx, req.Expression)
	if err != nil {
		log.Printf("Evaluation failed: %v", err)
		return handleEvaluationError(err)
	}

	return &pb.ExpressionResponse{
		Result: result,
	}, nil
}

func (s *Server) Ping(ctx context.Context, _ *pb.Empty) (*pb.Pong, error) {
	return &pb.Pong{Status: "OK"}, nil
}

func handleEvaluationError(err error) (*pb.ExpressionResponse, error) {
	switch e := err.(type) {
	case *calculator.EvaluationError:
		return nil, status.Errorf(
			codes.InvalidArgument,
			"evaluation error: %s",
			e.Message,
		)
	case *calculator.TimeoutError:
		return nil, status.Error(codes.DeadlineExceeded, "calculation timeout")
	default:
		return nil, status.Error(codes.Internal, "internal server error")
	}
}
