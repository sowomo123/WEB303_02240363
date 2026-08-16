package main

import (
	"context"
	"log"
	"net"

	pb "ecommerce-microservices/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type productServer struct {
	pb.UnimplementedProductServiceServer
}

var products = map[string]*pb.Product{
	"P001": {
		ProductId: "P001",
		Name:      "Laptop",
		Price:     75000,
	},
	"P002": {
		ProductId: "P002",
		Name:      "Mechanical Keyboard",
		Price:     4500,
	},
	"P003": {
		ProductId: "P003",
		Name:      "Wireless Mouse",
		Price:     1800,
	},
	"P004": {
		ProductId: "P004",
		Name:      "Monitor",
		Price:     25000,
	},
}

func (s *productServer) GetProduct(
	ctx context.Context,
	req *pb.GetProductRequest,
) (*pb.Product, error) {

	product, exists := products[req.GetProductId()]

	if !exists {
		return nil, status.Errorf(
			codes.NotFound,
			"product %s not found",
			req.GetProductId(),
		)
	}

	return product, nil
}

func main() {

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterProductServiceServer(
		grpcServer,
		&productServer{},
	)

	// IMPORTANT: Enable gRPC reflection
	reflection.Register(grpcServer)

	log.Println("Product Service running on port 50051")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}