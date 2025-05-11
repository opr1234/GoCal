package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/opr1234/calculator/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.CalculatorClient
}

func NewClient(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewCalculatorClient(conn),
	}, nil
}

func (c *Client) Evaluate(
	ctx context.Context,
	expr string,
	userID int32,
) (*pb.ExpressionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return c.client.Evaluate(ctx, &pb.ExpressionRequest{
		Expression: expr,
		UserId:     userID,
	})
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &pb.Empty{})
	return err
}
