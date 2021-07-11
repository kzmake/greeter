package handler

import (
	"context"
	"fmt"

	pb "github.com/kzmake/greeter/api/greeter/v1"
)

type greeter struct {
	pb.UnimplementedGreeterServer
}

var _ pb.GreeterServer = new(greeter)

func NewGreeter() pb.GreeterServer { return &greeter{} }

func (h *greeter) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{Msg: fmt.Sprintf("Hello, %s", req.Name)}, nil
}
